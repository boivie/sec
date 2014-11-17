package store

import (
	"crypto/x509"
	"encoding/pem"
	"errors"
	"github.com/boivie/sec/common"
	"github.com/boivie/sec/dao"
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

type Store interface {
	LoadCert(fingerprint string) (*x509.Certificate, error)
	StoreCert(cert *x509.Certificate) error
	CreateRequest(secret int64) (id int64, err error)
	GetRequest(id int64, secret int64) (obj dao.RequestDao, err error)
	UpdateRequest(id int64, oldVersion int32, update dao.RequestDao) (err error)
}

func (dbs DBStore) CreateRequest(secret int64) (id int64, err error) {
	iDao := dao.RequestDao{
		Secret:    secret,
		CreatedAt: time.Now()}
	dbs.state.DB.Create(&iDao)
	id = iDao.Id
	return
}

func (dbs DBStore) GetRequest(id int64, secret int64) (obj dao.RequestDao, err error) {
	var iDao dao.RequestDao
	dbs.state.DB.First(&iDao, id)
	if iDao.Id == 0 {
		err = errors.New("Not found")
	} else if iDao.Secret != secret {
		err = errors.New("Bad secret")
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

func NewDBStore(state *common.State) Store {
	return &DBStore{state}
}
