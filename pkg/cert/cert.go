package cert

import (
	"bytes"
	cryptorand "crypto/rand"
	"crypto/rsa"
	"crypto/x509"
	"crypto/x509/pkix"
	"encoding/pem"
	"fmt"
	"log"
	"math/big"
	"os"
	"time"
)

type Certificate struct {
	Cert       *bytes.Buffer
	CA         *bytes.Buffer
	PrivateKey *bytes.Buffer
}

/*
Store a self-signed certificate into Certificate.
*/
func (c *Certificate) GenerateSelfSigned() error {
	// CA config
	ca := &x509.Certificate{
		SerialNumber: big.NewInt(2020),
		Subject: pkix.Name{
			Organization: []string{"Kubernetes"},
		},
		NotBefore:             time.Now(),
		NotAfter:              time.Now().AddDate(1, 0, 0),
		IsCA:                  true,
		ExtKeyUsage:           []x509.ExtKeyUsage{x509.ExtKeyUsageClientAuth, x509.ExtKeyUsageServerAuth},
		KeyUsage:              x509.KeyUsageDigitalSignature | x509.KeyUsageCertSign,
		BasicConstraintsValid: true,
	}

	// CA private key
	caPrivKey, err := rsa.GenerateKey(cryptorand.Reader, 4096)
	if err != nil {
		fmt.Println(err)
		return err
	}

	// Self signed CA certificate
	caBytes, err := x509.CreateCertificate(cryptorand.Reader, ca, ca, &caPrivKey.PublicKey, caPrivKey)
	if err != nil {
		fmt.Println(err)
		return err

	}

	// PEM encode CA cert
	c.CA = new(bytes.Buffer)
	_ = pem.Encode(c.CA, &pem.Block{
		Type:  "CERTIFICATE",
		Bytes: caBytes,
	})
	var (
		webhookNamespace, _ = os.LookupEnv("WEBHOOK_NAMESPACE")
		// validationCfgName, _ = os.LookupEnv("VALIDATE_CONFIG") Not used here in below code
		webhookService, _ = os.LookupEnv("WEBHOOK_SERVICE")
	)
	dnsNames := []string{webhookService,
		webhookService + "." + webhookNamespace, webhookService + "." + webhookNamespace + ".svc", webhookService + "." + webhookNamespace + ".svc.cluster.local"}
	commonName := webhookService + "." + webhookNamespace + ".svc"

	// server cert config
	cert := &x509.Certificate{
		DNSNames:     dnsNames,
		SerialNumber: big.NewInt(1658),
		Subject: pkix.Name{
			CommonName:   commonName,
			Organization: []string{"Kubernetes"},
		},
		NotBefore:    time.Now(),
		NotAfter:     time.Now().AddDate(1, 0, 0),
		SubjectKeyId: []byte{1, 2, 3, 4, 6},
		ExtKeyUsage:  []x509.ExtKeyUsage{x509.ExtKeyUsageClientAuth, x509.ExtKeyUsageServerAuth},
		KeyUsage:     x509.KeyUsageDigitalSignature,
	}

	// server private key
	serverPrivKey, err := rsa.GenerateKey(cryptorand.Reader, 4096)
	if err != nil {
		fmt.Println(err)
		return err

	}

	// sign the server cert
	serverCertBytes, err := x509.CreateCertificate(cryptorand.Reader, cert, ca, &serverPrivKey.PublicKey, caPrivKey)
	if err != nil {
		fmt.Println(err)
		return err

	}

	// PEM encode the  server cert and key
	c.Cert = new(bytes.Buffer)
	_ = pem.Encode(c.Cert, &pem.Block{
		Type:  "CERTIFICATE",
		Bytes: serverCertBytes,
	})

	c.PrivateKey = new(bytes.Buffer)
	_ = pem.Encode(c.PrivateKey, &pem.Block{
		Type:  "RSA PRIVATE KEY",
		Bytes: x509.MarshalPKCS1PrivateKey(serverPrivKey),
	})

	return nil
}

func (c *Certificate) SaveCertToPath(dirPath string) {
	err := writeFile(dirPath+"/tls.crt", c.Cert)
	if err != nil {
		log.Panic(err)
	}

	err = writeFile(dirPath+"/tls.key", c.PrivateKey)
	if err != nil {
		log.Panic(err)
	}
}

func writeFile(filepath string, data *bytes.Buffer) error {
	f, err := os.Create(filepath)
	if err != nil {
		return err
	}
	defer f.Close()

	_, err = f.Write(data.Bytes())
	if err != nil {
		return err
	}
	return nil
}
