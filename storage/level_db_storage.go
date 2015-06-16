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
	appendChan           chan *appendCmd
	getChan              chan *get
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

func getKey(key string, index RecordIndex) []byte {
	return []byte(fmt.Sprintf("r:%s:%08x", key, index))
}

func getHeaderKey(key string) []byte {
	return []byte("r:" + key)
}

func queue(batch *leveldb.Batch, record Record) {
	batch.Put(getKey(record.Key, record.Index), record.Data)

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

func (s LevelDbStorage) doAppend(records []Record) (written []Record, err error) {
	batch := new(leveldb.Batch)
	for _, record := range records {
		record.Index = s.getLastRecordNbr(record.Key) + 1
		queue(batch, record)
		written = append(written, record)
	}
	return written, s.db.Write(batch, nil)
}

func (s LevelDbStorage) get(key string, from RecordIndex, to RecordIndex) []Record {
	records := []Record{}
	for idx := from; idx <= to; idx++ {
		dbk := getKey(key, idx)
		if data, err := s.db.Get(dbk, nil); err == nil {
			record := Record{
				Key: key,
				Index: idx,
				Data: data,
			}
			records = append(records, record)
		}
	}
	return records
}

func (s LevelDbStorage) monitor() {
	for {
		select {
		case c := <-s.getLastRecordNbrChan:
			c.resp <- s.getLastRecordNbr(c.key)
		case c := <-s.addChan:
			c.reply <- s.add(c.record)
		case c := <-s.appendChan:
			records, e := s.doAppend(c.records)
			c.reply <- appendResp{records, e}
		case c := <-s.getChan:
			c.reply <- s.get(c.key, c.from, c.to)
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
			make(chan *appendCmd),
			make(chan *get),
		}
		go ldbs.monitor()
		return &ldbs, err
	}
	return
}