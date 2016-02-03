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

type RecordContents struct {
	Message jws `json:"message"`
	Audit   jws `json:"audit"`
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

func getTopicIndex(msg RecordContents) (topic string, index storage.RecordIndex, err error) {
	if data, err := base64.StdEncoding.DecodeString(msg.Audit.Payload); err == nil {
		var auditMsg struct {
			Topic string `json:"topic"`
			Index storage.RecordIndex `json:"index"`
		}

		if err = json.Unmarshal(data, auditMsg); err == nil {
			topic = auditMsg.Topic
			index = auditMsg.Index
		}
	}
	return
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

	data, err := json.Marshal(msg)
	if err != nil {
		return
	}
	stor.Add(storage.Record{Key: topic, Index: idx, Data: data})

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
	topic := params["topic"]

	records := stor.GetAll(topic)
	ret := GetTopicResponse{
		Records: make([]RecordContents, len(records)),
	}

	for idx, record := range records {
		var msg RecordContents
		json.Unmarshal(record.Data, msg)
		ret.Records[idx] = msg
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
