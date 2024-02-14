package main

import (
	"context"
	"crypto"
	"crypto/rsa"
	"crypto/x509"
	"encoding/pem"
	"errors"
	"fmt"
	"os"
)

func certFromFile(_ context.Context, path string) (*x509.Certificate, error) {
	b, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}
	derBlock, _ := pem.Decode(b)
	if derBlock == nil {
		return nil, fmt.Errorf("invalid PEM in %q", path)
	}

	return x509.ParseCertificate(derBlock.Bytes)
}

func signerFromFile(_ context.Context, f string) (crypto.Signer, error) {
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
