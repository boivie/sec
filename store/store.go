package store

import (
	"crypto/x509"
	"github.com/boivie/sec/dao"
)

type CertInfo struct {
	Id          int64
	Parent      int64
	Fingerprint string
	Cert        *x509.Certificate
}

type Store interface {
	LoadCert(fingerprint string) (*CertInfo, error)
	StoreCert(cert *x509.Certificate) (id int64, err error)
	CreateRequest(secret int64) (id int64, err error)
	GetRequest(id int64) (obj dao.RequestDao, err error)
	UpdateRequest(id int64, oldVersion int32, update dao.RequestDao) (err error)
	StoreTemplate(name string, contents string) (err error)
	GetTemplate(name string) (contents string, err error)
	GetTemplateList() (names []string, err error)
}
