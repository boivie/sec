package cmd
import (
	"github.com/codegangsta/cli"
	"github.com/boivie/sec/storage"
	"github.com/gorilla/mux"
	"github.com/boivie/sec/httpapi"
	"net/http"
	"os"
	"fmt"
	"github.com/op/go-logging"
	"github.com/boivie/sec/auditor"
	"github.com/boivie/sec/app"
)

var (
	signalchan = make(chan os.Signal, 1)
	log = logging.MustGetLogger("sec")
)

var CmdServe = cli.Command{
	Name:  "serve",
	Usage: "serve http",
	Flags: []cli.Flag{
		cli.StringFlag{
			Name: "listen",
			Value: ":8080",
			Usage: "Address and port to listen on",
		},
		cli.StringFlag{
			Name: "auditor_key",
			Usage: "Auditor private key filename (PEM)",
		},
		cli.StringFlag{
			Name: "auditor_id",
			Usage: "Auditor identity",
		},
	},
	Action: cmdServe,
}

func signalHandler() {
	for {
		select {
		case sig := <-signalchan:
			fmt.Printf("!! Caught signal %d... shutting down\n", sig)
			return
		}
	}
}

func cmdServe(c *cli.Context) {
	stor, err := storage.New()
	if err != nil {
		panic("Failed to open storage")
	}

	if c.IsSet("auditor_id") {
		keyId, err := app.ParseKeyId(c.String("auditor_id"))
		if err != nil {
			panic(err)
		}
		priv, err := app.LoadKeyFromFile(c.String("auditor_key"), c.String("auditor_id"))
		if err != nil {
			panic(err)
		}

		auditor := auditor.Create(auditor.AuditorConfig{
			KeyId: keyId,
			PrivateKey: priv,
			Backend: stor,
		})
		httpapi.AddAuditor(keyId.Topic, auditor)
	}

	rtr := mux.NewRouter()
	httpapi.Register(rtr, stor)
	http.Handle("/", rtr)
	listen := c.String("listen")
	log.Info("HTTP listening on %s\n", listen)

	go http.ListenAndServe(listen, nil)
	signalHandler()
}