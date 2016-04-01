package cmd
import (
	"github.com/boivie/sec/app"
	"github.com/boivie/sec/storage"
	"fmt"
	jose "github.com/square/go-jose"

	"crypto/rsa"
	"crypto/rand"
	"github.com/codegangsta/cli"
	"crypto/x509"
	"encoding/pem"
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

	cfg := app.MessageTypeRootConfig{}
	cfg.Keys = []app.RootKey{
		app.RootKey{
			Identifier: "key1",
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
	}

	stor, err := storage.New()
	if err != nil {
		panic(err)
	}

	rootRecord, topicKey, err := app.CreateSignAndEncryptInitial(&cfg, jwkKey, nil)
	if err != nil {
		panic(err)
	}

	topic := app.GetTopic(rootRecord, topicKey);
	stor.Store(topic.RecordTopic, topic.RecordTopic, rootRecord)

	fmt.Printf("Created root %s\n", topic.Base58())

	p := pem.Block{Type: "RSA PRIVATE KEY", Bytes: x509.MarshalPKCS1PrivateKey(privateKey)}

	rootKeyFname := fmt.Sprintf("root-%s.pem", topic.Base58())
	keyOut, err := os.OpenFile(rootKeyFname, os.O_WRONLY | os.O_CREATE | os.O_TRUNC, 0600)
	if err != nil {
		panic(err)
	}
	pem.Encode(keyOut, &p)
	keyOut.Close()
	fmt.Printf("Wrote root key to %s\n", rootKeyFname)
	fmt.Printf("\nStart a local daemon by running:\n")
	fmt.Printf("%s serve --auditor_id=%s/key1 --auditor_key=root-%s.pem\n", os.Args[0], topic.Base58(), topic.Base58())
}