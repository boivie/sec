package storage

import (
	"github.com/syndtr/goleveldb/leveldb"
	"fmt"
	"encoding/hex"
	"github.com/golang/protobuf/proto"
	"math"
	. "github.com/boivie/sec/proto"
)

const KEY_FORMAT = "r:%64s:%08x"

type LevelDbStorage struct {
	db                   *leveldb.DB
	getLastRecordNbrChan chan *getRecordIndexReq
	addChan              chan *add
	getChan              chan *get
}

type RecordHeader struct {
	LastIndex RecordIndex
}

func (s LevelDbStorage) getLastRecordNbr(topic RecordTopic) RecordIndex {
	q := s.db.NewIterator(nil, nil)
	defer q.Release()
	q.Seek(getKey(topic, math.MaxInt32))
	if !q.Prev() {
		return -1
	}
	var topic16 string
	var idx RecordIndex
	fmt.Sscan(KEY_FORMAT, &topic16, &idx)
	return idx
}

func getKey(topic RecordTopic, index RecordIndex) []byte {
	return []byte(fmt.Sprintf(KEY_FORMAT, hex.EncodeToString(topic[:]), index))
}

func (s LevelDbStorage) add(topic RecordTopic, index RecordIndex, record Record) (err error) {
	lastIndex := s.getLastRecordNbr(topic)
	if index != (lastIndex + 1) {
		err = fmt.Errorf("Index requested %d, should be %d", index, lastIndex + 1)
	} else {
		batch := new(leveldb.Batch)
		data, err := proto.Marshal(&record)
		if err == nil {
			batch.Put(getKey(topic, index), data)
			err = s.db.Write(batch, nil)
		}
	}
	return
}

func (s LevelDbStorage) get(topic RecordTopic, from RecordIndex, to RecordIndex) ([]Record, error) {
	records := []Record{}
	for idx := from; idx <= to; idx++ {
		dbk := getKey(topic, idx)
		if data, err := s.db.Get(dbk, nil); err == nil {
			var record Record
			err := proto.Unmarshal(data, &record)
			if err != nil {
				return nil, err
			}
			records = append(records, record)
		}
	}
	return records, nil
}

func (s LevelDbStorage) monitor() {
	for {
		select {
		case c := <-s.getLastRecordNbrChan:
			c.resp <- s.getLastRecordNbr(c.topic)
		case c := <-s.addChan:
			c.reply <- s.add(c.topic, c.index, c.record)
		case c := <-s.getChan:
			r, e := s.get(c.topic, c.from, c.to)
			c.reply <- getresp{r, e}
		}
	}
}

func New() (ldbs RecordStorage, err error) {
	db, err := leveldb.OpenFile("testdb", nil)

	if err == nil {
		ldbs := LevelDbStorage{
			db,
			make(chan *getRecordIndexReq),
			make(chan *add),
			make(chan *get),
		}
		go ldbs.monitor()
		return &ldbs, err
	}
	return
}