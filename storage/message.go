package storage


type Message struct {
	Key   string
	Index RecordIndex

	Contents string
	RASignature string
}

type MessageStorage interface {
	GetLastRecordNbr(key string) RecordIndex
	// Validates the 'index'
	Add(message Message) error
	// Adds it last to the list - ignores 'index'.
	// Note: Will return the items that were appended (with index) and
	// will, if all succeeds, let err be nil, otherwise the error
	Append(additional []Message) ([]Message, error)
	Get(key string, from RecordIndex, to RecordIndex) []Message
	GetOne(key string, index RecordIndex) (Message, error)
	GetAll(key string) []Message
}
