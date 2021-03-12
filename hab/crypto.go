// crucible
// One-Time-Programmable (OTP) fusing tool
//
// Copyright (c) F-Secure Corporation
//
// Use of this source code is governed by the license
// that can be found in the LICENSE file.

// Package hab provides support fuctions for NXP HABv4 Secure Boot provisioning
// and executable signing.
package hab

import (
	"crypto/rsa"
	"crypto/x509"
	"encoding/pem"
	"errors"
	"fmt"

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

	privKey, err := x509.ParsePKCS1PrivateKey(block.Bytes)

	if err != nil {
		return nil, fmt.Errorf("failed to parse certificate, %v", err)
	}

	return privKey, nil
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
