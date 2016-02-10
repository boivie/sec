package app
import (
	"crypto/rsa"
	"encoding/binary"
	"bytes"
	"github.com/boivie/sec/storage"
	"strings"
	"encoding/pem"
	"crypto/x509"
	jose "github.com/square/go-jose"
	"os"
	"io/ioutil"
)

func CreatePublicKey(pub *rsa.PublicKey, kid string) JsonWebKey {
	e := make([]byte, 8)
	binary.BigEndian.PutUint64(e, uint64(pub.E))

	return JsonWebKey{
		Kid: kid,
		Kty: "RSA",
		N:   Base64URLEncode(pub.N.Bytes()),
		E:   Base64URLEncode(bytes.TrimLeft(e, "\x00")),
	}
}

type KeyId struct {
	Topic  storage.RecordTopic
	SubKey string
}

func ParseKeyId(k string) (ret KeyId, err error) {
	topic := k
	if strings.Contains(k, "/") {
		parts := strings.SplitN(k, "/", 2)
		topic = parts[0]
		ret.SubKey = parts[1]
	}
	ret.Topic, err = storage.DecodeTopic(topic)
	return
}

func LoadKeyFromFile(filename string) (jwk *jose.JsonWebKey, err error) {
	f, err := os.Open(filename)
	if err != nil {
		return
	}
	bytes, err := ioutil.ReadAll(f)
	if err != nil {
		return
	}
	block, _ := pem.Decode([]byte(bytes))

	privateKey, err := x509.ParsePKCS1PrivateKey(block.Bytes)
	if err != nil {
		return
	}

	jwk = &jose.JsonWebKey{Key: privateKey}
	return
}