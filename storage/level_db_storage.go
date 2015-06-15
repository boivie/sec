package storage

import (
	"github.com/syndtr/goleveldb/leveldb"
	"fmt"
	"encoding/json"
)

type getRecordIndexReq struct {
	key  string
	resp chan RecordIndex
}

type LevelDbStorage struct {
	db                   *leveldb.DB
	getLastRecordNbrChan chan *getRecordIndexReq
	addChan              chan *add
	appendChan           chan *append
}

type RecordHeader struct {
	LastIndex RecordIndex
}

func (s LevelDbStorage) getRecordHeader(key string) (header RecordHeader, err error) {
	if data, err := s.db.Get(getHeaderKey(key), nil); err == nil {
		err = json.Unmarshal(data, &header)
	}
	return
}

func (s LevelDbStorage) getLastRecordNbr(key string) RecordIndex {
	if header, err := s.getRecordHeader(key); err == nil {
		return header.LastIndex
	}
	return 0
}

func getKey(record Record) []byte {
	return []byte(fmt.Sprintf("r-%s-%d", record.Key, record.Index))
}

func getHeaderKey(key string) []byte {
	return []byte("r-" + key)
}

func queue(batch *leveldb.Batch, record Record) {
	batch.Put(getKey(record), record.Data)

	header := RecordHeader{LastIndex: record.Index}
	data, _ := json.Marshal(header)
	batch.Put(getHeaderKey(record.Key), data)
}

func (s LevelDbStorage) add(record Record) (err error) {
	lastIndex := s.getLastRecordNbr(record.Key)
	if record.Index != (lastIndex + 1) {
		err = fmt.Errorf("Index requested %d, should be %d", record.Index, lastIndex+1)
	} else {
		batch := new(leveldb.Batch)
		queue(batch, record)
		return s.db.Write(batch, nil)
	}

	return
}

func (s LevelDbStorage) append(records []Record) error {
	batch := new(leveldb.Batch)
	for _, record := range records {
		record.Index = s.getLastRecordNbr(record.Key) + 1
		queue(batch, record)
	}
	return s.db.Write(batch, nil)
}

func (s LevelDbStorage) monitor() {
	for {
		select {
		case c := <-s.getLastRecordNbrChan:
			c.resp <- s.getLastRecordNbr(c.key)
		case c := <-s.addChan:
			c.reply <- s.add(c.record)
		case c := <-s.appendChan:
			c.reply <- s.append(c.records)
		}
	}
}

func New() (ldbs Storage, err error) {
	db, err := leveldb.OpenFile("path/to/db", nil)

	if err == nil {
		ldbs := LevelDbStorage{
			db,
			make(chan *getRecordIndexReq),
			make(chan *add),
			make(chan *append),
		}
		go ldbs.monitor()
		return &ldbs, err
	}
	return
}