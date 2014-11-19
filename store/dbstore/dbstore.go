package dbstore

import (
	"crypto"
	"crypto/x509"
	"encoding/pem"
	"errors"
	"github.com/boivie/gojws"
	"github.com/boivie/sec/common"
	"github.com/boivie/sec/dao"
	"github.com/boivie/sec/store"
	"github.com/boivie/sec/utils"
	"github.com/op/go-logging"
	"time"
)

var (
	log = logging.MustGetLogger("sec")
)

type DBStore struct {
	state *common.State
}

type KeyProvider struct {
	Stor store.Store
}

func (sk KeyProvider) GetJWSKey(h gojws.Header) (key crypto.PublicKey, err error) {
	if h.X5t != "" {
		cert, err := sk.Stor.LoadCert(h.X5t)
		if err == nil {
			key = cert.PublicKey
		}
	} else if h.Jwk != "" {
		key, err = utils.LoadJwk(h.Jwk)
	} else {
		err = errors.New("No key specified")
	}
	return
}

func (dbs DBStore) CreateRequest(secret int64) (id int64, err error) {
	iDao := dao.RequestDao{
		Secret:    secret,
		CreatedAt: time.Now()}
	dbs.state.DB.Create(&iDao)
	id = iDao.Id
	return
}

func (dbs DBStore) GetRequest(id int64) (obj dao.RequestDao, err error) {
	var iDao dao.RequestDao
	dbs.state.DB.First(&iDao, id)
	if iDao.Id == 0 {
		err = errors.New("Not found")
	} else {
		obj = iDao
	}
	return
}

func (dbs DBStore) UpdateRequest(id int64, version int32, update dao.RequestDao) (err error) {
	count := dbs.state.DB.Table("requests").Where("id = ? AND version = ?", id, version).Updates(update).RowsAffected
	if count != 1 {
		err = errors.New("old_version")
	}
	return
}

func (dbs DBStore) LoadCert(fingerprint string) (cert *x509.Certificate, err error) {
	var cdao dao.CertDao
	dbs.state.DB.Where("fingerprint = ?", fingerprint).First(&cdao)
	if cdao.Id == 0 {
		log.Warning("Key not found: %s", fingerprint)
		return nil, errors.New("Key not found")
	}
	block, _ := pem.Decode([]byte(cdao.Pem))
	if block == nil {
		return nil, errors.New("Invalid PEM")
	}
	cert, err = x509.ParseCertificate(block.Bytes)
	if err != nil {
		return nil, errors.New("Failed to parse certificate")
	}

	return cert, nil
}

func (dbs DBStore) StoreCert(cert *x509.Certificate) (err error) {
	fingerprint := utils.GetCertFingerprint(cert)

	var d dao.CertDao
	dbs.state.DB.Where("fingerprint = ?", fingerprint).First(&d)
	if d.Id != 0 {
		log.Info("Cert %d with fingerprint %s already existed",
			d.Id, fingerprint)
		return
	}

	bytes, err := utils.GetCertPem(cert.Raw)
	if err == nil {
		d = dao.CertDao{
			Fingerprint: fingerprint,
			Pem:         string(bytes),
			NotBefore:   cert.NotBefore,
			NotAfter:    cert.NotAfter,
		}
		dbs.state.DB.Create(&d)
		log.Info("Added cert %d with fingerprint: %s", d.Id, fingerprint)
	}
	return
}

func (dbs DBStore) StoreTemplate(name string, payload string) (err error) {
	update := dao.TemplateDao{Name: name, Payload: payload}
	count := dbs.state.DB.Table("templates").Where("name = ?", name).Updates(update).RowsAffected
	if count == 0 {
		dbs.state.DB.Create(&update)
	}
	return
}

func (dbs DBStore) GetTemplate(name string) (contents string, err error) {
	var t dao.TemplateDao
	dbs.state.DB.Where("name = ?", name).First(&t)
	if t.Id == 0 {
		err = errors.New("not_found")
	} else {
		contents = t.Payload
	}
	return
}

func (dbs DBStore) GetTemplateList() (names []string, err error) {
	var templates []dao.TemplateDao
	dbs.state.DB.Find(&templates)
	for _, template := range templates {
		names = append(names, template.Name)
	}
	return
}

func NewDBStore(state *common.State) store.Store {
	return &DBStore{state}
}
