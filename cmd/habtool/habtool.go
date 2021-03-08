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
	"os"

	"github.com/f-secure-foundry/crucible/hab"
)

type Config struct {
	key1  string
	key2  string
	key3  string
	key4  string
	hash  string
	table string
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

func init() {
	conf = &Config{}

	log.SetFlags(0)
	log.SetOutput(os.Stdout)

	flag.Usage = func() {
		tags := ""

		if Revision != "" && Build != "" {
			tags = fmt.Sprintf("%s (%s)", Revision, Build)
		}

		log.Printf("habtool - NXP HABv4 Secure Boot tool %s", tags)
		log.Printf("Usage: habtool [options]\n")
		flag.PrintDefaults()
	}

	flag.StringVar(&conf.key1, "1", "", "SRK public key 1 in PEM format")
	flag.StringVar(&conf.key2, "2", "", "SRK public key 2 in PEM format")
	flag.StringVar(&conf.key3, "3", "", "SRK public key 3 in PEM format")
	flag.StringVar(&conf.key4, "4", "", "SRK public key 4 in PEM format")
	flag.StringVar(&conf.hash, "o", "", "Write SRK table hash to file")
	flag.StringVar(&conf.table, "O", "", "Write SRK table to file")
}

func main() {
	var err error

	flag.Parse()

	log.Println(warning)

	switch {
	case len(conf.hash) > 0 || len(conf.table) > 0:
		var f *os.File
		table, _ := hab.NewSRKTable(nil)

		for _, keyPath := range []string{conf.key1, conf.key2, conf.key3, conf.key4} {
			var key []byte

			if len(keyPath) > 0 {
				if key, err = os.ReadFile(keyPath); err != nil {
					break
				}

				if err = table.AddKey(key); err != nil {
					break
				}
			}
		}

		if err != nil {
			break
		}

		if len(conf.hash) > 0 {
			f, err = os.OpenFile(conf.hash, os.O_WRONLY|os.O_CREATE|os.O_EXCL|os.O_TRUNC, 0600)

			if err != nil {
				break
			}

			hash := table.Hash()
			log.Printf("SRK table hash: %x", hash)

			if _, err = f.Write(hash[:]); err != nil {
				break
			}

			log.Printf("SRK table hash written to %s", conf.hash)
		}

		if len(conf.table) > 0 {
			f, err = os.OpenFile(conf.table, os.O_WRONLY|os.O_CREATE|os.O_EXCL|os.O_TRUNC, 0600)

			if err != nil {
				break
			}

			if _, err = f.Write(table.Bytes()); err != nil {
				break
			}

			log.Printf("SRK table written to %s", conf.table)
		}
	default:
		flag.PrintDefaults()
	}

	if err != nil {
		log.Fatalf("error: %v", err)
	}
}
