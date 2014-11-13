package httpapi

import (
	"bytes"
	"crypto"
	"crypto/sha1"
	"crypto/sha256"
	"crypto/x509"
	"encoding/base32"
	"encoding/base64"
	"encoding/binary"
	"encoding/hex"
	"encoding/json"
	"encoding/pem"
	"errors"
	"github.com/boivie/gojws"
	"github.com/boivie/sec/common"
	"github.com/boivie/sec/dao"
	"github.com/gorilla/mux"
	"github.com/op/go-logging"
	"io"
	"io/ioutil"
	"math/rand"
	"net/http"
	"strconv"
	"strings"
	"time"
)

var (
	log   = logging.MustGetLogger("sec")
	state *common.State
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
		Id  string `json:"id"`
		Uri string `json:"url"`
	}{
		invitationId,
		state.BaseUrl + "/invitation/" + invitationId,
	})
}

type keyProvider struct {
}

func (sk keyProvider) GetJWSKey(h gojws.Header) (crypto.PublicKey, error) {
	var cert dao.CertDao
	state.DB.Where("fingerprint = ?", h.X5t).First(&cert)
	if cert.Id == 0 {
		log.Warning("Key not found: %s", h.X5t)
		return nil, errors.New("Key not found")
	}
	block, _ := pem.Decode([]byte(cert.Pem))
	if block == nil {
		return nil, errors.New("Invalid PEM")
	}
	x5c, err := x509.ParseCertificate(block.Bytes)
	if err != nil {
		return nil, errors.New("Failed to parse certificate")
	}
	return x5c.PublicKey, nil
}

func parseJws(tokenString string) (header gojws.Header, payload map[string]interface{}, err error) {
	kp := keyProvider{}
	var data []byte
	header, data, err = gojws.VerifyAndDecodeWithHeader(tokenString, kp)
	err = json.Unmarshal(data, &payload)
	return
}

func calculateHash(jws string) (ret string) {
	lastDot := strings.LastIndex(jws, ".")
	part12 := jws[0:lastDot]
	hash := sha256.New()
	hash.Write([]byte(part12))

	var buf bytes.Buffer
	encoder := base64.NewEncoder(base64.StdEncoding, &buf)
	encoder.Write(hash.Sum(nil))
	encoder.Close()
	return buf.String()
}

func validateChain(jwss []string) error {
	var parent string = ""
	for idx, jws := range jwss {
		header, payload, err := parseJws(jws)
		if err != nil {
			log.Warning("Entry %d, fail: %s",
				idx, err)
			return errors.New("jws_parse_failed")
		}
		hash := calculateHash(jws)
		log.Info("Processing jws %d (type %s, hash %s)",
			idx, header.Typ, hash)
		if idx == 0 {
			if _, present := payload["parent"]; present {
				log.Warning("Entry %d has parent %s, shouldn't",
					idx, payload["parent"])
				return errors.New("invalid_parent")
			}
		} else {
			if payload["parent"] != parent {
				log.Warning("Entry %d has parent %s != %s",
					idx, payload["parent"], parent)
				return errors.New("invalid_parent")
			}
		}
		parent = hash
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

func checkIdempotency(haystack string, needle string) bool {
	cleanNeedle := strings.Join(cleanAndSplit(needle), "\n")
	return strings.Contains(haystack, cleanNeedle)
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
	if checkIdempotency(iDao.Payload, string(b)) {
		jsonify(c, struct{}{})
		return
	}
	// TODO: Validate past and present jwss
	var jwss []string
	jwss = append(cleanAndSplit(iDao.Payload), cleanAndSplit(string(b))...)

	if err = validateChain(jwss); err != nil {
		http.Error(c, err.Error(), 400)
		return
	}
	iDao.Payload = strings.TrimSpace(strings.Join(jwss, "\n"))
	state.DB.Save(&iDao)
	jsonify(c, struct{}{})
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

func addPem(ascii string) {
	block, _ := pem.Decode([]byte(ascii))
	if block == nil {
		log.Warning("Invalid PEM")
		return
	}
	cert, err := x509.ParseCertificate(block.Bytes)
	if err != nil {
		log.Warning("Failed to parse PEM: %s", err)
		return
	}
	hash := sha1.New()
	hash.Write(cert.Raw)
	fingerprint := hex.EncodeToString(hash.Sum(nil))
	var d dao.CertDao

	state.DB.Where("fingerprint = ?", fingerprint).First(&d)
	if d.Id != 0 {
		log.Info("Cert %d with fingerprint %s already existed",
			d.Id, fingerprint)
	} else {
		d := dao.CertDao{
			Fingerprint: fingerprint,
			Pem:         ascii,
			NotBefore:   cert.NotBefore,
			NotAfter:    cert.NotAfter,
		}
		state.DB.Create(&d)
		log.Info("Adding cert %d with fingerprint: %s", d.Id, fingerprint)
	}
}

func AddCertificate(c http.ResponseWriter, r *http.Request) {
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
				pem := strings.Join(lines, "\n")
				addPem(pem)
				capturing = false
			}
		}
	}
}

func Register(rtr *mux.Router, _state *common.State) {
	state = _state
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
	rtr.HandleFunc("/cert/",
		AddCertificate).Methods("POST")

}
