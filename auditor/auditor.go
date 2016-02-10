package auditor
import (
	"github.com/boivie/sec/storage"
	"github.com/boivie/sec/proto"
	"github.com/boivie/sec/app"
	"encoding/json"
	jose "github.com/square/go-jose"
)

type AuditorConfig struct {
	KeyId      app.KeyId
	PrivateKey *jose.JsonWebKey
	Backend    storage.RecordStorage
}

type AuditorRequest struct {
	Reply   chan error
	Message *proto.Message
}

func addMessage(cfg *AuditorConfig, message *proto.Message, header *app.MessageTypeCommon) (err error) {
	// Validate topic
	var topic storage.RecordTopic
	if header.Topic != "" {
		topic, err = storage.DecodeTopic(header.Topic)
		if err != nil {
			return
		}
	} else {
		topic = app.GetTopic(message)
	}
	// Validate index

	audit := proto.Message{
	}

	record := proto.Record{
		Index: header.Index,
		Type: header.Resource,
		Message: message,
		Audit: &audit,
	}

	err = cfg.Backend.Add(topic, &record)
	return
}

func Create(cfg AuditorConfig) chan AuditorRequest {
	requests := make(chan AuditorRequest, 100)
	worker := func() {
		for req := range requests {
			var header app.MessageTypeCommon
			if err := json.Unmarshal(req.Message.Payload, &header); err != nil {
				req.Reply <- err
			} else {
				req.Reply <- addMessage(&cfg, req.Message, &header)
			}
		}
	}
	go worker()
	return requests
}