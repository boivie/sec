package cmd
import (
	"github.com/codegangsta/cli"
	"github.com/boivie/sec/storage"
	"github.com/boivie/sec/app"
	"github.com/boivie/sec/httpapi"
	"fmt"
	"crypto/rsa"
)

var CmdClaimIdentity = cli.Command{
	Name:  "claim",
	Usage: "claim identity",
	Action: cmdClaim,
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
			Name: "key",
			Usage: "RSA Key",
		},
	},
}

func cmdClaim(c *cli.Context) {
	root, err := storage.DecodeTopic(c.String("root"))
	if err != nil {
		panic(err)
	}
	offer, err := storage.DecodeTopicAndKey(c.Args()[0])
	if err != nil {
		panic(err)
	}
	key, err := app.LoadKeyFromFile(c.String("key"), "")

	rs, _ := httpapi.NewRemoteStorage(c.String("server"))
	parents, err := rs.GetAll(root, offer.RecordTopic)

	if len(parents) == 0 {
		fmt.Printf("Offer %s does not exist.\n", offer.Base58())
		return
	}
	if len(parents) > 1 {
		fmt.Printf("Offer %s has already been claimed.\n", offer.Base58())
		return
	}

	offerMsg, err := app.DecryptMessage(parents[0].EncryptedMessage, parents[0].Index, offer.Key)
	if err != nil {
		panic(err)
	}

	msg := app.MessageTypeIdentityClaim{}
	msg.PublicKey = app.CreatePublicKey(&key.Key.(*rsa.PrivateKey).PublicKey, "")

	record, err := app.CreateSignAndEncrypt(&msg, key, &root, &offerMsg, offer.RecordTopic, offer.Key)

	err = rs.Add(root, &offer.RecordTopic, 1, record.EncryptedMessage, offer.Key)
	if err != nil {
		panic(err)
	}

	fmt.Printf("Claimed identity at %s\n", offer.Base58())
}
