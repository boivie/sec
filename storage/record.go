package storage
import (
	"github.com/boivie/sec/proto"
	"github.com/tv42/base58"
	"math/big"
)

type RecordTopic [32]byte
type RecordIndex int32

type RecordStorage interface {
	GetLastRecordNbr(topic RecordTopic) RecordIndex
	Add(topic RecordTopic, record *proto.Record) error
	Get(topic RecordTopic, from RecordIndex, to RecordIndex) ([]proto.Record, error)
	GetOne(topic RecordTopic, index RecordIndex) (proto.Record, error)
	GetAll(topic RecordTopic) ([]proto.Record, error)
}

func (s *RecordTopic) Base58() string {
	x := new(big.Int)
	x.SetBytes(s[:])
	return string(base58.EncodeBig(nil, x))
}

func DecodeTopic(b58 string) (RecordTopic, error) {
	var topic RecordTopic
	x, err := base58.DecodeToBig([]byte(b58))
	if err == nil {
		copy(topic[:], x.Bytes())
	}
	return topic, nil
}
