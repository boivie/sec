package httpapi

import (
	"github.com/gorilla/mux"
	"io"
	"net/http"
	"io/ioutil"
	"github.com/op/go-logging"
	"sync"
	"fmt"
	"strconv"
	"github.com/boivie/sec/storage"
	"encoding/json"
	"encoding/base64"
	"github.com/boivie/sec/proto"
)

type Request struct {
	body      []byte
	requestId int64
}

var stor storage.RecordStorage

var (
	lock sync.Mutex
	auditor chan Request
	responses map[int64]chan bool = make(map[int64]chan bool)
	lastId int64
)
var log = logging.MustGetLogger("lovebeat")

type jws_header struct {
	Alg string `json:"alg"`
}
type jws struct {
	Header    jws_header `json:"header"`
	Protected string `json:"protected"`
	Payload   string `json:"payload"`
	Signature string `json:"signature"`
}

func createResponse() (int64, chan bool) {
	response := make(chan bool, 1)

	lock.Lock()
	defer lock.Unlock()

	lastId = lastId + 1
	requestId := lastId
	responses[requestId] = response
	return requestId, response
}

func StoreMessagesHandler(w http.ResponseWriter, r *http.Request) {
	log.Debug("Got connection to store")
	body, _ := ioutil.ReadAll(r.Body)

	requestId, response := createResponse()

	auditor <- Request{
		body: body,
		requestId: requestId,
	}
	log.Debug("Waiting for response from auditor")
	okey := <-response
	if okey {
		w.Header().Set("Content-Length", "3")
		io.WriteString(w, "{}\n")
	} else {
		w.Header().Set("Content-Length", "6")
		io.WriteString(w, "{BAD}\n")
	}
}

func AuditorListenHandler(w http.ResponseWriter, r *http.Request) {
	log.Debug("Got connection from auditor")
	flusher, ok := w.(http.Flusher)
	if !ok {
		panic("expected http.ResponseWriter to be an http.Flusher")
	}
	auditor = make(chan Request, 1)
	for {
		req := <-auditor
		fmt.Fprintf(w, "%d\n%s\n", req.requestId, req.body)
		flusher.Flush()
	}
}

func getResponse(requestId int64) chan bool {
	lock.Lock()
	defer lock.Unlock()
	response, ok := responses[requestId]
	if ok {
		delete(responses, requestId)
	}
	return response
}

func getTopicIndex(msg RecordContents) (topic storage.RecordTopic, index storage.RecordIndex, err error) {
	if data, err := base64.StdEncoding.DecodeString(msg.Audit.Payload); err == nil {
		var auditMsg struct {
			Topic string `json:"topic"`
			Index storage.RecordIndex `json:"index"`
		}

		if err = json.Unmarshal(data, &auditMsg); err == nil {
			var tbytes []byte
			tbytes, err = base64.StdEncoding.DecodeString(auditMsg.Topic)
			copy(topic[:], tbytes)
			index = auditMsg.Index
		}
	}
	return
}

func JwsMessageToProto(in jws) (msg proto.Message, err error) {
	msg.Header, err = json.Marshal(in.Header)
	if err != nil {
		return
	}

	msg.Protected, err = base64.StdEncoding.DecodeString(in.Protected)
	if err != nil {
		return
	}
	msg.Payload, err = base64.StdEncoding.DecodeString(in.Payload)
	if err != nil {
		return
	}
	msg.Signature, err = base64.StdEncoding.DecodeString(in.Signature)
	if err != nil {
		return
	}
	return
}

func ProtoMessageToJws(in *proto.Message) (out jws, err error) {
	err = json.Unmarshal(in.Header, &out.Header)
	if err == nil {
		out.Protected = base64.StdEncoding.EncodeToString(in.Protected)
		out.Payload = base64.StdEncoding.EncodeToString(in.Payload)
		out.Signature = base64.StdEncoding.EncodeToString(in.Signature)
	}
	return
}

type RecordContents struct {
	Message jws `json:"message"`
	Audit   jws `json:"audit"`
}

func AuditorStoreHandler(w http.ResponseWriter, r *http.Request) {
	body, err := ioutil.ReadAll(r.Body)
	if err != nil {
		return
	}
	var msg RecordContents
	err = json.Unmarshal(body, msg)
	if err != nil {
		return
	}

	topic, idx, err := getTopicIndex(msg)

	if err != nil {
		return
	}

	message, err := JwsMessageToProto(msg.Message)
	if err != nil {
		return
	}
	audit, err := JwsMessageToProto(msg.Audit)
	if err != nil {
		return
	}

	record := proto.Record{Index: int32(idx), Message: &message, Audit: &audit}
	stor.Add(topic, idx, record)

	// Optional
	if requestIdStr, ok := r.URL.Query()["request_id"]; ok {
		if requestId, err := strconv.ParseInt(requestIdStr[0], 10, 64); err == nil {
			if response := getResponse(requestId); response != nil {
				response <- true
			}
		}
	}

	w.Header().Set("Content-Length", "3")
	io.WriteString(w, "{}\n")
}

type GetTopicResponse struct {
	Records []RecordContents `json:"records"`
}

func GetTopicHandler(w http.ResponseWriter, r *http.Request) {
	params := mux.Vars(r)
	tbytes, err := base64.StdEncoding.DecodeString(params["topic"])
	var topic storage.RecordTopic
	copy(topic[:], tbytes)

	records, err := stor.GetAll(topic)
	if err != nil {
		return
	}
	ret := GetTopicResponse{
		Records: make([]RecordContents, 0, len(records)),
	}

	for _, record := range records {
		msg, err := ProtoMessageToJws(record.Message)
		if err != nil {
			return
		}
		audit, err := ProtoMessageToJws(record.Audit)
		if err != nil {
			return
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
	rtr.HandleFunc("/api/v1/store", StoreMessagesHandler).Methods("POST")
	rtr.HandleFunc("/api/v1/auditor/listen", AuditorListenHandler).Methods("GET")
	rtr.HandleFunc("/api/v1/auditor/store/{request_id:[a-z0-9.-]+}", AuditorStoreHandler).Methods("POST")
	rtr.HandleFunc("/api/v1/topics/{topic:[A-Za-z0-9]+}", GetTopicHandler).Methods("POST")
}
