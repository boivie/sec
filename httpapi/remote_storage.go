package httpapi
import (
	"github.com/franela/goreq"
	"github.com/boivie/sec/storage"
	"github.com/boivie/sec/proto"
	"errors"
	"time"
	"github.com/boivie/sec/app"
	"strconv"
)


type remoteStorage struct {
	server string
}

func (rs *remoteStorage) GetLastRecordNbr(root storage.RecordTopic, topic storage.RecordTopic) storage.RecordIndex {
	return 0
}

func (rs *remoteStorage) Add(root storage.RecordTopic, topic *storage.RecordTopic, index storage.RecordIndex, message []byte, key []byte) error {
	var body AddMessageRequest
	body.EncryptedMessage = app.Base64URLEncode(message)
	body.Key = app.Base64URLEncode(key)

	var uri string
	if topic == nil {
		uri = rs.server + "/roots/" + root.Base58() + "/topics/"
	} else {
		uri = rs.server + "/roots/" + root.Base58() + "/topics/" + topic.Base58() + "/" + strconv.Itoa(int(index))
	}

	req := goreq.Request{
		Method:      "POST",
		Uri:         uri,
		Accept:      "application/json",
		ContentType: "application/json",
		UserAgent:   "Sec/1.0",
		Timeout:     5 * time.Second,
		Body:        body,
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
func (rs *remoteStorage) Store(root storage.RecordTopic, topic storage.RecordTopic, record *proto.Record) error {
	return errors.New("Not implemented")
}

func (rs *remoteStorage) Get(root storage.RecordTopic, topic storage.RecordTopic, from storage.RecordIndex, to storage.RecordIndex) ([]proto.Record, error) {
	return nil, errors.New("Not implemented")
}
func (rs *remoteStorage) GetOne(root storage.RecordTopic, topic storage.RecordTopic, index storage.RecordIndex) (proto.Record, error) {
	return proto.Record{}, errors.New("Not implemented")
}
func (rs *remoteStorage) GetAll(root storage.RecordTopic, topic storage.RecordTopic) (ret []proto.Record, err error) {
	req := goreq.Request{
		Uri:         rs.server + "/roots/" + root.Base58() + "/topics/" + topic.Base58(),
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

	for i := 0; i < len(resp.Records); i++ {
		rrecord := resp.Records[i]
		message, err := app.Base64URLDecode(rrecord.Message)
		if err != nil {
			return nil, err
		}
		audit, err := app.Base64URLDecode(rrecord.Audit)
		if err != nil {
			return nil, err
		}

		ret = append(ret, proto.Record{
			Index: int32(rrecord.Index),
			Type: rrecord.Type,
			EncryptedMessage: message,
			EncryptedAudit: audit,
		})
	}

	return
}

func NewRemoteStorage(server string) (ldbs storage.RecordStorage, err error) {
	return &remoteStorage{server}, nil
}