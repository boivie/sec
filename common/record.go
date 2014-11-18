package common

import (
	"github.com/boivie/gojws"
)

type Record struct {
	Id      string
	Header  gojws.Header
	Payload map[string]interface{}
	New     bool
}
