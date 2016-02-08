package app
import (
	"encoding/base64"
	"strings"
)

func Base64URLEncode(data []byte) string {
	var result = base64.URLEncoding.EncodeToString(data)
	return strings.TrimRight(result, "=")
}

// Url-safe base64 decoder that adds padding
func Base64URLDecode(data string) ([]byte, error) {
	var missing = (4 - len(data) % 4) % 4
	data += strings.Repeat("=", missing)
	return base64.URLEncoding.DecodeString(data)
}

func MustBase64URLDecode(data string) []byte {
	bytes, err := Base64URLDecode(data)
	if err != nil {
		panic(err)
	}
	return bytes
}