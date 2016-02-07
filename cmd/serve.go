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