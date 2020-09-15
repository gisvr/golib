package main

import (
	"github.com/gisvr/golib/log"
	"crypto/ecdsa"
	"crypto/elliptic"
	"crypto/rand"
	"crypto/x509"
	"crypto/x509/pkix"
	"encoding/pem"
	"math/big"
	"os"
	"time"
)

func main() {
	priv, err := ecdsa.GenerateKey(elliptic.P256(), rand.Reader)
	if err != nil {
		log.Error("%+v", err)
		return
	}
	notBefore := time.Now()
	notAfter := notBefore.Add(time.Hour * 24 * 3650)
	serialNumberLimit := new(big.Int).Lsh(big.NewInt(1), 128)
	serialNumber, err := rand.Int(rand.Reader, serialNumberLimit)
	template := x509.Certificate{
		SerialNumber: serialNumber,
		Subject: pkix.Name{
			Country:            []string{"china"},
			Province:           []string{"beijing"},
			StreetAddress:      []string{},
			Organization:       []string{"bibox"},
			OrganizationalUnit: []string{"bibox"},
			CommonName:         "wallet",
		},
		NotBefore: notBefore,
		NotAfter:  notAfter,

		KeyUsage: x509.KeyUsageKeyEncipherment | x509.KeyUsageDigitalSignature,
		ExtKeyUsage: []x509.ExtKeyUsage{
			x509.ExtKeyUsageServerAuth,
			x509.ExtKeyUsageClientAuth,
		},
	}
	der, err := x509.CreateCertificate(rand.Reader, &template, &template, &priv.PublicKey, priv)
	pemFile, err := os.Create("aa.pem")
	if err != nil {
		log.Fatal("%+v", err)
	}
	pem.Encode(pemFile, &pem.Block{Type: "CERTIFICATE", Bytes: der})
	pemFile.Close()
	pemFile, err = os.Create("aa.key")
	if err != nil {
		log.Fatal("%+v", err)
	}
	privByte, err := x509.MarshalECPrivateKey(priv)
	if err != nil {
		log.Fatal("%+v", err)
	}
	pem.Encode(pemFile, &pem.Block{Type: "EC PRIVATE KEY", Bytes: privByte})
	pemFile.Close()
}
