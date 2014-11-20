package main

import (
	"crypto/aes"
	"crypto/rand"
	"crypto/x509"
	"crypto/x509/pkix"
	"encoding/json"
	"flag"
	"fmt"
	"github.com/boivie/gojws"
	"github.com/boivie/sec/bootstrap"
	"github.com/boivie/sec/common"
	"github.com/boivie/sec/dao"
	"github.com/boivie/sec/httpapi"
	"github.com/boivie/sec/store"
	"github.com/boivie/sec/store/dbstore"
	"github.com/boivie/sec/utils"
	_ "github.com/go-sql-driver/mysql"
	"github.com/gorilla/mux"
	"github.com/jinzhu/gorm"
	"github.com/op/go-logging"
	"math/big"
	"net/http"
	"os"
	"os/signal"
	"runtime"
	"strings"
	"syscall"
	"time"
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

func httpServer(port int16, state *common.State, csc chan common.RequestUpdated) {
	rtr := mux.NewRouter()
	httpapi.Register(rtr, state, csc)
	http.Handle("/", rtr)
	log.Info("HTTP server running on port %d\n", port)
	http.ListenAndServe(fmt.Sprintf(":%d", port), nil)
}

func addCert(state *common.State, stor store.Store, req dao.RequestDao, records []*common.Record, fingerprint string) {
	type Header struct {
		Parent string   `json:"parent"`
		Refs   []string `json:"refs"`
	}

	header := gojws.Header{
		Alg: gojws.ALG_RS256,
		Typ: "cert",
		X5t: utils.GetCertFingerprint(state.WebCert),
	}

	payload, _ := json.Marshal(struct {
		Hdr         Header `json:"header"`
		Fingerprint string `json:"fingerprint"`
	}{
		Header{Parent: records[len(records)-1].Id},
		fingerprint,
	})
	jws, _ := gojws.Sign(header, payload, state.WebKey)

	update := dao.RequestDao{
		Payload: req.Payload + "\n" + jws,
		Version: req.Version + 1,
	}
	stor.UpdateRequest(req.Id, req.Version, update)
	log.Info("Added cert to %d", req.Id)
}

func generateCert(state *common.State, id int64) (err error) {
	stor := dbstore.NewDBStore(state)
	req, _ := stor.GetRequest(id)
	kp := dbstore.KeyProvider{stor}
	records, err := utils.ParseRecords(kp, strings.Split(req.Payload, "\n"))
	if err != nil {
		return
	}
	if !utils.HasRecord(records, "claim") || utils.HasRecord(records, "cert") {
		return
	}
	// Sign the first claim.
	claim, _ := utils.GetFirstRecord(records, "claim")
	pubKey, err := utils.LoadJwk(claim.Header.Jwk)
	if err != nil {
		return
	}
	notBefore := time.Now()
	notAfter := notBefore.Add(2 * 365 * 24 * time.Hour)
	serialNumberLimit := new(big.Int).Lsh(big.NewInt(1), 128)
	serialNumber, err := rand.Int(rand.Reader, serialNumberLimit)
	if err != nil {
		return
	}
	subject := pkix.Name{CommonName: "Test Cert"}
	template := x509.Certificate{
		SerialNumber:          serialNumber,
		Subject:               subject,
		NotBefore:             notBefore,
		NotAfter:              notAfter,
		KeyUsage:              x509.KeyUsageDigitalSignature,
		BasicConstraintsValid: true,
		IsCA: false,
	}
	certDer, err := x509.CreateCertificate(rand.Reader, &template, state.IssueCert, pubKey, state.IssueKey)
	if err != nil {
		return
	}
	cert, _ := x509.ParseCertificate(certDer)
	issuerId, _ := stor.StoreCert(state.IssueCert, 0)
	stor.StoreCert(cert, issuerId)
	fingerprint := utils.GetCertFingerprint(cert)
	addCert(state, stor, req, records, fingerprint)

	return
}

func certSigner(state *common.State, csc chan common.RequestUpdated) {
	for {
		select {
		case c := <-csc:
			if c.Id == state.BootstrapRequestId {
				generateCert(state, c.Id)
			}
		}
	}
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
	state.DB.AutoMigrate(&dao.RequestDao{})
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

	// bootstrap!
	bootstrap.Bootstrap(&state)

	certSignerChan := make(chan common.RequestUpdated, 10)
	go certSigner(&state, certSignerChan)
	go httpServer(8989, &state, certSignerChan)

	log.Info("Ready to handle incoming connections")

	signalHandler()
}
