package storage

import (
	"github.com/syndtr/goleveldb/leveldb"
	"fmt"
	"github.com/golang/protobuf/proto"
	"math"
	. "github.com/boivie/sec/proto"
	"encoding/binary"
)

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
	var foundTopic RecordTopic
	copy(foundTopic[:], q.Key())
	idx := binary.BigEndian.Uint32(q.Key()[32:36])
	if foundTopic != topic {
		return -1
	}
	return RecordIndex(idx)
}

func getKey(topic RecordTopic, index RecordIndex) []byte {
	ret := make([]byte, 32 + 4)
	copy(ret, topic[:])
	binary.BigEndian.PutUint32(ret[32:], uint32(index))
	return ret
}

func (s LevelDbStorage) add(topic RecordTopic, record *Record) (err error) {
	lastIndex := int32(s.getLastRecordNbr(topic))
	if record.Index != (lastIndex + 1) {
		err = fmt.Errorf("Index requested %d, should be %d", record.Index, lastIndex + 1)
	} else {
		batch := new(leveldb.Batch)
		data, err := proto.Marshal(record)
		if err == nil {
			batch.Put(getKey(topic, RecordIndex(record.Index)), data)
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
			c.reply <- s.add(c.topic, c.record)
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