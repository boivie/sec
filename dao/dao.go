package dao

import (
	"time"
)

type RequestDao struct {
	Id        int64
	Secret    int64
	CreatedAt time.Time
	Payload   string `sql:"type:text"` // newline-separated JWS
}

type Template struct {
	Id      string `json:"id"`
	Name    string `json:"name"`
	Payload string `json:"payload"`
}

type TemplateDao struct {
	Id      int64
	Secret  int64
	Name    string `sql:"size:64"`
	Payload string `sql:"type:text"`
}

type IdentityDao struct {
	Id          int64
	Secret      int64
	ExpiresAt   time.Time
	RevokedAt   time.Time
	IdentityUrl string
	Digest      string
}

type CertDao struct {
	Id          int64
	Fingerprint string `sql:"size:40;unique"`
	Parent      string `sql:"size:40"`
	CreatedAt   time.Time
	NotBefore   time.Time
	NotAfter    time.Time
	Pem         string `sql:"type:text"`
}

func (t TemplateDao) TableName() string {
	return "templates"
}

func (t RequestDao) TableName() string {
	return "requests"
}

func (t IdentityDao) TableName() string {
	return "identities"
}

func (t CertDao) TableName() string {
	return "certs"
}
