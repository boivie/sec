package cmd
import (
	"github.com/codegangsta/cli"
	"github.com/boivie/sec/storage"
	"github.com/boivie/sec/app"
	"github.com/boivie/sec/httpapi"
	"fmt"
	"encoding/json"
)

var CmdIssueIdentity = cli.Command{
	Name:  "issue",
	Usage: "issue identity",
	Action: cmdIssue,
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
	},
}

func cmdIssue(c *cli.Context) {
	root, err := storage.DecodeTopic(c.String("root"))
	if err != nil {
		panic(err)
	}
	offer, err := storage.DecodeTopicAndKey(c.Args()[0])
	if err != nil {
		panic(err)
	}
	key, err := app.LoadKeyFromFile(c.String("issuer_key"), c.String("issuer_id"))

	rs, _ := httpapi.NewRemoteStorage(c.String("server"))
	parents, err := rs.GetAll(root, offer.RecordTopic)

	if len(parents) == 0 {
		fmt.Printf("Offer %s does not exist.\n", offer.Base58())
		return
	}
	if len(parents) == 1 {
		fmt.Printf("Offer %s has not been claimed.\n", offer.Base58())
		return
	}
	if len(parents) > 2 {
		fmt.Printf("Offer %s can't be claimed.\n", offer.Base58())
		return
	}

	claimMsg, err := app.DecryptMessage(parents[1].EncryptedMessage, parents[1].Index, offer.Key)
	var claim app.MessageTypeIdentityClaim
	if json.Unmarshal(claimMsg.Payload, &claim) != nil {
		fmt.Printf("Couldn't parse claim\n")
		return
	}

	pubKey, err := claim.PublicKey.ToPublicKey()
	if err != nil {
		fmt.Printf("Couldn't parse claim key\n")
		return
	}

	msg := app.MessageTypeIdentityIssue{}
	msg.PublicKey = app.CreatePublicKey(pubKey, "")
	msg.Title = "Test Identity"
	msg.Path = "/test/1"

	record, err := app.CreateSignAndEncrypt(&msg, key, &root, &claimMsg, offer.RecordTopic, offer.Key)

	err = rs.Add(root, &offer.RecordTopic, 2, record.EncryptedMessage, offer.Key)
	if err != nil {
		panic(err)
	}

	fmt.Printf("Issued identity at %s\n", offer.Base58())
}
