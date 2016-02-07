package main

import (
	"github.com/op/go-logging"
	"github.com/fatih/color"
	"fmt"
	"net/http"
	"github.com/gorilla/mux"
	"os"
	"github.com/boivie/sec/httpapi"
	"github.com/boivie/sec/storage"
	"github.com/boivie/sec/proto"
	"github.com/codegangsta/cli"
	"encoding/hex"
	"crypto/rsa"
	"crypto/rand"
	jose "github.com/square/go-jose"
	"github.com/boivie/sec/app"
	"time"
	"encoding/json"
	"crypto/sha256"
	"bytes"
)


var (
	signalchan = make(chan os.Signal, 1)
	log = logging.MustGetLogger("lovebeat")
)

func signalHandler() {
	for {
		select {
		case sig := <-signalchan:
			fmt.Printf("!! Caught signal %d... shutting down\n", sig)
			return
		}
	}
}

func getFlags() []cli.Flag {
	return []cli.Flag{
		cli.StringFlag{
			Name: "db",
			Value: "testdb",
			Usage: "Path to database",
		},
	}
}

func dumpDb(c *cli.Context) {
	stor, err := storage.New()
	if err != nil {
		panic("Failed to open storage")
	}
	topic, err := storage.DecodeTopic(c.Args().First())
	if err != nil {
		panic(err)
	}

	header := color.New(color.FgYellow)
	protected := color.New(color.FgCyan)
	payload := color.New(color.FgWhite)

	records, err := stor.GetAll(topic)
	for _, record := range records {
		header.Printf("record %s:%d (%s)\n", topic.Base58(), record.Index, record.Type)
		protected.Printf("protected %s\n", record.Message.Protected)

		var f interface{}
		d := json.NewDecoder(bytes.NewBuffer(record.Message.Payload))
		d.UseNumber()
		if err := d.Decode(&f); err != nil {
			panic(err)
		}
		s, _ := json.MarshalIndent(f, "", "  ")
		payload.Printf("\n%s\n\n", s)
	}
}

func bootstrap(c *cli.Context) {
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

func benchmark(c *cli.Context) {
	stor, err := storage.New()
	if err != nil {
		panic("Failed to open storage")
	}

	r := proto.Record{
		0,
		"type",
		&proto.Message{
			[]byte{1, 2, 3},
			[]byte{2, 3, 4},
			[]byte{3, 4, 5},
			[]byte{4, 5, 6},
		},
		&proto.Message{
			[]byte{1, 2, 3},
			[]byte{2, 3, 4},
			[]byte{3, 4, 5},
			[]byte{4, 5, 6},
		},
	}
	var topic storage.RecordTopic
	b, _ := hex.DecodeString("A9575FE17F1208F8AA9A795C8CC75F6E6B6DCDEBC6F385633B73BB45C1202DCF")
	copy(topic[:], b)
	for i := 0; i < 10000; i++ {
		if err := stor.Add(topic, stor.GetLastRecordNbr(topic) + 1, r); err != nil {
			panic(err)
		}
	}
}

func serveHttp(c *cli.Context) {
	stor, err := storage.New()
	if err != nil {
		panic("Failed to open storage")
	}
	a := func() {
		rtr := mux.NewRouter()
		httpapi.Register(rtr, stor)
		http.Handle("/", rtr)
		listen := c.String("listen")
		log.Info("HTTP listening on %s\n", listen)
		http.ListenAndServe(listen, nil)
	}
	go a()
	signalHandler()
}

func getCommands() []cli.Command {
	return []cli.Command{
		{
			Name: "init",
			Usage: "Bootstraps and creates root",
			Action: bootstrap,
		},
		{
			Name:  "dump",
			Usage: "dump database",
			Action: dumpDb,
		},
		{
			Name:  "serve",
			Usage: "serve http",
			Flags: []cli.Flag{
				cli.StringFlag{
					Name: "listen",
					Value: ":8080",
					Usage: "Address and port to listen on",
				},
			},
			Action: serveHttp,
		},
		{
			Name: "benchmark",
			Usage: "Run benchmark",
			Action: benchmark,
		},
		{
			Name:      "auditor",
			Usage:     "options for task templates",
			Action: func(c *cli.Context) {
				println("completed task: ", c.Args().First())
			},
		},
	}
}

func main() {
	app := cli.NewApp()
	app.Name = "sec"
	app.Usage = "Secure identification"
	app.Flags = getFlags()
	app.Commands = getCommands()
	app.Run(os.Args)
}
