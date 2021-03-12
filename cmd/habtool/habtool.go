// crucible
// One-Time-Programmable (OTP) fusing tool
//
// Copyright (c) F-Secure Corporation
//
// Use of this source code is governed by the license
// that can be found in the LICENSE file.

package main

import (
	"flag"
	"fmt"
	"log"
	"math/big"
	"os"

	"github.com/f-secure-foundry/crucible/hab"
)

type Config struct {
	input  string
	output string
	table  string

	srk1 string
	srk2 string
	srk3 string
	srk4 string

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
WARNING: enabling secure boot functionality on the USB armory SoC, unlike
similar features on modern PCs, is an irreversible action that permanently
fuses verification keys hashes on the device. This means that any errors in the
process or loss of the signing PKI will result in a bricked device incapable of
executing unsigned code. This is a security feature, not a bug.

This tool is EXPERIMENTAL and it is highly recommended to verify the SRK table
hash, before fusing, by comparing it with the NXP IMX_CST_TOOL 2.2 generated
one.
`

const usage = `Usage: habtool [OPTIONS]
  -h                  Show this help

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

	flag.StringVar(&conf.srk1, "1", "SRK_1_crt.pem", "SRK public key 1 in PEM format")
	flag.StringVar(&conf.srk2, "2", "SRK_2_crt.pem", "SRK public key 2 in PEM format")
	flag.StringVar(&conf.srk3, "3", "SRK_3_crt.pem", "SRK public key 3 in PEM format")
	flag.StringVar(&conf.srk4, "4", "SRK_4_crt.pem", "SRK public key 4 in PEM format")

	flag.StringVar(&conf.csfKey, "A", "CSF_1_key.pem", "CSF private key in PEM format")
	flag.StringVar(&conf.csfCrt, "a", "CSF_1_crt.pem", "CSF public  key in PEM format")
	flag.StringVar(&conf.imgKey, "B", "IMG_1_key.pem", "IMG private key in PEM format")
	flag.StringVar(&conf.imgCrt, "b", "IMG_1_crt.pem", "IMG public  key in PEM format")
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
