package main

import (
	"context"
	"crypto"
	"crypto/x509"
	"encoding/pem"
	"errors"
	"fmt"
	"io"
	"regexp"
	"time"

	kms "cloud.google.com/go/kms/apiv1"
	"cloud.google.com/go/kms/apiv1/kmspb"
	privateca "cloud.google.com/go/security/privateca/apiv1"
	"cloud.google.com/go/security/privateca/apiv1/privatecapb"
)

var GCPTimeout = 30 * time.Second

// Matchers for GCP resource IDs
var (
	getCertificatePathMatcher          = regexp.MustCompile("projects/[^/]+/locations/[^/]+/caPools/[^/]+/certificates/.+")
	getCertificateAuthorityPathMatcher = regexp.MustCompile("projects/[^/]+/locations/[^/]+/caPools/[^/]+/certificateAuthorities/.+")
)

// certFromGCP fetches the specified certificate from GCP.
// r must be either a GCP Certificate or CertificateAuthority resource ID.
func certFromGCP(ctx context.Context, r string) (*x509.Certificate, error) {
	c, err := privateca.NewCertificateAuthorityClient(ctx)
	if err != nil {
		return nil, err
	}
	defer c.Close()

	var p string
	if getCertificatePathMatcher.Match([]byte(r)) {
		req := &privatecapb.GetCertificateRequest{Name: r}
		resp, err := c.GetCertificate(ctx, req)
		if err != nil {
			return nil, err
		}
		p = resp.GetPemCertificate()
	} else if getCertificateAuthorityPathMatcher.Match([]byte(r)) {
		req := &privatecapb.GetCertificateAuthorityRequest{Name: r}
		ca, err := c.GetCertificateAuthority(ctx, req)
		if err != nil {
			return nil, err
		}
		p = ca.GetPemCaCertificates()[0]
	} else {
		return nil, fmt.Errorf("unknown resource ID type %q", r)
	}

	der, _ := pem.Decode([]byte(p))
	if der == nil {
		return nil, fmt.Errorf("invalid PEM at %q", r)
	}
	return x509.ParseCertificate(der.Bytes)
}

func signerFromGCP(ctx context.Context, f string) (crypto.Signer, error) {
	c, err := kms.NewKeyManagementClient(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to create KeyManagementClient: %v", err)
	}
	go func() {
		<-ctx.Done()
		c.Close()
	}()

	resp, err := c.GetPublicKey(ctx, &kmspb.GetPublicKeyRequest{Name: f})
	if err != nil {
		return nil, err
	}
	der, _ := pem.Decode([]byte(resp.GetPem()))
	if der == nil {
		return nil, fmt.Errorf("invalid PEM for public key at %q", f)
	}
	pubK, err := x509.ParsePKIXPublicKey(der.Bytes)
	if err != nil {
		return nil, err
	}

	s := &gcpSigner{
		ctx:     ctx,
		client:  c,
		keyName: f,
		pubK:    pubK,
	}
	return s, nil
}

type gcpSigner struct {
	// ctx controls the lifecycle for the signer as a whole.
	// It's unusual to have this here, but the interface for crypto.Signer doesn't
	// allow for passing it in per-operation.
	ctx     context.Context
	client  *kms.KeyManagementClient
	pubK    crypto.PublicKey
	keyName string
}

func (s *gcpSigner) Sign(_ io.Reader, digest []byte, opts crypto.SignerOpts) (signature []byte, err error) {
	if opts.HashFunc() != crypto.SHA256 {
		return nil, errors.New("only SHA256 digest is supported")
	}
	req := &kmspb.AsymmetricSignRequest{
		Name: s.keyName,
		Digest: &kmspb.Digest{
			Digest: &kmspb.Digest_Sha256{
				Sha256: digest,
			},
		},
	}
	ctx, cancel := context.WithTimeout(s.ctx, GCPTimeout)
	defer cancel()

	resp, err := s.client.AsymmetricSign(ctx, req)
	if err != nil {
		return nil, err
	}
	return resp.GetSignature(), nil
}

func (s *gcpSigner) Public() crypto.PublicKey {
	return s.pubK
}
