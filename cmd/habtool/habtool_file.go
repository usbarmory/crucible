package main

import (
	"crypto"
	"crypto/rsa"
	"crypto/x509"
	"encoding/pem"
	"errors"
	"os"
)

func keyFromFile(f string) (crypto.Signer, error) {
	b, err := os.ReadFile(f)
	if err != nil {
		return nil, err
	}
	derBlock, _ := pem.Decode(b)
	if derBlock == nil {
		return nil, errors.New("failed to parse certificate PEM")
	}

	key, err := x509.ParsePKCS8PrivateKey(derBlock.Bytes)
	if err == nil {
		switch privKey := key.(type) {
		case *rsa.PrivateKey:
			return privKey, nil
		default:
			return nil, errors.New("failed to parse key")
		}
	}

	return x509.ParsePKCS1PrivateKey(derBlock.Bytes)
}
