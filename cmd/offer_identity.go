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

	var jws httpapi.Jws
	jws.Header.Alg = "RS256"
	jws.Protected = base64.URLEncoding.EncodeToString(record.Message.Protected)
	jws.Payload = base64.URLEncoding.EncodeToString(record.Message.Payload)
	jws.Signature = base64.URLEncoding.EncodeToString(record.Message.Payload)

	req := goreq.Request{
		Method:      "POST",
		Uri:         c.String("server") + "/store",
		Accept:      "application/json",
		ContentType: "application/json",
		UserAgent:   "Sec/1.0",
		Timeout:     5 * time.Second,
		Body:        jws,
	}

	ret, err := req.Do()
	if err != nil {
		panic(err)
	}
	if ret.StatusCode != 200 {
		log.Error("Failed to post to server")
	}

	topic := app.GetTopic(record.Message)
	fmt.Printf("Offering identity at %s\n", topic.Base58())
}
