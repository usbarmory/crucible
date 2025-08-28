// crucible
// One-Time-Programmable (OTP) fusing tool
//
// Copyright (c) The crucible authors. All Rights Reserved.
//
// Use of this source code is governed by the license
// that can be found in the LICENSE file.

package main

import (
	"errors"
	"fmt"
	"log"
	"math/big"
	"strings"

	"github.com/usbarmory/crucible/fusemap"
	"github.com/usbarmory/crucible/otp"
	"github.com/usbarmory/crucible/util"
)

func read(tag string, f *fusemap.FuseMap, name string) (err error) {
	res, addr, off, size, err := otp.ReadNVMEM(conf.device, f, name)

	if err != nil {
		return
	}

	tag = fmt.Sprintf("%s addr:%#x off:%d len:%d", tag, addr, off, size)

	if conf.endianness == "little" {
		res = util.SwitchEndianness(res)
	}

	n := new(big.Int)
	n.SetBytes(res)

	var base string
	var format string
	var value string

	switch conf.base {
	case 2:
		base = "0b"
		format = "%0" + fmt.Sprintf("%d", size) + "b"
		value = fmt.Sprintf(format, n)
	case 10:
		value = fmt.Sprintf("%d", n)
	case 16:
		base = "0x"
		format = "%0" + fmt.Sprintf("%d", (size+3)/4) + "x"
		value = fmt.Sprintf(format, n)
	default:
		return errors.New("internal error, invalid base")
	}

	log.Printf("%s val:%s%s", tag, base, value)

	if conf.syslog {
		fmt.Println(value)
	} else if conf.list {
		if reg, ok := f.Registers[name]; ok {
			if conf.endianness == "little" {
				res = util.SwitchEndianness(res)
			}

			log.Println()
			log.Print(reg.BitMap(res))
		}
	}

	return
}

func blow(tag string, f *fusemap.FuseMap, name string, val string) (err error) {
	base := ""
	n := new(big.Int)

	switch conf.base {
	case 2:
		base = "0b"
	case 10:
	case 16:
		base = "0x"
	default:
		return errors.New("internal error, invalid base")
	}

	val = strings.TrimPrefix(val, base)
	n, ok := n.SetString(val, conf.base)

	if !ok {
		return errors.New("invalid value argument")
	}

	switch conf.endianness {
	case "big":
	case "little":
		n = n.SetBytes(util.SwitchEndianness(n.Bytes()))
	default:
		return errors.New("you must specify a valid endianness")
	}

	if !conf.force {
		log.Print(warning)
		log.Printf("%s reg:%s base:%d val:%s %s-endian\n\n", tag, name, conf.base, val, conf.endianness)

		if !confirm() {
			log.Fatal("you are not ready...")
		}
	}

	res, addr, off, size, err := otp.BlowNVMEM(conf.device, f, name, n.Bytes())

	if err != nil {
		return err
	}

	log.Printf("%s addr:%#x off:%d len:%d val:%s%s res:%#x", tag, addr, off, size, base, val, res)

	if conf.syslog {
		fmt.Printf("%#x\n", res)
	}

	return
}
