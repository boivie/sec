package store

import (
	"crypto/x509"
	"github.com/boivie/sec/dao"
)

type Store interface {
	LoadCert(fingerprint string) (*x509.Certificate, error)
	StoreCert(cert *x509.Certificate) error
	CreateRequest(secret int64) (id int64, err error)
	GetRequest(id int64, secret int64) (obj dao.RequestDao, err error)
	UpdateRequest(id int64, oldVersion int32, update dao.RequestDao) (err error)
	StoreTemplate(name string, contents string) (err error)
}
