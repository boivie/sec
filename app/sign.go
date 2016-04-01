package app
import (
	. "github.com/boivie/sec/proto"
	"github.com/boivie/sec/storage"
	"crypto/sha256"
	jose "github.com/square/go-jose"
	"encoding/json"
	"bytes"
	"encoding/binary"
	"crypto/cipher"
	"crypto/aes"
	"github.com/golang/protobuf/proto"
)

func CreateNonce(isMessage bool, index int32) []byte {
	buf := bytes.Buffer{}
	err := binary.Write(&buf, binary.BigEndian, int32(0))
	if err != nil {
		panic(err)
	}
	var typ int32 = 0
	if !isMessage {
		typ = 1
	}
	err = binary.Write(&buf, binary.BigEndian, typ)
	if err != nil {
		panic(err)
	}
	err = binary.Write(&buf, binary.BigEndian, index)
	if err != nil {
		panic(err)
	}
	return buf.Bytes()
}

func makeMessage(cfg MessageType, jwkKey *jose.JsonWebKey) (message Message, err error) {
	payload := SerializeJSON(cfg)

	signer, err := jose.NewSigner(jose.RS256, jwkKey)
	if err != nil {
		return
	}

	signer.SetNonceSource(NewFixedSizeB64(256))

	object, err := signer.Sign(payload)
	if err != nil {
		return
	}

	// We can't access the protected header without serializing - ugly workaround.
	serialized := object.FullSerialize()

	var parsed struct {
		Protected string `json:"protected"`
		Payload   string `json:"payload"`
		Signature string `json:"signature"`
	}
	err = json.Unmarshal([]byte(serialized), &parsed)
	if err != nil {
		return
	}

	signature := MustBase64URLDecode(parsed.Signature)
	message = Message{
		[]byte("{\"alg\":\"RS256\"}"),
		MustBase64URLDecode(parsed.Protected),
		MustBase64URLDecode(parsed.Payload),
		signature,
	}
	return
}

func encrypt(message *Message, key []byte, index int32) (ct []byte, err error) {
	nonce := CreateNonce(false, index)

	plaintext, err := proto.Marshal(message)
	if err != nil {
		return
	}

	aes, err := aes.NewCipher(key)
	if err != nil {
		return
	}

	aesgcm, err := cipher.NewGCM(aes)
	if err != nil {
		return
	}

	var ad []byte
	ct = aesgcm.Seal(nil, nonce, plaintext, ad)

	return
}

func CreateSignAndEncryptInitial(cfg MessageType, jwkKey *jose.JsonWebKey, root *storage.RecordTopic) (record *Record, key []byte, err error) {
	cfg.Initialize(root, nil, nil)

	message, err := makeMessage(cfg, jwkKey)

	key = sha256.New().Sum(message.Signature)[0:16]
	ct, err := encrypt(&message, key, 0)

	record = &Record{
		Index: 0,
		Type: cfg.Header().Resource,
		EncryptedMessage: ct,
	}

	return
}

func CreateSignAndEncrypt(cfg MessageType, jwkKey *jose.JsonWebKey, root *storage.RecordTopic, parent *Message, topic storage.RecordTopic, key []byte) (record *Record, err error) {
	cfg.Initialize(root, &topic, parent)

	message, err := makeMessage(cfg, jwkKey)

	ct, err := encrypt(&message, key, cfg.Header().Index)

	record = &Record{
		Index: cfg.Header().Index,
		Type: cfg.Header().Resource,
		EncryptedMessage: ct,
	}

	return
}

func DecryptMessage(encryptedMessage []byte, index int32, key []byte) (message Message, err error) {
	nonce := CreateNonce(false, index)

	aes, err := aes.NewCipher(key)
	if err != nil {
		return
	}

	aesgcm, err := cipher.NewGCM(aes)
	if err != nil {
		return
	}

	var ad []byte
	plain, err := aesgcm.Open(nil, nonce, encryptedMessage, ad)
	if err != nil {
		return
	}

	err = proto.Unmarshal(plain, &message)
	return
}


func GetTopic(root *Record, key []byte) storage.RecordTopicAndKey {
	return storage.RecordTopicAndKey{
		sha256.Sum256(root.EncryptedMessage),
		key,
	}
}