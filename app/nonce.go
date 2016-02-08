package app
import (
	"crypto/rand"
	"github.com/square/go-jose"
)

type FixedSizeB64Nonce struct {
	length int
}

func (s *FixedSizeB64Nonce)Nonce() (string, error) {
	b := make([]byte, s.length)
	_, err := rand.Read(b)
	if err != nil {
		return "", err
	}
	return Base64URLEncode(b), nil
}

func NewFixedSizeB64(bits int) jose.NonceSource {
	return &FixedSizeB64Nonce{(bits + 7) / 8}
}