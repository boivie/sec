package storage
import "errors"

type getRecordIndex struct {
	key   string
	reply chan RecordIndex
}

type add struct {
	record Record
	reply  chan error
}

type appendResp struct {
	written []Record;
	err     error
}

type appendCmd struct {
	records []Record
	reply   chan appendResp
}

type get struct {
	key   string
	from  RecordIndex
	to    RecordIndex
	reply chan []Record
}

func (s LevelDbStorage) GetLastRecordNbr(key string) (ret RecordIndex) {
	myc := make(chan RecordIndex)
	s.getLastRecordNbrChan <- &getRecordIndexReq{key, myc}
	ret = <-myc
	return
}

func (s LevelDbStorage) Add(record Record) (err error) {
	myc := make(chan error)
	s.addChan <- &add{record, myc}
	return <-myc
}

func (s LevelDbStorage) Get(key string, from RecordIndex, to RecordIndex) []Record {
	myc := make(chan []Record)
	s.getChan <- &get{key, from, to, myc}
	return <-myc
}

func (s LevelDbStorage) GetOne(key string, index RecordIndex) (Record, error) {
	records := s.Get(key, index, index)
	if len(records) == 0 {
		return Record{}, errors.New("Record not found")
	}
	return records[0], nil
}

func (s LevelDbStorage) GetAll(key string) []Record {
	last := s.getLastRecordNbr(key)
	if last == 0 {
		return []Record{}
	}
	return s.Get(key, 1, last)
}