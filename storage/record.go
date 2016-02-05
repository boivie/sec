package storage
import "github.com/boivie/sec/proto"

type RecordTopic [32]byte
type RecordIndex int32

type RecordStorage interface {
	GetLastRecordNbr(topic RecordTopic) RecordIndex
	Add(topic RecordTopic, index RecordIndex, record proto.Record) error
	Get(topic RecordTopic, from RecordIndex, to RecordIndex) ([]proto.Record, error)
	GetOne(topic RecordTopic, index RecordIndex) (proto.Record, error)
	GetAll(topic RecordTopic) ([]proto.Record, error)
}
