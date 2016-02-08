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

func createAccount(stor storage.RecordStorage, key *jose.JsonWebKey) storage.RecordTopic {
	account := app.MessageTypeAccountCreate{}

	accountRecord, err := app.CreateAndSign(&account, key, nil)
	if err != nil {
		panic(err)
	}

	accountTopic := app.GetTopic(accountRecord.Message)
	stor.Add(accountTopic, accountRecord)
	return accountTopic
}

func createAuditorIdentity(stor storage.RecordStorage, key *jose.JsonWebKey, accountId storage.RecordTopic) storage.RecordTopic {
	var offer app.MessageTypeIdentityOffer

	offerRecord, err := app.CreateAndSign(&offer, key, nil)
	if err != nil {
		panic(err)
	}

	identityTopic := app.GetTopic(offerRecord.Message)
	stor.Add(identityTopic, offerRecord)

	// Claim
	claim := app.MessageTypeIdentityClaim{}

	claimRecord, err := app.CreateAndSign(&claim, key, offerRecord)
	if err != nil {
		panic(err)
	}

	stor.Add(identityTopic, claimRecord)

	// Issue
	issue := app.MessageTypeIdentityIssue{}

	issueRecord, err := app.CreateAndSign(&issue, key, claimRecord)
	if err != nil {
		panic(err)
	}

	stor.Add(identityTopic, issueRecord)

	return identityTopic
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
	fmt.Println("Wrote root key to %s\n", rootKeyFname)

	cfg := app.MessageTypeRootConfig{}
	cfg.Roots.AuditorRoots = []app.Root{
		app.Root{
			PublicKey: app.CreatePublicKey(&privateKey.PublicKey, "@root/auditor_issuer_1"),
		},
	}
	cfg.Roots.IdentityRoots = []app.Root{
		app.Root{
			PublicKey: app.CreatePublicKey(&privateKey.PublicKey, "@root/identity_issuer_1"),
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

	// Create account
	accountId := createAccount(stor, jwkKey)
	// And create an auditor identity
	auditorId := createAuditorIdentity(stor, jwkKey, accountId)

	fmt.Printf("Created auditor identity %s\n", auditorId.Base58())
}