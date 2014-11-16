package store

import (
	"crypto/x509"
	"encoding/pem"
	"errors"
	"github.com/boivie/sec/common"
	"github.com/boivie/sec/dao"
	"github.com/boivie/sec/utils"
	"github.com/op/go-logging"
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
