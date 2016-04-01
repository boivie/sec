package auditor
import (
	"github.com/boivie/sec/storage"
	"github.com/boivie/sec/app"
	"encoding/json"
	jose "github.com/square/go-jose"
	"github.com/syndtr/goleveldb/leveldb/errors"
	"crypto/sha256"
	. "github.com/boivie/sec/proto"
	"fmt"
)

type AuditorConfig struct {
	KeyId      app.KeyId
	PrivateKey *jose.JsonWebKey
	Backend    storage.RecordStorage
}

type AuditorRequest struct {
	Reply            chan error
	Root             storage.RecordTopic
	Topic            *storage.RecordTopic
	Index            int32
	EncryptedMessage []byte
	Key              []byte
}

func addMessage(cfg *AuditorConfig, req AuditorRequest) (err error) {
	message, err := app.DecryptMessage(req.EncryptedMessage, req.Index, req.Key)
	if err != nil {
		return errors.New("Failed to decrypt message");
	}
	var header app.MessageTypeCommon
	err = json.Unmarshal(message.Payload, &header)
	if err != nil {
		return
	}

	// Validate topic
	var topic storage.RecordTopic
	if req.Topic != nil {
		topic, err = storage.DecodeTopic(header.Topic)
		if err != nil {
			return
		}
		if topic.Base58() != req.Topic.Base58() {
			return fmt.Errorf("Topic in message doesn't match %s <-> %s", topic.Base58(), req.Topic.Base58())
		}
	} else {
		topic = sha256.Sum256(req.EncryptedMessage)
	}

	// TODO: Validate a lot more

	var encryptedAudit []byte

	record := Record{
		Index: header.Index,
		Type: header.Resource,
		EncryptedMessage: req.EncryptedMessage,
		EncryptedAudit: encryptedAudit,
	}

	err = cfg.Backend.Store(req.Root, topic, &record)
	return
}

func Create(cfg AuditorConfig) chan AuditorRequest {
	requests := make(chan AuditorRequest, 100)
	worker := func() {
		for req := range requests {
			req.Reply <- addMessage(&cfg, req)
		}
	}
	go worker()
	return requests
}
