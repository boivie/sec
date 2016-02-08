package app
import (
	"time"
	"github.com/boivie/sec/proto"
	"encoding/json"
	"crypto/sha256"
)

type MessageTypeCommon struct {
	Resource string `json:"resource"`
	Index    int32 `json:"index,omitempty"`
	Parent   string `json:"parent,omitempty"`
	At       int64 `json:"at"`
	Ref      string `json:"ref,omitempty"`
}

type JsonWebKey struct {
	Kid string `json:"kid,omitempty"`
	Kty string `json:"kty,omitempty"`
	Alg string `json:"alg,omitempty"`
	N   string `json:"n,omitempty"`
	E   string `json:"e,omitempty"`
}

type Root struct {
	PublicKey JsonWebKey `json:"public_key"`
}

type MessageTypeRootConfigRoots struct {
	AuditorRoots  []Root `json:"auditor_roots"`
	IdentityRoots []Root `json:"identity_roots"`
}

type MessageTypeRootConfig struct {
	MessageTypeCommon
	Roots MessageTypeRootConfigRoots `json:"roots"`
}

func initializeFromParent(target *MessageTypeCommon, parent *proto.Record) {
	target.At = time.Now().UnixNano() / 1000000
	if parent == nil {
		target.Index = 0
	} else {
		signatureHash := sha256.Sum256(parent.Message.Signature)

		var parentHeader MessageTypeCommon
		json.Unmarshal(parent.Message.Payload, &parentHeader)
		target.Index = parentHeader.Index + 1
		target.Parent = Base64URLEncode(signatureHash[:])
	}
}


func (m *MessageTypeRootConfig) Initialize(parent *proto.Record) {
	m.Resource = "root.config"
	initializeFromParent(&m.MessageTypeCommon, parent)
}

type MessageTypeAccountCreate struct {
	MessageTypeCommon
}

func (m *MessageTypeAccountCreate) Initialize(parent *proto.Record) {
	m.Resource = "account.create"
	initializeFromParent(&m.MessageTypeCommon, parent)
}

type MessageTypeIdentityOffer struct {
	MessageTypeCommon
	Title string `json:"title"`
}

func (m *MessageTypeIdentityOffer) Initialize(parent *proto.Record) {
	m.Resource = "identity.offer"
	initializeFromParent(&m.MessageTypeCommon, parent)
}

type MessageTypeIdentityClaim struct {
	MessageTypeCommon
}

func (m *MessageTypeIdentityClaim) Initialize(parent *proto.Record) {
	m.Resource = "identity.claim"
	initializeFromParent(&m.MessageTypeCommon, parent)
}

type MessageTypeIdentityIssue struct {
	MessageTypeCommon
	Title     string `json:"title"`
	PublicKey *JsonWebKey `json:"public_key"`
	Path      string `json:"path"`
}

func (m *MessageTypeIdentityIssue) Initialize(parent *proto.Record) {
	m.Resource = "identity.issue"
	initializeFromParent(&m.MessageTypeCommon, parent)
}

type MessageType interface {
	Initialize(parent *proto.Record)
	Header() *MessageTypeCommon
}

func (m *MessageTypeCommon) Header() *MessageTypeCommon {
	return m
}