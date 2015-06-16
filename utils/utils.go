package utils
import (
	"crypto/sha256"
	"encoding/hex"
	"time"
)

func RecordChecksum(data []byte) string {
	hasher := sha256.New()
	hasher.Write(data)
	return hex.EncodeToString(hasher.Sum(nil))
}

func Now() int64 {
	return time.Now().UnixNano() / 1e6
}

func NowStr() string {
	return time.Now().Format(time.RFC3339)
}
