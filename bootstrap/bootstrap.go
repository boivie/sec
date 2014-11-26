package bootstrap

import (
	"crypto/rand"
	"crypto/rsa"
	"crypto/x509"
	"crypto/x509/pkix"
	"encoding/json"
	"encoding/pem"
	"github.com/boivie/gojws"
	"github.com/boivie/sec/common"
	"github.com/boivie/sec/dao"
	"github.com/boivie/sec/store"
	"github.com/boivie/sec/store/dbstore"
	"github.com/boivie/sec/utils"
	"github.com/op/go-logging"
	"math/big"
	"os"
	"time"
)

var (
	log = logging.MustGetLogger("sec")
)

func generateKeyAndCert(subject pkix.Name) (priv *rsa.PrivateKey, cert *x509.Certificate) {
	log.Info("Generating private key, 2048 bits")
	priv, err := rsa.GenerateKey(rand.Reader, 2048)
	if err != nil {
		log.Fatalf("failed to generate private key: %v", err)
	}
	notBefore := time.Now()
	notAfter := notBefore.Add(30 * 365 * 24 * time.Hour)
	serialNumberLimit := new(big.Int).Lsh(big.NewInt(1), 128)
	serialNumber, err := rand.Int(rand.Reader, serialNumberLimit)
	if err != nil {
		log.Fatalf("failed to generate serial number: %s", err)
	}

	template := x509.Certificate{
		SerialNumber: serialNumber,
		Subject:      subject,
		NotBefore:    notBefore,
		NotAfter:     notAfter,

		KeyUsage:              x509.KeyUsageKeyEncipherment | x509.KeyUsageDigitalSignature | x509.KeyUsageCertSign,
		BasicConstraintsValid: true,
		IsCA: true,
	}

	certDer, err := x509.CreateCertificate(rand.Reader, &template, &template, &priv.PublicKey, priv)

	if err != nil {
		log.Fatalf("Failed to create certificate: %s", err)
	}

	cert, err = x509.ParseCertificate(certDer)

	return
}

func writeCert(cert *x509.Certificate, filename string) {
	certOut, err := os.Create(filename)
	if err != nil {
		log.Fatalf("failed to open %s for writing: %s", filename, err)
	}
	pem.Encode(certOut, &pem.Block{Type: "CERTIFICATE", Bytes: cert.Raw})
	certOut.Close()
	log.Info("written %s\n", filename)
}

func writeKey(priv *rsa.PrivateKey, filename string) {
	der := x509.MarshalPKCS1PrivateKey(priv)

	keyOut, err := os.OpenFile(filename,
		os.O_WRONLY|os.O_CREATE|os.O_TRUNC, 0600)
	if err != nil {
		log.Fatalf("failed to open %s for writing:", filename, err)
	}
	pem.Encode(keyOut, &pem.Block{Type: "RSA PRIVATE KEY", Bytes: der})
	keyOut.Close()
	log.Info("written %s\n", filename)
}

type TemplateField struct {
	Name string `json:"name"`
}

type TemplateInput struct {
	Name   string          `json:"name"`
	Title  string          `json:"title"`
	Fields []TemplateField `json:"fields"`
}

func signTemplate(tmpl TemplateInput, priv *rsa.PrivateKey, cert *x509.Certificate) (contents string, err error) {
	fingerprint := utils.GetCertFingerprint(cert)
	header := gojws.Header{
		Alg:    gojws.ALG_RS256,
		Typ:    "template",
		X5t256: fingerprint,
	}
	tmplBytes, err := json.Marshal(tmpl)
	if err != nil {
		return
	}

	j, err := gojws.Sign(header, tmplBytes, priv)
	if err == nil {
		contents = string(j)
	}
	return
}

func createOffer(state *common.State, s store.Store, priv *rsa.PrivateKey, cert *x509.Certificate) (requestId int64) {
	secret := utils.GenerateSecret()
	requestId, _ = s.CreateRequest(secret)
	requestStringId := utils.GetStringId(requestId, secret, state.IdCrypto)
	fingerprint := utils.GetCertFingerprint(cert)

	type Header struct {
		Parent string `json:"parent"`
	}

	header := gojws.Header{
		Alg:    gojws.ALG_RS256,
		Typ:    "create",
		X5t256: fingerprint,
	}
	create, _ := json.Marshal(struct {
		Hdr       Header `json:"header"`
		RequestId string `json:"request_id"`
	}{
		Header{},
		requestStringId,
	})
	j1, _ := gojws.Sign(header, create, priv)

	header.Typ = "offer"
	offer, _ := json.Marshal(struct {
		Hdr      Header   `json:"header"`
		Template string   `json:"template"`
		Fields   struct{} `json:"fields"`
	}{
		Header{Parent: utils.GetFingerprint([]byte(j1))},
		"issuer",
		struct{}{},
	})
	j2, _ := gojws.Sign(header, offer, priv)

	update := dao.RequestDao{
		Payload: j1 + "\n" + j2,
		Version: 1,
	}
	s.UpdateRequest(requestId, 0, update)
	log.Info("Created offer at %s", requestStringId)

	return
}

func Bootstrap(state *common.State) {
	s := dbstore.NewDBStore(state)
	webSubj := pkix.Name{CommonName: "Test Web Cert"}
	webPriv, webCert := generateKeyAndCert(webSubj)
	writeKey(webPriv, "web.key")
	writeCert(webCert, "web.crt")
	s.StoreCert(webCert, 0)

	issuerSubj := pkix.Name{CommonName: "Test Issuer"}
	issuerPriv, issuerCert := generateKeyAndCert(issuerSubj)
	writeKey(issuerPriv, "issuer.key")
	writeCert(issuerCert, "issuer.crt")
	s.StoreCert(issuerCert, 0)

	// Add the two templates:
	issuer := TemplateInput{
		Name:   "issuer",
		Title:  "Certificate Issuer",
		Fields: []TemplateField{{"title"}},
	}
	isigned, _ := signTemplate(issuer, webPriv, webCert)
	s.StoreTemplate(issuer.Name, isigned)

	client := TemplateInput{
		Name:   "client",
		Title:  "Client",
		Fields: []TemplateField{{"username"}},
	}
	csigned, _ := signTemplate(client, webPriv, webCert)
	s.StoreTemplate(client.Name, csigned)

	// Create the bootstrap offer
	state.BootstrapRequestId = createOffer(state, s, webPriv, webCert)
	state.WebKey = webPriv
	state.WebCert = webCert
	state.IssueKey = issuerPriv
	state.IssueCert = issuerCert
}
