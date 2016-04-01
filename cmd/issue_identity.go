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
	offer, err := storage.DecodeTopic(c.Args()[0])
	if err != nil {
		panic(err)
	}
	key, err := app.LoadKeyFromFile(c.String("issuer_key"), c.String("issuer_id"))

	rs := httpapi.RemoteStorage{c.String("server")}
	parents, err := rs.GetAll(offer)

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

	var claim app.MessageTypeIdentityClaim
	if json.Unmarshal(parents[1].Message.Payload, &claim) != nil {
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

	record, err := app.CreateAndSign(&msg, key, nil, &parents[1])

	err = rs.Add(offer, record)
	if err != nil {
		panic(err)
	}

	fmt.Printf("Issued identity at %s\n", offer.Base58())
}
