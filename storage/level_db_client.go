package storage
import (
	"errors"
	"github.com/boivie/sec/proto"
	"fmt"
)

type getRecordIndex struct {
	root  RecordTopic
	topic RecordTopic
	reply chan RecordIndex
}

type add struct {
	root   RecordTopic
	topic  RecordTopic
	record *proto.Record
	reply  chan error
}

type getresp struct {
	records []proto.Record
	err     error
}

type get struct {
	root  RecordTopic
	topic RecordTopic
	from  RecordIndex
	to    RecordIndex
	reply chan getresp
}

type getRecordIndexReq struct {
	root  RecordTopic
	topic RecordTopic
	resp  chan RecordIndex
}

func (s LevelDbStorage) GetLastRecordNbr(root RecordTopic, topic RecordTopic) (ret RecordIndex) {
	myc := make(chan RecordIndex)
	s.getLastRecordNbrChan <- &getRecordIndexReq{root, topic, myc}
	ret = <-myc
	return
}

func (s LevelDbStorage) Add(root RecordTopic, topic *RecordTopic, index RecordIndex, message []byte, Key []byte) (err error) {
	return errors.New("Not implemented")
}

func (s LevelDbStorage) Store(root RecordTopic, topic RecordTopic, record *proto.Record) (err error) {
	myc := make(chan error)
	fmt.Printf("Adding record %s:%d\n", topic.Base58(), record.Index)
	s.addChan <- &add{root, topic, record, myc}
	return <-myc
}

func (s LevelDbStorage) Get(root RecordTopic, topic RecordTopic, from RecordIndex, to RecordIndex) ([]proto.Record, error) {
	myc := make(chan getresp)
	s.getChan <- &get{root, topic, from, to, myc}
	resp := <-myc
	return resp.records, resp.err
}

func (s LevelDbStorage) GetOne(root RecordTopic, topic RecordTopic, index RecordIndex) (proto.Record, error) {
	records, err := s.Get(root, topic, index, index)
	if err != nil {
		return proto.Record{}, err
	}
	if len(records) == 0 {
		return proto.Record{}, errors.New("Record not found")
	}
	return records[0], nil
}

func (s LevelDbStorage) GetAll(root RecordTopic, topic RecordTopic) ([]proto.Record, error) {
	last := s.GetLastRecordNbr(root, topic)
	if last == -1 {
		return []proto.Record{}, nil
	}
	return s.Get(root, topic, 0, last + 1)
}