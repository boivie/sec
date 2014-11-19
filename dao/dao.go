package dao

import (
	"time"
)

type RequestDao struct {
	Id         int64
	Secret     int64
	Version    int32
	CreatedAt  time.Time
	ModifiedAt time.Time
	Payload    string `sql:"type:text"` // newline-separated JWS
}

type Template struct {
	Id      string `json:"id"`
	Name    string `json:"name"`
	Payload string `json:"payload"`
}

type TemplateDao struct {
	Id      int64
	Name    string `sql:"size:32;unique"`
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
	Fingerprint string `sql:"size:28;unique"`
	Parent      int64
	CreatedAt   time.Time
	NotBefore   time.Time
	NotAfter    time.Time
	Der         string `sql:"type:text"`
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
