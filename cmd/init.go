package cmd
import (
	"github.com/boivie/sec/app"
	"github.com/boivie/sec/storage"
	"fmt"
	jose "github.com/square/go-jose"

	"crypto/rsa"
	"crypto/rand"
	"github.com/codegangsta/cli"
	"encoding/pem"
	"crypto/x509"
	"os"
)

var CmdInit = cli.Command{
	Name: "init",
	Usage: "Bootstraps and creates root",
	Action: cmdInit,
}

func cmdInit(c *cli.Context) {
	privateKey, err := rsa.GenerateKey(rand.Reader, 2048)
	if err != nil {
		panic(err)
	}

	p := pem.Block{Type: "RSA PRIVATE KEY", Bytes: x509.MarshalPKCS1PrivateKey(privateKey)}

	rootKeyFname := "root-key.pem"
	keyOut, err := os.OpenFile(rootKeyFname, os.O_WRONLY | os.O_CREATE | os.O_TRUNC, 0600)
	if err != nil {
		panic(err)
	}
	pem.Encode(keyOut, &p)
	keyOut.Close()
	fmt.Printf("Wrote root key to %s\n", rootKeyFname)

	cfg := app.MessageTypeRootConfig{}
	cfg.Keys = []app.RootKey{
		app.RootKey{
			Identifier: "@root/key0",
			PublicKey: app.CreatePublicKey(&privateKey.PublicKey, ""),
			Usage: app.KeyUsage{
				Auditor: &app.KeyUsageAuditor{},
				IssueIdentities: &app.KeyUsageIssueIdentities{},
				RequestSigning: &app.KeyUsageRequestSigning{},
			},
		},
	}

	jwkKey := &jose.JsonWebKey{
		Key: privateKey,
		KeyID: "@root/root",
	}

	stor, err := storage.New()
	if err != nil {
		panic(err)
	}

	rootRecord, err := app.CreateAndSign(&cfg, jwkKey, nil)
	if err != nil {
		panic(err)
	}

	topic := app.GetTopic(rootRecord.Message)
	stor.Add(topic, rootRecord)
	fmt.Printf("Created root %s\n", topic.Base58())
}