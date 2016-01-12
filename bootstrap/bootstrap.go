package bootstrap
import (
	"crypto/rand"
	"crypto/rsa"
	"log"
	"time"
	"math/big"
	"crypto/x509"
	"crypto/x509/pkix"
	"encoding/pem"
	"os"
	"github.com/boivie/sec/storage"
	"github.com/boivie/sec/utils"
	"github.com/boivie/sec/messages"
)

const rsaBits = 2048
const validFor = 30*365*24*time.Hour

func generateRootKeyAndCert(id string) (*rsa.PrivateKey, []byte) {
	priv, err := rsa.GenerateKey(rand.Reader, rsaBits)
	if err != nil {
		log.Fatalf("failed to generate private key: %s", err)
	}
	notBefore := time.Now()
	notAfter := notBefore.Add(validFor)

	serialNumberLimit := new(big.Int).Lsh(big.NewInt(1), 128)
	serialNumber, err := rand.Int(rand.Reader, serialNumberLimit)
	if err != nil {
		log.Fatalf("failed to generate serial number: %s", err)
	}

	template := x509.Certificate{
		SerialNumber: serialNumber,
		Subject: pkix.Name{
			CommonName: id,
		},
		NotBefore: notBefore,
		NotAfter:  notAfter,

		KeyUsage:              x509.KeyUsageKeyEncipherment | x509.KeyUsageDigitalSignature | x509.KeyUsageCertSign,
		ExtKeyUsage:           []x509.ExtKeyUsage{x509.ExtKeyUsageServerAuth},
		BasicConstraintsValid: true,
		IsCA: true,
	}

	derBytes, err := x509.CreateCertificate(rand.Reader, &template, &template, &priv.PublicKey, priv)
	if err != nil {
		log.Fatalf("Failed to create certificate: %s", err)
	}

	certOut, err := os.Create("cert.pem")
	if err != nil {
		log.Fatalf("failed to open cert.pem for writing: %s", err)
	}
	pemBytes := pem.EncodeToMemory(&pem.Block{Type: "CERTIFICATE", Bytes: derBytes})
	certOut.Write(pemBytes)
	certOut.Close()
	log.Print("written cert.pem\n")

	keyOut, err := os.OpenFile("key.pem", os.O_WRONLY|os.O_CREATE|os.O_TRUNC, 0600)
	if err != nil {
		log.Fatalf("failed to open key.pem for writing: %s", err)
	}
	pem.Encode(keyOut, &pem.Block{Type: "RSA PRIVATE KEY", Bytes: x509.MarshalPKCS1PrivateKey(priv)})
	keyOut.Close()
	log.Print("written key.pem\n")

	return priv, pemBytes
}

func Bootstrap(stor storage.MessageStorage) {
	ts := utils.NowStr()
	offer := messages.Serialize(messages.CertSigningRequest{Timestamp: ts})
	id := utils.RecordChecksum(offer)
	_, pemBytes := generateRootKeyAndCert(id)

	// Add root certificate
	stor.Add(storage.Message{id, 1, offer})
	cert := messages.Serialize(messages.SignedCert{PEM: string(pemBytes)})
	stor.Add(storage.Record{id, 2, cert})
	log.Printf("Created root cert %s\n", id)

	// Add root config
	rootConfig := messages.Serialize(messages.RootConfig{RootCert: id})
	configId := utils.RecordChecksum(rootConfig)
	stor.Add(storage.Record{configId, 1, rootConfig})
	log.Printf("Created root config %s\n", configId)
}