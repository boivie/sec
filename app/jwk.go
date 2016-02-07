package app
import (
	"crypto/rsa"
	"encoding/binary"
	"bytes"
)

func CreatePublicKey(pub *rsa.PublicKey, kid string) JsonWebKey {
	e := make([]byte, 8)
	binary.BigEndian.PutUint64(e, uint64(pub.E))

	return JsonWebKey{
		Kid: kid,
		Kty: "RSA",
		N:   base64URLEncode(pub.N.Bytes()),
		E:   base64URLEncode(bytes.TrimLeft(e, "\x00")),
	}
}