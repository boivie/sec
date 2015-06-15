package storage

type Record struct {
	Key   string
	Index RecordIndex
	Data  []byte
}

type RecordIndex int32

type Storage interface {
	GetLastRecordNbr(key string) RecordIndex
	// Validates the 'index'
	Add(record Record) error
	// Adds it last to the list - ignores 'index'.
	Append(additional []Record) error
	Get(key string, from RecordIndex, to RecordIndex) []Record
	GetOne(key string, index RecordIndex) (Record, error)
	GetAll(key string) []Record
}