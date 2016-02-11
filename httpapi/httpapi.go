package httpapi

import (
	"github.com/gorilla/mux"
	"io"
	"net/http"
	"io/ioutil"
	"github.com/op/go-logging"
	"strconv"
	"github.com/boivie/sec/storage"
	"encoding/json"
	"encoding/base64"
	"github.com/boivie/sec/proto"
	"github.com/boivie/sec/auditor"
	"github.com/boivie/sec/app"
	"errors"
	"fmt"
)


var stor storage.RecordStorage

var (
	auditors = make(map[storage.RecordTopic]chan auditor.AuditorRequest)
)
var log = logging.MustGetLogger("lovebeat")

func AddAuditor(id storage.RecordTopic, requests chan auditor.AuditorRequest) {
	log.Info("Adding auditor for %s\n", id.Base58())
	auditors[id] = requests
}

type JwsHeader struct {
	Alg string `json:"alg"`
}
type Jws struct {
	Header    JwsHeader `json:"header"`
	Protected string `json:"protected"`
	Payload   string `json:"payload"`
	Signature string `json:"signature"`
}

func httpError(w http.ResponseWriter, r *http.Request, reason string, status int) {
	http.Error(w, reason, status)
	fmt.Printf("HTTP %d %s (%s)\n", status, r.RequestURI, reason)
}

func getRoot(msg *proto.Message) (topic storage.RecordTopic, err error) {
	var header app.MessageTypeCommon
	if err = json.Unmarshal(msg.Payload, &header); err != nil {
		fmt.Println("1")
		return
	}

	if header.Topic != "" {
		topic, err = storage.DecodeTopic(header.Topic)
		if err != nil {
			fmt.Println("2")
			return
		}

		var first proto.Record
		if first, err = stor.GetOne(topic, 0); err != nil {
			fmt.Println("3")
			return
		}

		if err = json.Unmarshal(first.Message.Payload, &header); err != nil {
			fmt.Println("4")
			return
		}
	}

	if header.Root == "" {
		err = errors.New("Unknown topic")
		fmt.Println("5")
		return
	}
	topic, err = storage.DecodeTopic(header.Root)
	return
}

func StoreMessagesHandler(w http.ResponseWriter, r *http.Request) {
	log.Debug("Got connection to store")
	body, _ := ioutil.ReadAll(r.Body)

	var jwsMsg Jws
	err := json.Unmarshal(body, &jwsMsg)
	if err != nil {
		httpError(w, r, "Failed to parse JSON", http.StatusBadRequest)
		return
	}

	message, err := JwsMessageToProto(&jwsMsg)
	if err != nil {
		httpError(w, r, "Failed to parse JWS message", http.StatusBadRequest)
		return
	}

	root, err := getRoot(message)
	if err != nil {
		httpError(w, r, "Unknown resource", http.StatusBadRequest)
		return
	}
	aud, ok := auditors[root]
	if !ok {
		httpError(w, r, "Service unavailable", http.StatusServiceUnavailable)
		return
	}

	reply := make(chan error, 1)
	aud <- auditor.AuditorRequest{
		Message: message,
		Reply: reply,
	}

	err = <-reply
	if err == nil {
		w.Header().Set("Content-Length", "3")
		io.WriteString(w, "{}\n")
	} else {
		http.Error(w, "Bad request", http.StatusBadRequest)
	}
}

func JwsMessageToProto(in *Jws) (msg *proto.Message, err error) {
	msg = &proto.Message{}
	msg.Header, err = json.Marshal(in.Header)
	if err != nil {
		return
	}

	msg.Protected, err = base64.URLEncoding.DecodeString(in.Protected)
	if err != nil {
		return
	}
	msg.Payload, err = base64.URLEncoding.DecodeString(in.Payload)
	if err != nil {
		return
	}
	msg.Signature, err = base64.URLEncoding.DecodeString(in.Signature)
	if err != nil {
		return
	}
	return
}

func ProtoMessageToJws(in *proto.Message) (out *Jws, err error) {
	var header JwsHeader
	err = json.Unmarshal(in.Header, &header)
	if err == nil {
		out = &Jws{
			Header: header,
			Protected: base64.URLEncoding.EncodeToString(in.Protected),
			Payload: base64.URLEncoding.EncodeToString(in.Payload),
			Signature: base64.URLEncoding.EncodeToString(in.Signature),
		}
	}
	return
}

type RecordContents struct {
	Message *Jws `json:"message"`
	Audit   *Jws `json:"audit,omitempty"`
}

type GetTopicResponse struct {
	Records []RecordContents `json:"records"`
}

func GetTopicHandler(w http.ResponseWriter, r *http.Request) {
	params := mux.Vars(r)
	topic, err := storage.DecodeTopic(params["topic"])
	if err != nil {
		http.Error(w, "Invalid topic", http.StatusNotFound)
		return
	}
	records, err := stor.GetAll(topic)
	if err != nil {
		http.Error(w, "Failed to fetch records", http.StatusInternalServerError)
		return
	}
	ret := GetTopicResponse{
		Records: make([]RecordContents, 0, len(records)),
	}

	for _, record := range records {
		msg, err := ProtoMessageToJws(record.Message)
		if err != nil {
			http.Error(w, "Failed to parse message", http.StatusInternalServerError)
			return
		}
		var audit *Jws
		if record.Audit != nil {
			audit, err = ProtoMessageToJws(record.Audit)
			if err != nil {
				http.Error(w, "Failed to parse audit", http.StatusInternalServerError)
				return
			}
		}
		ret.Records = append(ret.Records, RecordContents{
			Message: msg,
			Audit: audit,
		})
	}

	var encoded, _ = json.MarshalIndent(ret, "", "  ")
	w.Header().Add("Content-Type", "application/json")
	w.Header().Add("Content-Length", strconv.Itoa(len(encoded) + 1))
	w.Write(encoded)
	io.WriteString(w, "\n")
}


func Register(rtr *mux.Router, stor_ storage.RecordStorage) {
	stor = stor_
	rtr.HandleFunc("/store", StoreMessagesHandler).Methods("POST")
	rtr.HandleFunc("/topics/{topic:[A-Za-z0-9]+}", GetTopicHandler).Methods("GET")
}
