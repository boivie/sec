package common

import (
	"crypto/cipher"
	"github.com/jinzhu/gorm"
)

type State struct {
	DB                 gorm.DB
	BaseUrl            string
	IdCrypto           cipher.Block
	BootstrapRequestId string
}
