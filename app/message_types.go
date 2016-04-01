package app
import (
	"time"
	"github.com/boivie/sec/proto"
	"encoding/json"
	"crypto/sha256"
	"github.com/boivie/sec/storage"
	"crypto/rsa"
	"math/big"
	"fmt"
)

type MessageTypeCommon struct {
	Resource string `json:"resource"`
	Root     string `json:"root,omitempty"`
	Topic    string `json:"topic,omitempty"`
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

func stringToBigInt(d string) (*big.Int, error) {
	data, err := Base64URLDecode(d)
	if err != nil {
		fmt.Printf("Failed to base64: '%s' %v\n", d, err)
		return nil, err
	}
	return new(big.Int).SetBytes(data), nil
}

func (key *JsonWebKey) ToPublicKey() (*rsa.PublicKey, error) {
	N, err := stringToBigInt(key.N)
	if err != nil {
		return nil, err
	}
	E, err := stringToBigInt(key.E)
	if err != nil {
		return nil, err
	}

	return &rsa.PublicKey{
		N: N,
		E: int(E.Int64()),
	}, nil
}

type KeyUsageAuditor struct{}
type KeyUsageIssueIdentities struct{}
type KeyUsageRequestSigning struct{}
type KeyUsagePerformSigning struct{}

type KeyUsage struct {
	Auditor         *KeyUsageAuditor `json:"auditor,omitempty"`
	IssueIdentities *KeyUsageIssueIdentities `json:"issue_identities,omitempty"`
	RequestSigning  *KeyUsageRequestSigning `json:"request_signing,omitempty"`
	PerformSigning  *KeyUsagePerformSigning `json:"perform_signing,omitempty"`
}

type RootKey struct {
	Identifier string `json:"identifier"`
	PublicKey  JsonWebKey `json:"public_key"`
	Usage      KeyUsage `json:"usage"`
}

type MessageTypeRootConfig struct {
	MessageTypeCommon
	Keys []RootKey `json:"keys"`
}

func initializeFromParent(target *MessageTypeCommon, root *storage.RecordTopic, topic *storage.RecordTopic, parent *proto.Message) {
	if root != nil {
		target.Root = root.Base58()
	}
	target.At = time.Now().UnixNano() / 1000000
	if parent == nil {
		target.Index = 0
	} else {
		var parentTopic storage.RecordTopic = sha256.Sum256(parent.Signature)
		target.Parent = Base64URLEncode(parentTopic[:])

		var parentHeader MessageTypeCommon
		json.Unmarshal(parent.Payload, &parentHeader)
		target.Index = parentHeader.Index + 1
		if topic != nil {
			target.Topic = topic.Base58()
		}
	}
}


func (m *MessageTypeRootConfig) Initialize(root *storage.RecordTopic, topic *storage.RecordTopic, parent *proto.Message) {
	m.Resource = "root.config"
	initializeFromParent(&m.MessageTypeCommon, root, topic, parent)
}

type MessageTypeAccountCreate struct {
	MessageTypeCommon
}

func (m *MessageTypeAccountCreate) Initialize(root *storage.RecordTopic, topic *storage.RecordTopic, parent *proto.Message) {
	m.Resource = "account.create"
	initializeFromParent(&m.MessageTypeCommon, root, nil, nil)
}

type MessageTypeIdentityOffer struct {
	MessageTypeCommon
	Title string `json:"title"`
}

func (m *MessageTypeIdentityOffer) Initialize(root *storage.RecordTopic, topic *storage.RecordTopic, parent *proto.Message) {
	m.Resource = "identity.offer"
	initializeFromParent(&m.MessageTypeCommon, root, nil, nil)
}

type MessageTypeIdentityClaim struct {
	MessageTypeCommon
	PublicKey JsonWebKey `json:"public_key"`
}

func (m *MessageTypeIdentityClaim) Initialize(root *storage.RecordTopic, topic *storage.RecordTopic, parent *proto.Message) {
	m.Resource = "identity.claim"
	initializeFromParent(&m.MessageTypeCommon, root, topic, parent)
}

type MessageTypeIdentityIssue struct {
	MessageTypeCommon
	Title     string `json:"title"`
	PublicKey JsonWebKey `json:"public_key"`
	Path      string `json:"path"`
}

func (m *MessageTypeIdentityIssue) Initialize(root *storage.RecordTopic, topic *storage.RecordTopic, parent *proto.Message) {
	m.Resource = "identity.issue"
	initializeFromParent(&m.MessageTypeCommon, root, topic, parent)
}

type MessageType interface {
	Initialize(root *storage.RecordTopic, topic *storage.RecordTopic, parent *proto.Message)
	Header() *MessageTypeCommon
}

func (m *MessageTypeCommon) Header() *MessageTypeCommon {
	return m
}

