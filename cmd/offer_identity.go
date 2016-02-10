package cmd
import (
	"github.com/codegangsta/cli"
	"github.com/boivie/sec/storage"
	"github.com/boivie/sec/app"
	"github.com/boivie/sec/httpapi"
	"encoding/base64"
	"github.com/franela/goreq"
	"time"
	"fmt"
	"github.com/boivie/sec/proto"
	"errors"
)

var CmdOfferIdentity = cli.Command{
	Name:  "offer",
	Usage: "offer identity",
	Action: cmdOffer,
	Flags: []cli.Flag{
		cli.StringFlag{
			Name: "root",
			Usage: "Root",
		},
		cli.StringFlag{
			Name: "server",
			Value: "http://localhost:8080",
			Usage: "Address and port to listen on",
		},
		cli.StringFlag{
			Name: "issuer_id",
			Usage: "Issuer id",
		},
		cli.StringFlag{
			Name: "issuer_key",
			Usage: "Issuer key",
		},
		cli.StringFlag{
			Name: "ref",
			Usage: "Message reference",
		},
	},
}

type RemoteStorage struct {
	Server string
}

func (rs *RemoteStorage) GetLastRecordNbr(topic storage.RecordTopic) storage.RecordIndex {
	return 0
}

func (rs *RemoteStorage) Add(topic storage.RecordTopic, record *proto.Record) error {
	var jws httpapi.Jws
	jws.Header.Alg = "RS256"
	jws.Protected = base64.URLEncoding.EncodeToString(record.Message.Protected)
	jws.Payload = base64.URLEncoding.EncodeToString(record.Message.Payload)
	jws.Signature = base64.URLEncoding.EncodeToString(record.Message.Payload)

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
	var resp httpapi.GetTopicResponse
	err = httpResponse.Body.FromJsonTo(&resp)
	if err != nil {
		return
	}

	for i := 0; i < len(resp.Records); i++ {
		var record proto.Record
		record.Message, err = httpapi.JwsMessageToProto(resp.Records[i].Message)
		if err != nil {
			return
		}
		record.Audit, err = httpapi.JwsMessageToProto(resp.Records[i].Audit)
		if err != nil {
			return
		}
		ret = append(ret, record)
	}

	return
}


func cmdOffer(c *cli.Context) {
	root, err := storage.DecodeTopic(c.String("root"))
	if err != nil {
		panic(err)
	}
	msg := app.MessageTypeIdentityOffer{}
	msg.Title = c.Args()[0]
	msg.MessageTypeCommon.Ref = c.String("ref")

	key, err := app.LoadKeyFromFile(c.String("issuer_key"))
	key.KeyID = c.String("issuer_id")

	record, err := app.CreateAndSign(&msg, key, &root, nil)

	rs := RemoteStorage{c.String("server")}
	topic := app.GetTopic(record.Message)

	err = rs.Add(topic, record)
	if err != nil {
		panic(err)
	}

	fmt.Printf("Offering identity at %s\n", topic.Base58())
}
