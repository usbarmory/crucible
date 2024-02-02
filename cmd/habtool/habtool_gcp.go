package main

import (
	"context"
	"crypto"
	"crypto/rsa"
	"errors"
	"fmt"
	"io"
	"time"

	kms "cloud.google.com/go/kms/apiv1"
	"cloud.google.com/go/kms/apiv1/kmspb"
)

var GCPSignTimeout = 30 * time.Second

func signerFromGCP(f string) (crypto.Signer, error) {
	s := &gcpSigner{
		keyName: f,
	}
	return s, nil
}

type gcpSigner struct {
	keyName string
}

func (s *gcpSigner) Sign(_ io.Reader, digest []byte, opts crypto.SignerOpts) (signature []byte, err error) {
	if opts.HashFunc() != crypto.SHA256 {
		return nil, errors.New("only SHA256 digest is supported")
	}
	c, err := kms.NewKeyManagementClient(context.Background())
	if err != nil {
		return nil, fmt.Errorf("failed to create KeyManagementClient: %v", err)
	}
	defer c.Close()

	req := &kmspb.AsymmetricSignRequest{
		Name: s.keyName,
		Digest: &kmspb.Digest{
			Digest: &kmspb.Digest_Sha256{
				Sha256: digest,
			},
		},
	}
	ctx, cancel := context.WithTimeout(context.Background(), GCPSignTimeout)
	defer cancel()

	resp, err := c.AsymmetricSign(ctx, req)
	if err != nil {
		return nil, err
	}
	return resp.GetSignature(), nil
}

func (s *gcpSigner) Public() crypto.PublicKey {
	// TODO(al): fetch and fill this in
	return &rsa.PublicKey{}
}
