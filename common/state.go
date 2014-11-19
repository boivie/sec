package common

import (
	"crypto/cipher"
	"crypto/rsa"
	"crypto/x509"
	"github.com/jinzhu/gorm"
)

type RequestUpdated struct {
	Id         int64
	StringId   string
	OldRecords []*Record
	NewRecords []*Record
}

type State struct {
	DB                 gorm.DB
	BaseUrl            string
	IdCrypto           cipher.Block
	BootstrapRequestId int64
	WebKey             *rsa.PrivateKey
	WebCert            *x509.Certificate
	IssueKey           *rsa.PrivateKey
	IssueCert          *x509.Certificate
}
