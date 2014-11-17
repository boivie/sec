package httpapi

import (
	"bytes"
	"crypto/x509"
	"encoding/base32"
	"encoding/binary"
	"encoding/json"
	"encoding/pem"
	"errors"
	"github.com/boivie/gojws"
	"github.com/boivie/sec/common"
	"github.com/boivie/sec/dao"
	"github.com/boivie/sec/store"
	"github.com/boivie/sec/store/dbstore"
	"github.com/boivie/sec/utils"
	"github.com/gorilla/mux"
	"github.com/op/go-logging"
	"io"
	"io/ioutil"
	"math/rand"
	"net/http"
	"strconv"
	"strings"
)

var (
	log   = logging.MustGetLogger("sec")
	state *common.State
	stor  store.Store
)

func jsonify(c http.ResponseWriter, s interface{}) {
	var encoded []byte
	if str, ok := s.(string); ok {
		encoded = []byte(str)
	} else {
		encoded, _ = json.MarshalIndent(s, "", "  ")
	}
	c.Header().Add("Content-Type", "application/json")
	c.Header().Add("Content-Length", strconv.Itoa(len(encoded)+1))
	c.Write(encoded)
	io.WriteString(c, "\n")
}

func GetTemplateList(c http.ResponseWriter, r *http.Request) {
	names, _ := stor.GetTemplateList()

	s := struct {
		Templates []string `json:"templates"`
	}{
		names,
	}
	jsonify(c, s)
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

func generateSecret() int64 {
	return rand.Int63()
}

func getStringId(id int64, secret int64) string {
	buf := new(bytes.Buffer)
	binary.Write(buf, binary.LittleEndian, id)
	binary.Write(buf, binary.LittleEndian, secret)
	bytes := buf.Bytes()

	encrypted := make([]byte, 16)
	state.IdCrypto.Encrypt(encrypted, bytes)

	encoder := base32.StdEncoding
	return strings.Replace(encoder.EncodeToString(encrypted), "=", "", -1)
}

func parseInvitationId(id string) (dbId int64, secret int64, err error) {
	for {
		if len(id)%8 == 0 {
			break
		}
		id = id + "="
	}
	data, err := base32.StdEncoding.DecodeString(id)
	if err != nil {
		return
	}
	decrypted := make([]byte, 16)
	state.IdCrypto.Decrypt(decrypted, data)

	buf := bytes.NewBuffer(decrypted)
	binary.Read(buf, binary.LittleEndian, &dbId)
	binary.Read(buf, binary.LittleEndian, &secret)

	return
}

func CreateRequest(c http.ResponseWriter, r *http.Request) {
	secret := generateSecret()
	id, err := stor.CreateRequest(secret)
	if err != nil {
		log.Error("Failed to create request: %v", err)
		http.Error(c, "internal_error", 500)
		return
	}
	invitationId := getStringId(id, secret)
	log.Info("Created invitation %s", invitationId)

	jsonify(c, struct {
		Id    string `json:"id"`
		Uri   string `json:"url"`
		Qruri string `json:"qr_url"`
	}{
		invitationId,
		state.BaseUrl + "/request/" + invitationId,
		strings.ToUpper(state.BaseUrl) + "/R/" + invitationId,
	})
}

type Part struct {
	header  gojws.Header
	payload map[string]interface{}
	hash    string
}

type PartValidator func(invitation dao.RequestDao, parts []Part, idx int) error

func ValidateInvitation(iDao dao.RequestDao, parts []Part, idx int) error {
	invitationId := getStringId(iDao.Id, iDao.Secret)
	if parts[idx].payload["invitation_id"] != invitationId {
		log.Warning("Invitation part's invitation id doesn't match")
		return errors.New("invalid_invitation_id")
	}
	return nil
}

var validators = map[string]PartValidator{
	"invitation": ValidateInvitation,
}

type Header struct {
	Parent string   `json:"parent"`
	Refs   []string `json:"refs"`
}

func validateHeaders(parts []Part) bool {
	hash2part := make(map[string]Part)
	for _, part := range parts {
		hash2part[part.hash] = part
	}

	var parent string = ""
	for idx, part := range parts {
		log.Info("Processing jws %d (type %s, hash %s)",
			idx, part.header.Typ, part.hash)

		var header Header
		headerJson, _ := json.Marshal(part.payload["header"])
		if err := json.Unmarshal(headerJson, &header); err != nil {
			return false
		}

		for _, ref := range header.Refs {
			if _, present := hash2part[ref]; !present {
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
		parent = part.hash
	}
	return true
}

func validateChain(invitation dao.RequestDao, jwss []string) error {
	parts := make([]Part, 0)

	for idx, jws := range jwss {
		kp := dbstore.KeyProvider{stor}
		header, payload, err := utils.ParseJws(jws, kp)
		if err != nil {
			log.Warning("Entry %d, fail: %v",
				idx, err)
			return errors.New("jws_parse_failed")
		}
		hash := utils.GetFingerprint([]byte(jws))
		parts = append(parts, Part{header, payload, hash})
	}

	if !validateHeaders(parts) {
		return errors.New("invalid_parents")
	}

	for idx, part := range parts {
		if validator, ok := validators[part.header.Typ]; ok {
			if err := validator(invitation, parts, idx); err != nil {
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
	dbId, secret, err := parseInvitationId(params["id"])
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

	iDao, err := stor.GetRequest(dbId, secret)
	if err != nil {
		log.Warning("UpdateRequest(%s) -> %v", params["id"], err)
		http.NotFound(c, r)
		return
	}

	jwss := append(cleanAndSplit(iDao.Payload),
		filterNew(iDao.Payload, string(b))...)

	last_hash := utils.GetFingerprint([]byte(jwss[len(jwss)-1]))

	if err = validateChain(iDao, jwss); err != nil {
		http.Error(c, err.Error(), 400)
		return
	}
	payload := strings.TrimSpace(strings.Join(jwss, "\n"))

	update := dao.RequestDao{Version: iDao.Version + 1, Payload: payload}
	err = stor.UpdateRequest(dbId, iDao.Version, update)
	if err != nil {
		http.Error(c, "request_stale", 400)
		return
	}

	jsonify(c, struct {
		Hash string `json:"hash"`
	}{last_hash})
}

func GetRequest(c http.ResponseWriter, r *http.Request) {
	params := mux.Vars(r)
	dbId, secret, err := parseInvitationId(params["id"])
	if err != nil {
		log.Info("Invalid id:", err)
		http.NotFound(c, r)
	} else {
		obj, err := stor.GetRequest(dbId, secret)
		if err != nil {
			log.Warning("GetRequest(%s) -> %v", params["id"], err)
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
	err = stor.StoreCert(cert)
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
	jsonify(c, struct {
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
		pemString, err := utils.GetCertPem(cert.Raw)
		if err != nil {
			log.Error("Failed to parse cert: %v", err)
			http.Error(c, "internal_error", 500)
		} else {
			io.WriteString(c, pemString)
			io.WriteString(c, "\n")
		}
	}
}

func Register(rtr *mux.Router, _state *common.State) {
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

}
