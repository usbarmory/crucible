// crucible
// One-Time-Programmable (OTP) fusing tool
//
// Copyright (c) F-Secure Corporation
//
// Use of this source code is governed by the license
// that can be found in the LICENSE file.

package main

import (
	"crypto/rsa"
	"crypto/x509"
	"encoding/pem"
	"errors"
	"flag"
	"fmt"
	"log"
	"math/big"
	"os"

	"github.com/usbarmory/crucible/hab"
)

type Config struct {
	input  string
	output string
	table  string

	srk1 string
	srk2 string
	srk3 string
	srk4 string

	srkKey string
	srkCrt string
	csfKey string
	csfCrt string
	imgKey string
	imgCrt string

	index  int
	engine string
	sdp    bool
	dcd    string
}

// build information, initialized at compile time (see Makefile)
var Revision string
var Build string

var conf *Config

const warning = `
████████████████████████████████████████████████████████████████████████████████

                                **  WARNING  **

Enabling NXP HABv4 secure boot is an irreversible action that permanently fuses
verification keys hashes on the device.

Any errors in the process or loss of the signing PKI will result in a bricked
device incapable of executing unsigned code. This is a security feature, not a
bug.

The use of this tool is therefore **at your own risk**.

████████████████████████████████████████████████████████████████████████████████
`

const usage = `Usage: habtool [OPTIONS]
  -h                  Show this help

SRK CA creation options:
  -C <output path>    SRK private key in PEM format
  -c <output path>    SRK public  key in PEM format

CSF/IMG certificates creation options:
  -C <input path>     SRK private key in PEM format
  -c <input path>     SRK public  key in PEM format

  -A <output path>    CSF private key in PEM format
  -a <output path>    CSF public  key in PEM format
  -B <output path>    IMG private key in PEM format
  -b <output path>    IMG public  key in PEM format

SRK table creation options:
  -1 <input path>     SRK public key 1 in PEM format
  -2 <input path>     SRK public key 2 in PEM format
  -3 <input path>     SRK public key 3 in PEM format
  -4 <input path>     SRK public key 4 in PEM format

  -o <output path>    Write SRK table hash to file
  -t <output path>    Write SRK table to file

Executable signing options:
  -A <input path>     CSF private key in PEM format
  -a <input path>     CSF public  key in PEM format
  -B <input path>     IMG private key in PEM format
  -b <input path>     IMG public  key in PEM format
  -t <input path>     Read SRK table from file
  -x <1-4>            Index for SRK key
  -e <id>             Crypto engine (e.g. 0x1b for HAB_ENG_DCP)
  -i <input path>     Image file w/ IVT header (e.g. boot.imx)

  -o <output path>    Write CSF to file

  -s                  Serial download mode
  -S <address>        Serial download DCD OCRAM address
                      (depends on mfg tool, default: 0x00910000)
`

func init() {
	conf = &Config{}

	log.SetFlags(0)
	log.SetOutput(os.Stdout)

	flag.Usage = func() {
		tags := ""

		if Revision != "" && Build != "" {
			tags = fmt.Sprintf("%s (%s)", Revision, Build)
		}

		log.Printf("habtool - NXP HABv4 Secure Boot utility %s", tags)
		fmt.Println(usage)
	}

	flag.StringVar(&conf.input, "i", "", "Image file w/ IVT header (e.g. boot.imx)")
	flag.StringVar(&conf.output, "o", "", "output")
	flag.StringVar(&conf.table, "t", "SRK_1_2_3_4_table.bin", "SRK table")

	flag.StringVar(&conf.srk1, "1", "", "SRK public key 1 in PEM format")
	flag.StringVar(&conf.srk2, "2", "", "SRK public key 2 in PEM format")
	flag.StringVar(&conf.srk3, "3", "", "SRK public key 3 in PEM format")
	flag.StringVar(&conf.srk4, "4", "", "SRK public key 4 in PEM format")

	flag.StringVar(&conf.srkKey, "C", "", "SRK private key in PEM format")
	flag.StringVar(&conf.srkCrt, "c", "", "SRK public  key in PEM format")
	flag.StringVar(&conf.csfKey, "A", "", "CSF private key in PEM format")
	flag.StringVar(&conf.csfCrt, "a", "", "CSF public  key in PEM format")
	flag.StringVar(&conf.imgKey, "B", "", "IMG private key in PEM format")
	flag.StringVar(&conf.imgCrt, "b", "", "IMG public  key in PEM format")

	flag.IntVar(&conf.index, "x", 1, "Index for SRK key")
	flag.StringVar(&conf.engine, "e", "0xff", "Crypto engine (e.g. 0x1b for HAB_ENG_DCP)")

	flag.BoolVar(&conf.sdp, "s", false, "Serial download mode")
	flag.StringVar(&conf.dcd, "S", "0x00910000", "Serial download DCD OCRAM address")
}

func main() {
	var err error

	flag.Parse()

	log.Println(warning)

	switch {
	case len(conf.srkKey) > 0 && len(conf.srkCrt) > 0 &&
		len(conf.csfKey) > 0 && len(conf.csfCrt) > 0 &&
		len(conf.imgKey) > 0 && len(conf.imgCrt) > 0:
		err = genCerts()
	case len(conf.srkKey) > 0 && len(conf.srkCrt) > 0 &&
		len(conf.csfKey) == 0 && len(conf.csfCrt) == 0 &&
		len(conf.imgKey) == 0 && len(conf.imgCrt) == 0:
		err = genCA()
	case len(conf.table) > 0 && len(conf.input) > 0 && len(conf.output) > 0:
		err = sign()
	case len(conf.table) > 0 && len(conf.output) > 0:
		err = genSRKTable()
	default:
		fmt.Println(usage)
	}

	if err != nil {
		log.Fatalf("error: %v", err)
	}
}

func saveCert(tag string, keyPath string, keyPEMBlock []byte, certPath string, certPEMBlock []byte) (err error) {
	var keyFile, certFile *os.File

	if keyFile, err = os.OpenFile(keyPath, os.O_WRONLY|os.O_CREATE|os.O_EXCL|os.O_TRUNC, 0600); err != nil {
		return
	}

	if certFile, err = os.OpenFile(certPath, os.O_WRONLY|os.O_CREATE|os.O_EXCL|os.O_TRUNC, 0600); err != nil {
		return
	}

	if _, err = keyFile.Write(keyPEMBlock); err != nil {
		return
	}

	log.Printf("%s private key written to %s", tag, keyPath)

	if _, err = certFile.Write(certPEMBlock); err != nil {
		return
	}

	log.Printf("%s public  key written to %s", tag, certPath)

	return
}

func genCerts() (err error) {
	var signingKey *rsa.PrivateKey

	SRKKeyPEMBlock, err := os.ReadFile(conf.srkKey)

	if err != nil {
		return
	}

	SRKCertPEMBlock, err := os.ReadFile(conf.srkCrt)

	if err != nil {
		return
	}

	caKey, _ := pem.Decode(SRKKeyPEMBlock)

	if caKey == nil {
		return errors.New("failed to parse SRK key PEM")
	}

	caCert, _ := pem.Decode(SRKCertPEMBlock)

	if caCert == nil {
		return errors.New("failed to parse SRK certificate PEM")
	}

	ca, err := x509.ParseCertificate(caCert.Bytes)

	if err != nil {
		return
	}

	caPriv, err := x509.ParsePKCS8PrivateKey(caKey.Bytes)

	if err != nil {
		return
	}

	switch k := caPriv.(type) {
	case *rsa.PrivateKey:
		signingKey = k
	default:
		return errors.New("failed to parse SRK key")
	}

	log.Printf("generating and signing CSF keypair")
	CSFKeyPEMBlock, CSFCertPEMBlock, err := hab.NewCertificate("CSF", hab.DEFAULT_KEY_LENGTH, hab.DEFAULT_KEY_EXPIRY, ca, signingKey)

	if err != nil {
		return
	}

	log.Printf("generating and signing IMG keypair")
	IMGKeyPEMBlock, IMGCertPEMBlock, err := hab.NewCertificate("IMG", hab.DEFAULT_KEY_LENGTH, hab.DEFAULT_KEY_EXPIRY, ca, signingKey)

	if err != nil {
		return
	}

	if err = saveCert("CSF", conf.csfKey, CSFKeyPEMBlock, conf.csfCrt, CSFCertPEMBlock); err != nil {
		return
	}

	if err = saveCert("IMG", conf.imgKey, IMGKeyPEMBlock, conf.imgCrt, IMGCertPEMBlock); err != nil {
		return
	}

	return
}

func genCA() (err error) {
	log.Printf("generating SRK certification authority")
	SRKKeyPEMBlock, SRKCertPEMBlock, err := hab.NewCA(hab.DEFAULT_KEY_LENGTH, hab.DEFAULT_KEY_EXPIRY)

	if err != nil {
		return
	}

	return saveCert("SRK", conf.srkKey, SRKKeyPEMBlock, conf.srkCrt, SRKCertPEMBlock)
}

func sign() (err error) {
	var f *os.File

	opts := hab.SignOptions{
		Index: conf.index,
		SDP:   conf.sdp,
	}

	if opts.CSFKeyPEMBlock, err = os.ReadFile(conf.csfKey); err != nil {
		return
	}

	if opts.CSFCertPEMBlock, err = os.ReadFile(conf.csfCrt); err != nil {
		return
	}

	if opts.IMGKeyPEMBlock, err = os.ReadFile(conf.imgKey); err != nil {
		return
	}

	if opts.IMGCertPEMBlock, err = os.ReadFile(conf.imgCrt); err != nil {
		return
	}

	if opts.Table, err = os.ReadFile(conf.table); err != nil {
		return
	}

	engine := new(big.Int)
	engine.SetString(conf.engine, 0)
	opts.Engine = int(engine.Int64())

	dcd := new(big.Int)
	dcd.SetString(conf.dcd, 0)
	opts.DCD = uint32(dcd.Int64())

	input, err := os.ReadFile(conf.input)

	if err != nil {
		return
	}

	log.Printf("generating signatures for %s", conf.input)
	output, err := hab.Sign(input, opts)

	if err != nil {
		return
	}

	if f, err = os.OpenFile(conf.output, os.O_WRONLY|os.O_CREATE|os.O_EXCL|os.O_TRUNC, 0600); err != nil {
		return
	}

	if _, err = f.Write(output); err != nil {
		return
	}

	log.Printf("CSF file written to %s", conf.output)

	return
}

func genSRKTable() (err error) {
	var f *os.File

	table, _ := hab.NewSRKTable(nil)

	for _, keyPath := range []string{conf.srk1, conf.srk2, conf.srk3, conf.srk4} {
		var key []byte

		if len(keyPath) > 0 {
			if key, err = os.ReadFile(keyPath); err != nil {
				return err
			}

			if err = table.AddKey(key); err != nil {
				return err
			}
		}
	}

	if f, err = os.OpenFile(conf.output, os.O_WRONLY|os.O_CREATE|os.O_EXCL|os.O_TRUNC, 0600); err != nil {
		return
	}

	hash := table.Hash()
	log.Printf("SRK table hash: %x", hash)

	if _, err = f.Write(hash[:]); err != nil {
		return
	}

	log.Printf("SRK table hash written to %s", conf.output)

	if f, err = os.OpenFile(conf.table, os.O_WRONLY|os.O_CREATE|os.O_EXCL|os.O_TRUNC, 0600); err != nil {
		return
	}

	if _, err = f.Write(table.Bytes()); err != nil {
		return
	}

	log.Printf("SRK table written to %s", conf.table)

	return
}
