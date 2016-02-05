package main

import (
	"github.com/op/go-logging"
	"fmt"
	"net/http"
	"github.com/gorilla/mux"
	"os"
	"github.com/boivie/sec/config"
	"github.com/boivie/sec/httpapi"
	"flag"
	"github.com/boivie/sec/storage"
	"github.com/boivie/sec/proto"
	"encoding/base64"
)


var (
	signalchan = make(chan os.Signal, 1)
	log = logging.MustGetLogger("lovebeat")

	debug = flag.Bool("debug", false, "Enable debug logs")
	cfgFile = flag.String("config", "/etc/lovebeat.cfg", "Configuration file")
)


func httpServer(cfg *config.ConfigBind, stor storage.RecordStorage) {
	rtr := mux.NewRouter()
	httpapi.Register(rtr, stor)
	http.Handle("/", rtr)
	log.Info("HTTP listening on %s\n", cfg.Listen)
	http.ListenAndServe(cfg.Listen, nil)
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

func main() {
	flag.Parse()

	if *debug {
		logging.SetLevel(logging.DEBUG, "lovebeat")
	} else {
		logging.SetLevel(logging.INFO, "lovebeat")
	}

	var cfg = config.ReadConfig(*cfgFile)

	stor, err := storage.New()
	if err != nil {
		panic("Failed to open storage")
	}
	go httpServer(&cfg.Http, stor)

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
	b, _ := base64.StdEncoding.DecodeString("aK/6H6851b73Qkm/swPuCSIcaja1Ysg6b2CNAsklDLY=")
	copy(topic[:], b)
	stor.Add(topic, 0, r)

	signalHandler()
}
