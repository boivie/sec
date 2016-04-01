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
	"github.com/boivie/sec/auditor"
	"github.com/boivie/sec/app"
	"fmt"
)


var stor storage.RecordStorage
var auditors = make(map[storage.RecordTopic]chan auditor.AuditorRequest)
var log = logging.MustGetLogger("lovebeat")

func AddAuditor(id storage.RecordTopic, requests chan auditor.AuditorRequest) {
	log.Info("Adding auditor for %s\n", id.Base58())
	auditors[id] = requests
}

type JwsHeader struct {
	Alg string `json:"alg"`
}
type AddMessageRequest struct {
	EncryptedMessage string `json:"message"`
	Key              string `json:"key"`
}

func httpError(w http.ResponseWriter, r *http.Request, reason string, status int) {
	http.Error(w, reason, status)
	fmt.Printf("HTTP %d %s (%s)\n", status, r.RequestURI, reason)
}

func NewTopicHandler(w http.ResponseWriter, r *http.Request) {
	fmt.Printf("in: %s\n", r.RequestURI)

	params := mux.Vars(r)
	root, err := storage.DecodeTopic(params["root"])
	if err != nil {
		httpError(w, r, "Invalid root", http.StatusNotFound)
		return
	}

	body, _ := ioutil.ReadAll(r.Body)

	var req AddMessageRequest
	err = json.Unmarshal(body, &req)
	if err != nil {
		httpError(w, r, "Failed to parse JSON", http.StatusBadRequest)
		return
	}

	aud, ok := auditors[root]
	if !ok {
		httpError(w, r, "Service unavailable", http.StatusServiceUnavailable)
		return
	}

	message, err := app.Base64URLDecode(req.EncryptedMessage)
	if err != nil {
		httpError(w, r, "Failed to parse encrypted message", http.StatusBadRequest)
		return
	}

	key, err := app.Base64URLDecode(req.Key)
	if err != nil {
		httpError(w, r, "Failed to parse encrypted key", http.StatusBadRequest)
		return
	}

	reply := make(chan error, 1)
	aud <- auditor.AuditorRequest{
		Root: root,
		Topic: nil,
		Index: 0,
		EncryptedMessage: message,
		Key: key,
		Reply: reply,
	}
	err = <-reply

	if err == nil {
		w.Header().Set("Content-Length", "3")
		io.WriteString(w, "{}\n")
	} else {
		fmt.Printf("Got error from auditor: %v\n", err)
		httpError(w, r, "Bad request", http.StatusBadRequest)
	}
}

func AddMessageToTopicHandler(w http.ResponseWriter, r *http.Request) {
	fmt.Printf("%s\n", r.RequestURI)

	params := mux.Vars(r)
	root, err := storage.DecodeTopic(params["root"])
	if err != nil {
		httpError(w, r, "Invalid root", http.StatusNotFound)
		return
	}

	topic, err := storage.DecodeTopic(params["topic"])
	if err != nil {
		httpError(w, r, "Invalid topic", http.StatusNotFound)
		return
	}

	index, err := strconv.ParseInt(params["index"], 10, 31)
	if err != nil {
		httpError(w, r, "Invalid index", http.StatusNotFound)
		return
	}

	body, _ := ioutil.ReadAll(r.Body)

	var req AddMessageRequest
	err = json.Unmarshal(body, &req)
	if err != nil {
		httpError(w, r, "Failed to parse JSON", http.StatusBadRequest)
		return
	}

	aud, ok := auditors[root]
	if !ok {
		httpError(w, r, "Service unavailable", http.StatusServiceUnavailable)
		return
	}

	message, err := app.Base64URLDecode(req.EncryptedMessage)
	if err != nil {
		httpError(w, r, "Failed to parse encrypted message", http.StatusBadRequest)
		return
	}

	key, err := app.Base64URLDecode(req.Key)
	if err != nil {
		httpError(w, r, "Failed to parse encrypted key", http.StatusBadRequest)
		return
	}

	reply := make(chan error, 1)
	aud <- auditor.AuditorRequest{
		Root: root,
		Topic: &topic,
		Index: int32(index),
		EncryptedMessage: message,
		Key: key,
		Reply: reply,
	}
	err = <-reply

	if err == nil {
		w.Header().Set("Content-Length", "3")
		io.WriteString(w, "{}\n")
	} else {
		fmt.Printf("Got error from auditor: %v\n", err)
		httpError(w, r, "Bad request", http.StatusBadRequest)
	}
}


type RecordContents struct {
	Index   int `json:"index"`
	Type    string `json:"type"`
	Message string `json:"message"`
	Audit   string `json:"audit,omitempty"`
}

type GetTopicResponse struct {
	Records []RecordContents `json:"records"`
}

func GetTopicHandler(w http.ResponseWriter, r *http.Request) {
	fmt.Printf("%s\n", r.RequestURI)

	params := mux.Vars(r)
	root, err := storage.DecodeTopic(params["root"])
	if err != nil {
		httpError(w, r, "Invalid root", http.StatusNotFound)
		return
	}
	topic, err := storage.DecodeTopic(params["topic"])
	if err != nil {
		httpError(w, r, "Invalid topic", http.StatusNotFound)
		return
	}
	records, err := stor.GetAll(root, topic)
	if err != nil {
		httpError(w, r, "Failed to fetch records", http.StatusInternalServerError)
		return
	}
	ret := GetTopicResponse{
		Records: make([]RecordContents, 0, len(records)),
	}

	for _, record := range records {
		ret.Records = append(ret.Records, RecordContents{
			Index: int(record.Index),
			Type: record.Type,
			Message: app.Base64URLEncode(record.EncryptedMessage),
			Audit: app.Base64URLEncode(record.EncryptedAudit),
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
	rtr.HandleFunc("/roots/{root:[A-Za-z0-9]+}/topics/", NewTopicHandler).Methods("POST")
	rtr.HandleFunc("/roots/{root:[A-Za-z0-9]+}/topics/{topic:[A-Za-z0-9]+}/{index:[0-9]+}", AddMessageToTopicHandler).Methods("POST")
	rtr.HandleFunc("/roots/{root:[A-Za-z0-9]+}/topics/{topic:[A-Za-z0-9]+}", GetTopicHandler).Methods("GET")
}
