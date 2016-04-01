package storage
import (
	"github.com/boivie/sec/proto"
	"github.com/tv42/base58"
	"math/big"
	"github.com/syndtr/goleveldb/leveldb/errors"
)

type RecordTopic [32]byte
type RecordTopicAndKey struct {
	RecordTopic RecordTopic
	Key         []byte
}
type RecordIndex int32

type RecordStorage interface {
	GetLastRecordNbr(root RecordTopic, topic RecordTopic) RecordIndex
	Add(root RecordTopic, topic *RecordTopic, index RecordIndex, message []byte, Key []byte) error
	Store(root RecordTopic, topic RecordTopic, record *proto.Record) error
	Get(root RecordTopic, topic RecordTopic, from RecordIndex, to RecordIndex) ([]proto.Record, error)
	GetOne(root RecordTopic, topic RecordTopic, index RecordIndex) (proto.Record, error)
	GetAll(root RecordTopic, topic RecordTopic) ([]proto.Record, error)
}

func (s *RecordTopic) Base58() string {
	x := new(big.Int)
	x.SetBytes(s[:])
	return string(base58.EncodeBig(nil, x))
}

func (s *RecordTopicAndKey) Base58() string {
	x := new(big.Int)
	x.SetBytes(append(s.RecordTopic[:], s.Key[:]...))
	return string(base58.EncodeBig(nil, x))
}

func DecodeTopic(b58 string) (RecordTopic, error) {
	var topic RecordTopic
	x, err := base58.DecodeToBig([]byte(b58))
	if err != nil {
		return topic, err
	}
	bytes := x.Bytes()
	if len(bytes) == 48 {
		bytes = bytes[0:32]
	} else if len(bytes) != 32 {
		return topic, errors.New("Invalid topic length");
	}

	copy(topic[:], bytes)
	return topic, nil
}

func DecodeTopicAndKey(b58 string) (RecordTopicAndKey, error) {
	var topic RecordTopicAndKey
	x, err := base58.DecodeToBig([]byte(b58))
	if err != nil {
		return topic, err
	}
	bytes := x.Bytes()

	if len(bytes) != 48 {
		return topic, errors.New("Invalid topic length");
	}

	copy(topic.RecordTopic[:], bytes[0:32])
	topic.Key = bytes[32:48]
	return topic, nil
}