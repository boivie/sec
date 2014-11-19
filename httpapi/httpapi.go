package httpapi

import (
	"crypto/x509"
	"encoding/json"
	"encoding/pem"
	"errors"
	"github.com/boivie/sec/common"
	"github.com/boivie/sec/dao"
	"github.com/boivie/sec/store"
	"github.com/boivie/sec/store/dbstore"
	"github.com/boivie/sec/utils"
	"github.com/gorilla/mux"
	"github.com/op/go-logging"
	"html/template"
	"io"
	"io/ioutil"
	"net/http"
	"strconv"
	"strings"
)

var (
	log            = logging.MustGetLogger("sec")
	state          *common.State
	stor           store.Store
	certSignerChan chan common.RequestUpdated
)

func GetTemplateList(c http.ResponseWriter, r *http.Request) {
	names, _ := stor.GetTemplateList()

	s := struct {
		Templates []string `json:"templates"`
	}{
		names,
	}
	utils.Jsonify(c, s)
}

func GetTemplate(c http.ResponseWriter, r *http.Request) {
	params := mux.Vars(r)
	name := params["name"]
	t, err := stor.GetTemplate(name)
	if err != nil {
		log.Info("Invalid template: %v", err)
		http.NotFound(c, r)
		return
	}
	c.Header().Add("Content-Length", strconv.Itoa(len(t)+1))
	io.WriteString(c, t)
	io.WriteString(c, "\n")
}

func CreateRequest(c http.ResponseWriter, r *http.Request) {
	secret := utils.GenerateSecret()
	id, err := stor.CreateRequest(secret)
	if err != nil {
		log.Error("Failed to create request: %v", err)
		http.Error(c, "internal_error", 500)
		return
	}
	invitationId := utils.GetStringId(id, secret, state.IdCrypto)
	log.Info("Created invitation %s", invitationId)

	utils.Jsonify(c, struct {
		Id    string `json:"id"`
		Uri   string `json:"url"`
		Qruri string `json:"qr_url"`
	}{
		invitationId,
		state.BaseUrl + "/request/" + invitationId,
		strings.ToUpper(state.BaseUrl) + "/R/" + invitationId,
	})
}

type RecordValidator func(invitation dao.RequestDao, records []*common.Record, idx int) error

func ValidateInvitation(iDao dao.RequestDao, records []*common.Record, idx int) error {
	invitationId := utils.GetStringId(iDao.Id, iDao.Secret, state.IdCrypto)
	if records[idx].Payload["invitation_id"] != invitationId {
		log.Warning("Invitation record's invitation id doesn't match")
		return errors.New("invalid_invitation_id")
	}
	return nil
}

var validators = map[string]RecordValidator{
	"invitation": ValidateInvitation,
}

type Header struct {
	Parent string   `json:"parent"`
	Refs   []string `json:"refs"`
}

func validateHeaders(records []*common.Record) bool {
	hash2record := make(map[string]*common.Record)
	for _, record := range records {
		hash2record[record.Id] = record
	}

	var parent string = ""
	for idx, record := range records {
		log.Info("Processing jws %d (type %s, id %s)",
			idx, record.Header.Typ, record.Id)

		var header Header
		headerJson, _ := json.Marshal(record.Payload["header"])
		if err := json.Unmarshal(headerJson, &header); err != nil {
			return false
		}

		for _, ref := range header.Refs {
			if _, present := hash2record[ref]; !present {
				log.Warning("Entry %d has ref %s, not found",
					idx, ref)
				return false
			}
		}

		if idx == 0 {
			if header.Parent != "" {
				log.Warning("Entry %d has parent %s, shouldn't",
					idx, header.Parent)
				return false
			}
		} else {
			if header.Parent != parent {
				log.Warning("Entry %d has parent %s != %s",
					idx, header.Parent, parent)
				return false
			}
		}
		parent = record.Id
	}
	return true
}

func validateRecords(req dao.RequestDao, records []*common.Record) error {
	if !validateHeaders(records) {
		return errors.New("invalid_parents")
	}

	for idx, record := range records {
		if validator, ok := validators[record.Header.Typ]; ok {
			if err := validator(req, records, idx); err != nil {
				return err
			}
		}
	}

	return nil
}

func cleanAndSplit(s string) (ret []string) {
	ret = make([]string, 0)
	for _, line := range strings.Split(s, "\n") {
		line := strings.TrimSpace(line)
		if line != "" {
			ret = append(ret, line)
		}
	}
	return
}

func filterNew(current string, requested string) []string {
	ret := make([]string, 0)
	for _, jws := range cleanAndSplit(requested) {
		if !strings.Contains(current, jws) {
			ret = append(ret, jws)
		}
	}
	return ret
}

func UpdateRequest(c http.ResponseWriter, r *http.Request) {
	params := mux.Vars(r)
	dbId, secret, err := utils.ParseStringId(params["id"], state.IdCrypto)
	if err != nil {
		log.Info("Invalid id:", err)
		http.NotFound(c, r)
		return
	}
	b, err := ioutil.ReadAll(r.Body)
	if err != nil {
		log.Info("Invalid JWS:", err)
		http.NotFound(c, r)
		return
	}

	iDao, err := stor.GetRequest(dbId)
	if err != nil {
		log.Warning("UpdateRequest(%s) -> %v", params["id"], err)
		http.NotFound(c, r)
		return
	} else if iDao.Secret != secret {
		log.Warning("UpdateRequest(%s) -> wrong secret")
		http.NotFound(c, r)
		return
	}

	olds := cleanAndSplit(iDao.Payload)
	news := filterNew(iDao.Payload, string(b))
	kp := dbstore.KeyProvider{stor}

	orecords, err := utils.ParseRecords(kp, olds)
	if err != nil {
		http.Error(c, err.Error(), 400)
		return
	}
	nrecords, err := utils.ParseRecords(kp, news)
	if err != nil {
		http.Error(c, err.Error(), 400)
		return
	}
	arecords := make([]*common.Record, len(orecords)+len(nrecords))
	copy(arecords, orecords)
	copy(arecords[len(orecords):], nrecords)

	if err := validateRecords(iDao, arecords); err != nil {
		http.Error(c, err.Error(), 400)
		return
	}

	payload := strings.TrimSpace(strings.Join(append(olds, news...), "\n"))

	update := dao.RequestDao{Version: iDao.Version + 1, Payload: payload}
	err = stor.UpdateRequest(dbId, iDao.Version, update)
	if err != nil {
		http.Error(c, "request_stale", 400)
		return
	}

	event := common.RequestUpdated{
		dbId,
		params["id"],
		orecords,
		nrecords,
	}
	certSignerChan <- event

	utils.Jsonify(c, struct {
		Hash string `json:"hash"`
	}{arecords[len(arecords)-1].Id})
}

func GetRequest(c http.ResponseWriter, r *http.Request) {
	params := mux.Vars(r)
	dbId, secret, err := utils.ParseStringId(params["id"], state.IdCrypto)
	if err != nil {
		log.Info("Invalid id:", err)
		http.NotFound(c, r)
	} else {
		obj, err := stor.GetRequest(dbId)
		if err != nil {
			log.Warning("GetRequest(%s) -> %v", params["id"], err)
			return
		} else if obj.Secret != secret {
			log.Info("Invalid id:", err)
			http.NotFound(c, r)
			return
		}
		io.WriteString(c, obj.Payload)
		io.WriteString(c, "\n")
	}
}

func addPem(ascii string) (fingerprint string, err error) {
	block, _ := pem.Decode([]byte(ascii))
	if block == nil {
		log.Warning("Invalid PEM")
		err = errors.New("invalid_pem")
		return
	}
	cert, err := x509.ParseCertificate(block.Bytes)
	if err != nil {
		log.Warning("Failed to parse PEM: %s", err)
		err = errors.New("invalid_pem")
		return
	}

	fingerprint = utils.GetCertFingerprint(cert)
	_, err = stor.StoreCert(cert)
	return
}

func AddCertificate(c http.ResponseWriter, r *http.Request) {
	type CertRet struct {
		Fingerprint string `json:"fingerprint"`
		Url         string `json:"url"`
	}
	ret := make([]CertRet, 0)

	b, err := ioutil.ReadAll(r.Body)
	if err != nil {
		log.Info("Invalid cert:", err)
		http.NotFound(c, r)
		return
	}
	var lines = make([]string, 0)
	var capturing = false
	for _, line := range strings.Split(string(b), "\n") {
		if strings.HasPrefix(line, "-----BEGIN CERTIFICATE-----") {
			capturing = true
			lines = make([]string, 0)
		}
		if capturing {
			lines = append(lines, line)
			if strings.HasPrefix(line, "-----END CERTIFICATE-----") {
				pem := strings.TrimSpace(strings.Join(lines, "\n"))
				fingerprint, err := addPem(pem)
				if err != nil {
					http.Error(c, err.Error(), 400)
					return
				}
				ret = append(ret, CertRet{
					fingerprint,
					state.BaseUrl + "/cert/" + fingerprint,
				})
				capturing = false
			}
		}
	}
	utils.Jsonify(c, struct {
		Certs []CertRet `json:"certs"`
	}{
		ret,
	})
}

func GetCertificate(c http.ResponseWriter, r *http.Request) {
	params := mux.Vars(r)
	fingerprint := params["fingerprint"]

	cert, err := stor.LoadCert(fingerprint)
	if err != nil {
		http.NotFound(c, r)
	} else {
		pemString, err := utils.GetCertPem(cert.Cert.Raw)
		if err != nil {
			log.Error("Failed to parse cert: %v", err)
			http.Error(c, "internal_error", 500)
		} else {
			io.WriteString(c, pemString)
			io.WriteString(c, "\n")
		}
	}
}

func GetStart(c http.ResponseWriter, r *http.Request) {
	const START_TMPL = `
First Install the app<p>
After that, <a href="/request/{{.RequestId}}">install your first certificate</a>
`
	req, err := stor.GetRequest(state.BootstrapRequestId)
	if err != nil {
		http.Error(c, "internal_error", 500)
	}
	tmpl, err := template.New("start").Parse(START_TMPL)
	if err != nil {
		http.Error(c, "internal_error", 500)
	} else {
		tmpl.Execute(c, struct {
			RequestId string
		}{
			utils.GetStringId(req.Id, req.Secret, state.IdCrypto),
		})
	}
}

func Register(rtr *mux.Router, _state *common.State, csc chan common.RequestUpdated) {
	certSignerChan = csc
	state = _state
	stor = dbstore.NewDBStore(_state)
	rtr.HandleFunc("/template/",
		GetTemplateList).Methods("GET")
	rtr.HandleFunc("/template/{name:[a-z-]+}",
		GetTemplate).Methods("GET")
	rtr.HandleFunc("/request/",
		CreateRequest).Methods("POST")
	rtr.HandleFunc("/request/{id:[A-Za-z0-9]+}",
		UpdateRequest).Methods("POST")
	rtr.HandleFunc("/request/{id:[A-Za-z0-9]+}",
		GetRequest).Methods("GET")
	rtr.HandleFunc("/R/{id:[A-Za-z0-9]+}",
		GetRequest).Methods("GET")
	rtr.HandleFunc("/cert/",
		AddCertificate).Methods("POST")
	rtr.HandleFunc("/cert/{fingerprint:[a-zA-Z0-9_-]+}",
		GetCertificate).Methods("GET")
	rtr.HandleFunc("/start",
		GetStart).Methods("GET")

}
