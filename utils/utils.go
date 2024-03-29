package utils

import (
	"bytes"
	"crypto"
	"crypto/cipher"
	"crypto/rsa"
	"crypto/sha256"
	"crypto/x509"
	"encoding/base32"
	"encoding/base64"
	"encoding/binary"
	"encoding/json"
	"encoding/pem"
	"errors"
	"fmt"
	"github.com/boivie/gojws"
	"github.com/boivie/sec/common"
	"github.com/op/go-logging"
	"io"
	"math/big"
	"math/rand"
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
	hash := sha256.New()
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

func LoadJwk(key *gojws.Jwk) (crypto.PublicKey, error) {
	switch key.KeyType {
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
		return nil, fmt.Errorf("Unknown JWS key type %s", key.KeyType)
	}
}

func Jsonify(c http.ResponseWriter, s interface{}) {
	var encoded []byte
	if str, ok := s.(string); ok {
		encoded = []byte(str)
	} else {
		encoded, _ = json.MarshalIndent(s, "", "  ")
	}
	c.Header().Add("Content-Type", "application/vnd.sec+json")
	c.Header().Add("Content-Length", strconv.Itoa(len(encoded)+1))
	c.Write(encoded)
	io.WriteString(c, "\n")
}

func GenerateSecret() int64 {
	return rand.Int63()
}

func GetStringId(id int64, secret int64, crypt cipher.Block) string {
	buf := new(bytes.Buffer)
	binary.Write(buf, binary.LittleEndian, id)
	binary.Write(buf, binary.LittleEndian, secret)
	bytes := buf.Bytes()

	encrypted := make([]byte, 16)
	crypt.Encrypt(encrypted, bytes)

	encoder := base32.StdEncoding
	return strings.Replace(encoder.EncodeToString(encrypted), "=", "", -1)
}

func ParseStringId(id string, crypt cipher.Block) (dbId int64, secret int64, err error) {
	for {
		if len(id)%8 == 0 {
			break
		}
		id = id + "="
	}
	data, err := base32.StdEncoding.DecodeString(id)
	if err != nil {
		return
	}
	decrypted := make([]byte, 16)
	crypt.Decrypt(decrypted, data)

	buf := bytes.NewBuffer(decrypted)
	binary.Read(buf, binary.LittleEndian, &dbId)
	binary.Read(buf, binary.LittleEndian, &secret)

	return
}

func HasRecord(records []*common.Record, typ string) bool {
	_, err := GetFirstRecord(records, typ)
	return err == nil
}

func GetFirstRecord(records []*common.Record, typ string) (record *common.Record, err error) {
	for _, r := range records {
		if r.Header.Typ == typ {
			return r, nil
		}
	}
	return nil, errors.New("Not found")
}

func ParseRecord(kp gojws.KeyProvider, jws string) (*common.Record, error) {
	header, payload, e2 := ParseJws(jws, kp)
	if e2 != nil {
		return nil, errors.New("jws_parse_failed")
	}
	hash := GetFingerprint([]byte(jws))
	return &common.Record{hash, header, payload}, nil
}

func ParseRecords(kp gojws.KeyProvider, jwss []string) (records []*common.Record, err error) {
	records = make([]*common.Record, 0)

	var r *common.Record
	for _, jws := range jwss {
		r, err = ParseRecord(kp, jws)
		if err != nil {
			return
		}
		records = append(records, r)
	}
	return
}
