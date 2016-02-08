package storage
import (
	"errors"
	"github.com/boivie/sec/proto"
)

type getRecordIndex struct {
	topic RecordTopic
	reply chan RecordIndex
}

type add struct {
	topic  RecordTopic
	record *proto.Record
	reply  chan error
}

type appendResp struct {
	written []proto.Record;
	err     error
}

type appendCmd struct {
	records []proto.Record
	reply   chan appendResp
}

type getresp struct {
	records []proto.Record
	err     error
}

type get struct {
	topic RecordTopic
	from  RecordIndex
	to    RecordIndex
	reply chan getresp
}

type getRecordIndexReq struct {
	topic RecordTopic
	resp  chan RecordIndex
}

func (s LevelDbStorage) GetLastRecordNbr(topic RecordTopic) (ret RecordIndex) {
	myc := make(chan RecordIndex)
	s.getLastRecordNbrChan <- &getRecordIndexReq{topic, myc}
	ret = <-myc
	return
}

func (s LevelDbStorage) Add(topic RecordTopic, record *proto.Record) (err error) {
	myc := make(chan error)
	s.addChan <- &add{topic, record, myc}
	return <-myc
}

func (s LevelDbStorage) Get(topic RecordTopic, from RecordIndex, to RecordIndex) ([]proto.Record, error) {
	myc := make(chan getresp)
	s.getChan <- &get{topic, from, to, myc}
	resp := <-myc
	return resp.records, resp.err
}

func (s LevelDbStorage) GetOne(topic RecordTopic, index RecordIndex) (proto.Record, error) {
	records, err := s.Get(topic, index, index)
	if err != nil {
		return proto.Record{}, err
	}
	if len(records) == 0 {
		return proto.Record{}, errors.New("Record not found")
	}
	return records[0], nil
}

func (s LevelDbStorage) GetAll(topic RecordTopic) ([]proto.Record, error) {
	last := s.getLastRecordNbr(topic)
	if last == -1 {
		return []proto.Record{}, nil
	}
	return s.Get(topic, 0, last + 1)
}