package app

type MessageTypeCommon struct {
	Resource string `json:"resource"`
	Topic    string `json:"topic,omitempty"`
	Index    int32 `json:"index,omitempty"`
	Parent   string `json:"parent,omitempty"`
	At       int64 `json:"at"`
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