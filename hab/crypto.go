// crucible
// One-Time-Programmable (OTP) fusing tool
//
// Copyright (c) F-Secure Corporation
//
// Use of this source code is governed by the license
// that can be found in the LICENSE file.

package hab

import (
	"bytes"
	"crypto/rand"
	"crypto/rsa"
	"crypto/x509"
	"crypto/x509/pkix"
	"encoding/pem"
	"errors"
	"fmt"
	"math/big"
	"time"

	"go.mozilla.org/pkcs7"
)

func padCert(buf []byte) []byte {
	if pad := (4 - (len(buf) % 4)) % 4; pad > 0 {
		buf = append(buf, make([]byte, pad)...)
	}

	return buf
}

func parseCert(certPEMBlock []byte) (*rsa.PublicKey, []byte, error) {
	block, _ := pem.Decode([]byte(certPEMBlock))

	if block == nil {
		return nil, nil, errors.New("failed to parse certificate PEM")
	}

	cert, err := x509.ParseCertificate(block.Bytes)

	if err != nil {
		return nil, nil, fmt.Errorf("failed to parse certificate, %v", err)
	}

	switch pubKey := cert.PublicKey.(type) {
	case *rsa.PublicKey:
		return pubKey, cert.Raw, nil
	default:
		return nil, nil, fmt.Errorf("unexpected public key type %T", pubKey)
	}
}

func parseKey(keyPEMBlock []byte) (*rsa.PrivateKey, error) {
	block, _ := pem.Decode([]byte(keyPEMBlock))

	if block == nil {
		return nil, errors.New("failed to parse certificate PEM")
	}

	key, err := x509.ParsePKCS8PrivateKey(block.Bytes)

	if err == nil {
		switch privKey := key.(type) {
		case *rsa.PrivateKey:
			return privKey, nil
		default:
			return nil, errors.New("failed to parse key")
		}
	}

	return x509.ParsePKCS1PrivateKey(block.Bytes)
}

func sign(buf []byte, certPEMBlock []byte, privKey *rsa.PrivateKey) (sig []byte, err error) {
	block, _ := pem.Decode([]byte(certPEMBlock))

	if block == nil {
		return nil, errors.New("failed to parse certificate PEM")
	}

	cert, err := x509.ParseCertificate(block.Bytes)

	if err != nil {
		return nil, fmt.Errorf("failed to parse certificate, %v", err)
	}

	data, err := pkcs7.NewSignedData(buf)

	if err != nil {
		return nil, fmt.Errorf("cannot initialize signed data: %s", err)
	}

	data.SetDigestAlgorithm(pkcs7.OIDDigestAlgorithmSHA256)
	data.SetEncryptionAlgorithm(pkcs7.OIDEncryptionAlgorithmRSA)

	if err = data.AddSigner(cert, privKey, pkcs7.SignerInfoConfig{}); err != nil {
		return nil, fmt.Errorf("cannot add signer: %s", err)
	}

	data.Detach()

	return data.Finish()
}

// NewCA generates a certificate authority suitable for signing HABv4 CSF/IMG
// certificates.
func NewCA(keyLength int, keyExpiry int) (pemKey []byte, pemCert []byte, err error) {
	serialSize := new(big.Int).Lsh(big.NewInt(1), 160)
	serial, _ := rand.Int(rand.Reader, serialSize)

	ca := x509.Certificate{
		BasicConstraintsValid: true,
		IsCA:                  true,
		SerialNumber:          serial,
		Subject: pkix.Name{
			CommonName: fmt.Sprintf("SRK_sha256_%d", keyLength),
		},
		SignatureAlgorithm: x509.SHA256WithRSA,
		PublicKeyAlgorithm: x509.RSA,
		NotBefore:          time.Now(),
		NotAfter:           time.Now().AddDate(0, 0, keyExpiry),
		KeyUsage:           x509.KeyUsageCertSign,
	}

	key, err := rsa.GenerateKey(rand.Reader, keyLength)

	if err != nil {
		return
	}

	privKey, err := x509.MarshalPKCS8PrivateKey(key)

	if err != nil {
		return
	}

	cert, err := x509.CreateCertificate(rand.Reader, &ca, &ca, &key.PublicKey, key)

	if err != nil {
		return
	}

	keyBuf := new(bytes.Buffer)
	pem.Encode(keyBuf, &pem.Block{Type: "PRIVATE KEY", Bytes: privKey})

	certBuf := new(bytes.Buffer)
	pem.Encode(certBuf, &pem.Block{Type: "CERTIFICATE", Bytes: cert})

	return keyBuf.Bytes(), certBuf.Bytes(), nil
}

// NewCertificate generates a certificate suitable for HABv4 signing, the tag
// string (e.g. "CSF" or "IMG") is used in the certificate Common Name to
// distinguish its role.
func NewCertificate(tag string, keyLength int, keyExpiry int, parent *x509.Certificate, signer *rsa.PrivateKey) (pemKey []byte, pemCert []byte, err error) {
	// serial should be 0
	var serial big.Int

	certificate := x509.Certificate{
		SerialNumber: &serial,
		Subject: pkix.Name{
			CommonName: fmt.Sprintf("%s_sha256_%d", tag, keyLength),
		},
		SignatureAlgorithm: x509.SHA256WithRSA,
		PublicKeyAlgorithm: x509.RSA,
		NotBefore:          time.Now(),
		NotAfter:           time.Now().AddDate(0, 0, keyExpiry),
	}

	key, err := rsa.GenerateKey(rand.Reader, keyLength)

	if err != nil {
		return
	}

	privKey, err := x509.MarshalPKCS8PrivateKey(key)

	if err != nil {
		return
	}

	cert, err := x509.CreateCertificate(rand.Reader, &certificate, parent, &key.PublicKey, signer)

	if err != nil {
		return
	}

	keyBuf := new(bytes.Buffer)
	pem.Encode(keyBuf, &pem.Block{Type: "PRIVATE KEY", Bytes: privKey})

	certBuf := new(bytes.Buffer)
	pem.Encode(certBuf, &pem.Block{Type: "CERTIFICATE", Bytes: cert})

	return keyBuf.Bytes(), certBuf.Bytes(), nil
}
