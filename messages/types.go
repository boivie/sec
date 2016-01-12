package messages
import (
	"encoding/json"
)


type CertSigningRequest struct {
	Timestamp string `json:"timestamp"`
}

type SignedCert struct {
	PEM string `json:"pem"`
}

type RootConfig struct {
	RootCert string `json:"root_cert"`
}

func Serialize(m interface{}) []byte {
	b, _ := json.Marshal(m)
	return b
}