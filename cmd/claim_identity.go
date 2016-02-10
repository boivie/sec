package cmd
import (
	"github.com/codegangsta/cli"
)

var CmdClaimIdentity = cli.Command{
	Name:  "claim",
	Usage: "claim identity",
	Action: cmdClaim,
	Flags: []cli.Flag{
		cli.StringFlag{
			Name: "offer",
			Usage: "Offer ID",
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

}
