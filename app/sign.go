package app
import (
	. "github.com/boivie/sec/proto"
	"github.com/boivie/sec/storage"
	"crypto/sha256"
	jose "github.com/square/go-jose"
	"encoding/json"
)

func CreateAndSign(cfg MessageType, jwkKey *jose.JsonWebKey, root *storage.RecordTopic, parent *Record) (*Record, error) {
	cfg.Initialize(root, parent)

	payload := SerializeJSON(cfg)

	signer, err := jose.NewSigner(jose.RS256, jwkKey)
	if err != nil {
		return nil, err
	}

	signer.SetNonceSource(NewFixedSizeB64(256))

	object, err := signer.Sign(payload)
	if err != nil {
		return nil, err
	}

	// We can't access the protected header without serializing - ugly workaround.
	serialized := object.FullSerialize()

	var parsed struct {
		Protected string `json:"protected"`
		Payload   string `json:"payload"`
		Signature string `json:"signature"`
	}
	err = json.Unmarshal([]byte(serialized), &parsed)
	if err != nil {
		return nil, err
	}

	message := Message{
		[]byte("{\"alg\":\"RS256\"}"),
		MustBase64URLDecode(parsed.Protected),
		MustBase64URLDecode(parsed.Payload),
		MustBase64URLDecode(parsed.Signature),
	}

	return &Record{
		Index: cfg.Header().Index,
		Type: cfg.Header().Resource,
		Message: &message,
	}, nil
}

func GetTopic(m *Message) storage.RecordTopic {
	var topic storage.RecordTopic = sha256.Sum256(m.Signature)
	return topic
}