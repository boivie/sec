package utils

import (
	"bytes"
	"crypto/sha1"
	"crypto/sha256"
	"crypto/x509"
	"encoding/base64"
	"encoding/pem"
	"strings"
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
