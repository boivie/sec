package httpapi

import (
	"bytes"
	"crypto"
	"crypto/rsa"
	"crypto/x509"
	"encoding/base32"
	"encoding/binary"
	"encoding/json"
	"encoding/pem"
	"errors"
	"fmt"
	"github.com/boivie/gojws"
	"github.com/boivie/sec/common"
	"github.com/boivie/sec/dao"
	"github.com/boivie/sec/store"
	"github.com/boivie/sec/utils"
	"github.com/gorilla/mux"
	"github.com/op/go-logging"
	"io"
	"io/ioutil"
	"math/big"
	"math/rand"
	"net/http"
	"strconv"
	"strings"
	"time"
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

func GetInvitationTemplateList(c http.ResponseWriter, r *http.Request) {
	type TemplateBrief struct {
		Id   string `json:"id"`
		Name string `json:"name"`
		Url  string `json:"url"`
	}
	var ret []TemplateBrief

	var daos []dao.TemplateDao
	state.DB.Find(&daos)

	for _, t := range daos {
		sid := getStringId(t.Id, t.Secret)
		url := state.BaseUrl + "/identity/template/" + sid
		ret = append(ret, TemplateBrief{sid, t.Name, url})
	}

	s := struct {
		Templates []TemplateBrief `json:"templates"`
	}{
		ret,
	}
	jsonify(c, s)
}

func GetInvitationTemplate(c http.ResponseWriter, r *http.Request) {
	params := mux.Vars(r)
	if id, err := strconv.Atoi(params["id"]); err != nil {
		http.NotFound(c, r)
		return
	} else {
		var t_dao dao.TemplateDao
		state.DB.First(&t_dao, id)
		if t_dao.Id == 0 {
			http.NotFound(c, r)
		} else {
			ret := dao.Template{
				Id:      getStringId(t_dao.Id, t_dao.Secret),
				Name:    t_dao.Name,
				Payload: t_dao.Payload}
			jsonify(c, ret)
		}
	}
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

func CreateInvitation(c http.ResponseWriter, r *http.Request) {
	iDao := dao.InvitationDao{
		Secret:    generateSecret(),
		CreatedAt: time.Now()}
	state.DB.Create(&iDao)

	invitationId := getStringId(iDao.Id, iDao.Secret)
	log.Info("Created invitation %s", invitationId)

	jsonify(c, struct {
		Id    string `json:"id"`
		Uri   string `json:"url"`
		Qruri string `json:"qr_url"`
	}{
		invitationId,
		state.BaseUrl + "/invitation/" + invitationId,
		strings.ToUpper(state.BaseUrl) + "/R/" + invitationId,
	})
}

type keyProvider struct {
}

func loadJwk(jwk string) (crypto.PublicKey, error) {
	var key struct {
		Kty string `json:"kty"`
		N   string `json:"n"`
		E   string `json:"e"`
	}
	err := json.Unmarshal([]byte(jwk), &key)
	if err != nil {
		return nil, fmt.Errorf("Failed to unmarshal key: %v", err)
	}

	switch key.Kty {
	case "RSA":
		if key.N == "" || key.E == "" {
			return nil, errors.New("Malformed JWS RSA key")
		}

		data, err := utils.B64decode(key.E)
		if err != nil {
			return nil, errors.New("Malformed JWS RSA key")
		}
		if len(data) < 4 {
			ndata := make([]byte, 4)
			copy(ndata[4-len(data):], data)
			data = ndata
		}

		pubKey := &rsa.PublicKey{
			N: &big.Int{},
			E: int(binary.BigEndian.Uint32(data[:])),
		}

		data, err = utils.B64decode(key.N)
		if err != nil {
			return nil, errors.New("Malformed JWS RSA key")
		}
		pubKey.N.SetBytes(data)

		return pubKey, nil

	default:
		return nil, fmt.Errorf("Unknown JWS key type %s", key.Kty)
	}
}

func (sk keyProvider) GetJWSKey(h gojws.Header) (key crypto.PublicKey, err error) {
	if h.X5t != "" {
		cert, err := stor.LoadCert(h.X5t)
		if err == nil {
			key = cert.PublicKey
		}
	} else if h.Jwk != "" {
		key, err = loadJwk(h.Jwk)
	} else {
		err = errors.New("No key specified")
	}
	return
}

func parseJws(tokenString string) (header gojws.Header, payload map[string]interface{}, err error) {
	kp := keyProvider{}
	var data []byte
	header, data, err = gojws.VerifyAndDecodeWithHeader(tokenString, kp)
	if err != nil {
		log.Warning("%v", err)
	}
	err = json.Unmarshal(data, &payload)
	return
}

type Part struct {
	header  gojws.Header
	payload map[string]interface{}
	hash    string
}

type PartValidator func(invitation dao.InvitationDao, parts []Part, idx int) error

func ValidateInvitation(iDao dao.InvitationDao, parts []Part, idx int) error {
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

func validateChain(invitation dao.InvitationDao, jwss []string) error {
	parts := make([]Part, 0)

	for idx, jws := range jwss {
		header, payload, err := parseJws(jws)
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

func UpdateInvitation(c http.ResponseWriter, r *http.Request) {
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
	var iDao dao.InvitationDao

	// TODO: "FOR UPDATE"
	state.DB.First(&iDao, dbId)
	if iDao.Id == 0 {
		http.NotFound(c, r)
		return
	}
	if iDao.Secret != secret {
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
	iDao.Payload = strings.TrimSpace(strings.Join(jwss, "\n"))
	state.DB.Save(&iDao)
	jsonify(c, struct {
		Hash string `json:"hash"`
	}{last_hash})
}

func GetInvitation(c http.ResponseWriter, r *http.Request) {
	params := mux.Vars(r)
	dbId, secret, err := parseInvitationId(params["id"])
	if err != nil {
		log.Info("Invalid id:", err)
		http.NotFound(c, r)
	} else {
		var iDao dao.InvitationDao
		state.DB.First(&iDao, dbId)
		if iDao.Id == 0 {
			http.NotFound(c, r)
			return
		}
		if iDao.Secret != secret {
			http.NotFound(c, r)
			return
		}

		io.WriteString(c, iDao.Payload)
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
			http.Error(c, err.Error(), 500)
		} else {
			io.WriteString(c, pemString)
			io.WriteString(c, "\n")
		}
	}
}

func Register(rtr *mux.Router, _state *common.State) {
	state = _state
	stor = store.NewDBStore(_state)
	rtr.HandleFunc("/identity/template/",
		GetInvitationTemplateList).Methods("GET")
	rtr.HandleFunc("/identity/template/{id:[0-9]+}",
		GetInvitationTemplate).Methods("GET")
	rtr.HandleFunc("/invitation/",
		CreateInvitation).Methods("POST")
	rtr.HandleFunc("/invitation/{id:[A-Za-z0-9]+}",
		UpdateInvitation).Methods("POST")
	rtr.HandleFunc("/invitation/{id:[A-Za-z0-9]+}",
		GetInvitation).Methods("GET")
	rtr.HandleFunc("/R/{id:[A-Za-z0-9]+}",
		GetInvitation).Methods("GET")
	rtr.HandleFunc("/cert/",
		AddCertificate).Methods("POST")
	rtr.HandleFunc("/cert/{fingerprint:[a-zA-Z0-9_-]+}",
		GetCertificate).Methods("GET")

}
