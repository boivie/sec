package storage

type Record struct {
	Key   string
	Index RecordIndex
	Data  []byte
}

type RecordIndex int32

type RecordStorage interface {
	GetLastRecordNbr(key string) RecordIndex
	Add(record Record) error
	Get(key string, from RecordIndex, to RecordIndex) []Record
	GetOne(key string, index RecordIndex) (Record, error)
	GetAll(key string) []Record
}
