package cmd
import (
	"github.com/codegangsta/cli"
	"github.com/boivie/sec/storage"
	"github.com/fatih/color"
	"encoding/json"
	"bytes"
	"crypto/sha256"
	"github.com/boivie/sec/app"
	"github.com/boivie/sec/httpapi"
	"fmt"
)

var CmdDump = cli.Command{
	Name:  "dump",
	Usage: "dump database",
	Action: cmdDump,
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
	},
}

func cmdDump(c *cli.Context) {
	root, err := storage.DecodeTopic(c.String("root"))
	if err != nil {
		panic(err)
	}

	var stor storage.RecordStorage;

	if (c.String("server") == "") {
		fmt.Printf("Using file backend\n")
		stor, err = storage.New()
		if err != nil {
			panic("Failed to open storage")
		}
	} else {
		fmt.Printf("Using server backend\n")
		stor, _ = httpapi.NewRemoteStorage(c.String("server"))
	}

	topic, err := storage.DecodeTopicAndKey(c.Args().First())
	if err != nil {
		panic(err)
	}

	header := color.New(color.FgYellow)
	protected := color.New(color.FgCyan)
	payload := color.New(color.FgWhite)
	signature := color.New(color.FgGreen)
	topicHeader := color.New(color.FgHiWhite)
	errorHeader := color.New(color.FgHiRed)

	topicHeader.Printf("topic_full %s\n", topic.Base58())
	topicHeader.Printf("topic_id   %s\n", topic.RecordTopic.Base58())
	topicHeader.Printf("key        %s\n\n", app.Base64URLEncode(topic.Key))

	records, err := stor.GetAll(root, topic.RecordTopic)
	if err != nil {
		errorHeader.Printf("Failed to get records\n")
	}

	for _, record := range records {
		header.Printf("index      %d\n", record.Index)
		header.Printf("type       %s\n", record.Type)

		message, err := app.DecryptMessage(record.EncryptedMessage, record.Index, topic.Key)
		if err != nil {
			errorHeader.Printf("error: Failed to decrypt message (invalid key or topic)\n")
			return
		}
		signatureHash := sha256.Sum256(message.Signature)
		b64SigHash := app.Base64URLEncode(signatureHash[:])
		signature.Printf("hash       %s\n", b64SigHash)
		protected.Printf("protected  %s\n", message.Protected)

		var f interface{}
		d := json.NewDecoder(bytes.NewBuffer(message.Payload))
		d.UseNumber()
		if err := d.Decode(&f); err != nil {
			panic(err)
		}
		s, _ := json.MarshalIndent(f, "", "  ")
		payload.Printf("\n%s\n\n", s)
	}
}