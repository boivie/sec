package httpapi
import (
	"github.com/franela/goreq"
	"github.com/boivie/sec/storage"
	"github.com/boivie/sec/proto"
	"errors"
	"time"
	"encoding/base64"
	"fmt"
)


type RemoteStorage struct {
	Server string
}

func (rs *RemoteStorage) GetLastRecordNbr(topic storage.RecordTopic) storage.RecordIndex {
	return 0
}

func (rs *RemoteStorage) Add(topic storage.RecordTopic, record *proto.Record) error {
	var jws Jws
	jws.Header.Alg = "RS256"
	jws.Protected = base64.URLEncoding.EncodeToString(record.Message.Protected)
	jws.Payload = base64.URLEncoding.EncodeToString(record.Message.Payload)
	jws.Signature = base64.URLEncoding.EncodeToString(record.Message.Signature)

	req := goreq.Request{
		Method:      "POST",
		Uri:         rs.Server + "/store",
		Accept:      "application/json",
		ContentType: "application/json",
		UserAgent:   "Sec/1.0",
		Timeout:     5 * time.Second,
		Body:        jws,
	}

	ret, err := req.Do()
	if err != nil {
		return err
	}
	if ret.StatusCode != 200 {
		return errors.New("Invalid status code")
	}
	return nil
}
func (rs *RemoteStorage) Get(topic storage.RecordTopic, from storage.RecordIndex, to storage.RecordIndex) ([]proto.Record, error) {
	return nil, errors.New("Not implemented")
}
func (rs *RemoteStorage) GetOne(topic storage.RecordTopic, index storage.RecordIndex) (proto.Record, error) {
	return proto.Record{}, errors.New("Not implemented")
}
func (rs *RemoteStorage) GetAll(topic storage.RecordTopic) (ret []proto.Record, err error) {
	req := goreq.Request{
		Uri:         rs.Server + "/topics/" + topic.Base58(),
		Accept:      "application/json",
		UserAgent:   "Sec/1.0",
		Timeout:     5 * time.Second,
	}

	httpResponse, err := req.Do()
	if err != nil {
		return
	}
	if httpResponse.StatusCode != 200 {
		err = errors.New("Invalid status code")
		return
	}
	var resp GetTopicResponse
	err = httpResponse.Body.FromJsonTo(&resp)
	if err != nil {
		return
	}
	fmt.Printf("JSON got %d\n", len(resp.Records))

	for i := 0; i < len(resp.Records); i++ {
		var record proto.Record
		record.Message, err = JwsMessageToProto(resp.Records[i].Message)
		if err != nil {
			fmt.Printf("Error 1\n")
			return
		}
		if resp.Records[i].Audit != nil {
			record.Audit, err = JwsMessageToProto(resp.Records[i].Audit)
			if err != nil {
				fmt.Printf("Error 2\n")
				return
			}
		}
		ret = append(ret, record)
	}

	return
}
