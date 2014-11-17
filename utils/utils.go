package utils

import (
	"bytes"
	"crypto"
	"crypto/rsa"
	"crypto/sha1"
	"crypto/sha256"
	"crypto/x509"
	"encoding/base64"
	"encoding/binary"
	"encoding/json"
	"encoding/pem"
	"errors"
	"fmt"
	"github.com/boivie/gojws"
	"github.com/op/go-logging"
	"io"
	"math/big"
	"net/http"
	"strconv"
	"strings"
)

var (
	log = logging.MustGetLogger("sec")
)

func B64encode(data []byte) string {
	return strings.Replace(
		base64.URLEncoding.EncodeToString(data),
		"=", "", -1)
}

func B64decode(str string) ([]byte, error) {
	lenMod4 := len(str) % 4
	if lenMod4 > 0 {
		str = str + strings.Repeat("=", 4-lenMod4)
	}

	return base64.URLEncoding.DecodeString(str)
}

func GetFingerprint(data []byte) (ret string) {
	hash := sha256.New()
	hash.Write(data)
	return B64encode(hash.Sum(nil))
}

func GetCertPem(der []byte) (contents string, err error) {
	var buf bytes.Buffer
	err = pem.Encode(&buf, &pem.Block{Type: "CERTIFICATE", Bytes: der})
	if err != nil {
		return
	}
	contents = string(buf.Bytes())
	return
}

func GetKeyPem(der []byte) (contents string, err error) {
	var buf bytes.Buffer
	err = pem.Encode(&buf, &pem.Block{Type: "RSA PRIVATE KEY", Bytes: der})
	if err != nil {
		return
	}
	contents = string(buf.Bytes())
	return
}

func GetCertFingerprint(cert *x509.Certificate) string {
	hash := sha1.New()
	hash.Write(cert.Raw)
	return B64encode(hash.Sum(nil))
}

func ParseJws(tokenString string, kp gojws.KeyProvider) (header gojws.Header, payload map[string]interface{}, err error) {
	var data []byte
	header, data, err = gojws.VerifyAndDecodeWithHeader(tokenString, kp)
	if err != nil {
		log.Warning("%v", err)
	}
	err = json.Unmarshal(data, &payload)
	return
}

func LoadJwk(jwk string) (crypto.PublicKey, error) {
	var key struct {
		Kty string `json:"kty"`
		N   string `json:"n"`
		E   string `json:"e"`
	}
	err := json.Unmarshal([]byte(jwk), &key)
	if err != nil {
		return nil, fmt.Errorf("Failed to unmarshal key: %v", err)
	}

	switch key.Kty {
	case "RSA":
		if key.N == "" || key.E == "" {
			return nil, errors.New("Malformed JWS RSA key")
		}

		data, err := B64decode(key.E)
		if err != nil {
			return nil, errors.New("Malformed JWS RSA key")
		}
		if len(data) < 4 {
			ndata := make([]byte, 4)
			copy(ndata[4-len(data):], data)
			data = ndata
		}

		pubKey := &rsa.PublicKey{
			N: &big.Int{},
			E: int(binary.BigEndian.Uint32(data[:])),
		}

		data, err = B64decode(key.N)
		if err != nil {
			return nil, errors.New("Malformed JWS RSA key")
		}
		pubKey.N.SetBytes(data)

		return pubKey, nil

	default:
		return nil, fmt.Errorf("Unknown JWS key type %s", key.Kty)
	}
}

func Jsonify(c http.ResponseWriter, s interface{}) {
	var encoded []byte
	if str, ok := s.(string); ok {
		encoded = []byte(str)
	} else {
		encoded, _ = json.MarshalIndent(s, "", "  ")
	}
	c.Header().Add("Content-Type", "application/json")
	c.Header().Add("Content-Length", strconv.Itoa(len(encoded)+1))
	c.Write(encoded)
	io.WriteString(c, "\n")
}
