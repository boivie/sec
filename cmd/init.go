package cmd
import (
	"github.com/boivie/sec/app"
	"github.com/boivie/sec/proto"
	"crypto/sha256"
	"github.com/boivie/sec/storage"
	"fmt"
	jose "github.com/square/go-jose"

	"crypto/rsa"
	"crypto/rand"
	"time"
	"encoding/json"
	"github.com/codegangsta/cli"
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

	jwkKey := &jose.JsonWebKey{
		Key: privateKey,
		KeyID: "@root/root",
	}

	cfg := app.MessageTypeRootConfig{}
	cfg.MessageTypeCommon.Resource = "root.config"
	cfg.MessageTypeCommon.At = time.Now().UnixNano() / 1000000
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

	payload := app.SerializeJSON(cfg)

	signer, err := jose.NewSigner(jose.RS256, jwkKey)
	if err != nil {
		panic(err)
	}

	signer.SetNonceSource(app.NewFixedSizeB64(256))
	//	signer.EmbedJwk(false)

	object, err := signer.Sign(payload)
	if err != nil {
		panic(err)
	}

	// We can't access the protected header without serializing - ugly workaround.
	serialized := object.FullSerialize()

	var parsed struct {
		Protected string `json:"protected"`
		Payload   string `json:"payload"`
		Signature string `json:"signature"`
	}
	err = json.Unmarshal([]byte(serialized), &parsed)
	if err != nil {
		panic(err)
	}

	stor, err := storage.New()
	if err != nil {
		panic(err)
	}

	var topic storage.RecordTopic = sha256.Sum256(app.MustBase64URLDecode(parsed.Signature))

	r := proto.Record{
		Index: 0,
		Type: "root.config",
		Message: &proto.Message{
			[]byte("{\"alg\":\"RS256\"}"),
			app.MustBase64URLDecode(parsed.Protected),
			app.MustBase64URLDecode(parsed.Payload),
			app.MustBase64URLDecode(parsed.Signature),
		},
	}
	stor.Add(topic, 0, r)

	fmt.Printf("Created root %s\n", topic.Base58())
}