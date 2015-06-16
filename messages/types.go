package messages
import "encoding/json"

type Message struct {
}

type CertSigningRequest struct {
	Message
	Timestamp string `json:"timestamp"`
}

type SignedCert struct {
	Message
	PEM string `json:"pem"`
}


func (m Message) Serialize() []byte {
	b, _ := json.Marshal(m)
	return b
}
