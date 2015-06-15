package storage

type getRecordIndex struct {
	key   string
	reply chan RecordIndex
}

type add struct {
	record Record
	reply  chan error
}

type append struct {
	records []Record
	reply   chan error
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
	err = <-myc
	return
}

func (s LevelDbStorage) Append(records []Record) (err error) {
	myc := make(chan error)
	s.appendChan <- &append{records, myc}
	err = <-myc
	return
}

func (s LevelDbStorage) Get(key string, from RecordIndex, to RecordIndex) []Record {
	return nil
}

func (s LevelDbStorage) GetOne(key string, index RecordIndex) (Record, error) {
	return Record{}, nil
}

func (s LevelDbStorage) GetAll(key string) []Record {
	return nil
}