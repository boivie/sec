package main

import (
	"crypto/aes"
	"flag"
	"fmt"
	"github.com/boivie/sec/common"
	"github.com/boivie/sec/dao"
	"github.com/boivie/sec/httpapi"
	_ "github.com/go-sql-driver/mysql"
	"github.com/gorilla/mux"
	"github.com/jinzhu/gorm"
	"github.com/op/go-logging"
	"net/http"
	"os"
	"os/signal"
	"runtime"
	"strings"
	"syscall"
)

var log = logging.MustGetLogger("sec")

const (
	VERSION = "0.1.0"
)

var (
	debug       = flag.Bool("debug", false, "print statistics sent to graphite")
	showVersion = flag.Bool("version", false, "print version string")
	workDir     = flag.String("workdir", "work", "working directory")
	cfgFile     = flag.String("config", "/etc/sec.cfg", "configuration file")
)

var (
	signalchan = make(chan os.Signal, 1)
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

func httpServer(port int16, state *common.State) {
	rtr := mux.NewRouter()
	httpapi.Register(rtr, state)
	http.Handle("/", rtr)
	log.Info("HTTP server running on port %d\n", port)
	http.ListenAndServe(fmt.Sprintf(":%d", port), nil)
}

func getHostname() string {
	var hostname, err = os.Hostname()
	if err != nil {
		return fmt.Sprintf("unknown_%d", os.Getpid())
	}
	return strings.Split(hostname, ".")[0]
}

func initDb(state *common.State) {
	state.DB.AutoMigrate(&dao.TemplateDao{})
	state.DB.AutoMigrate(&dao.InvitationDao{})
	state.DB.AutoMigrate(&dao.IdentityDao{})
	state.DB.AutoMigrate(&dao.CertDao{})
}

func main() {
	flag.Parse()

	var format = logging.MustStringFormatter("%{level} %{message}")
	logging.SetFormatter(format)
	if *debug {
		logging.SetLevel(logging.DEBUG, "sec")
	} else {
		logging.SetLevel(logging.INFO, "sec")
	}
	log.Debug("Debug logs enabled")

	if *showVersion {
		fmt.Printf("sec v%s (built w/%s)\n", VERSION, runtime.Version())
		return
	}

	db, _ := gorm.Open("mysql", "root:@/sec?charset=utf8&parseTime=True")
	block, _ := aes.NewCipher([]byte("example key 1234"))
	state := common.State{
		DB:       db,
		BaseUrl:  "http://localhost:8989",
		IdCrypto: block,
	}
	initDb(&state)

	var hostname = getHostname()
	log.Info("Sec v%s started as host %s, PID %d", VERSION, hostname, os.Getpid())
	log.Info("Base URL: '%s'", state.BaseUrl)

	signal.Notify(signalchan, syscall.SIGTERM)

	go httpServer(8989, &state)

	log.Info("Ready to handle incoming connections")

	signalHandler()
}
