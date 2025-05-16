// crucible
// One-Time-Programmable (OTP) fusing tool
//
// Copyright (c) WithSecure Corporation
//
// Use of this source code is governed by the license
// that can be found in the LICENSE file.

package main

import (
	"bufio"
	"embed"
	"errors"
	"flag"
	"fmt"
	"io/fs"
	"log"
	"log/syslog"
	"os"

	"github.com/usbarmory/crucible/fusemap"
)

type Config struct {
	force      bool
	list       bool
	syslog     bool
	base       int
	endianness string
	device     string
	fusemaps   string
	fusemap    string
	processor  string
	reference  string

	fusemapDir fs.FS
}

// Bundled fusemaps
//
//go:embed fusemaps
var fusemaps embed.FS

// build information, initialized at compile time (see Makefile)
var Revision string
var Build string

var conf *Config

const splash = `
 ▄████▄   ██▀███   █    ██  ▄████▄   ██▓ ▄▄▄▄    ██▓    ▓█████
▒██▀ ▀█  ▓██ ▒ ██▒ ██  ▓██▒▒██▀ ▀█  ▓██▒▓█████▄ ▓██▒    ▓█   ▀
▒▓█    ▄ ▓██ ░▄█ ▒▓██  ▒██░▒▓█    ▄ ▒██▒▒██▒ ▄██▒██░    ▒███
▒▓▓▄ ▄██▒▒██▀▀█▄  ▓▓█  ░██░▒▓▓▄ ▄██▒░██░▒██░█▀  ▒██░    ▒▓█  ▄
▒ ▓███▀ ░░██▓ ▒██▒▒▒█████▓ ▒ ▓███▀ ░░██░░▓█  ▀█▓░██████▒░▒████▒
░ ░▒ ▒  ░░ ▒▓ ░▒▓░░▒▓▒ ▒ ▒ ░ ░▒ ▒  ░░▓  ░▒▓███▀▒░ ▒░▓  ░░░ ▒░ ░
  ░  ▒     ░▒ ░ ▒░░░▒░ ░ ░   ░  ▒    ▒ ░▒░▒   ░ ░ ░ ▒  ░ ░ ░  ░
░          ░░   ░  ░░░ ░ ░ ░         ▒ ░ ░    ░   ░ ░      ░
░ ░         ░        ░     ░ ░       ░   ░          ░  ░   ░  ░
░                          ░                  ░

                  Where SoCs meet their fate.
`

const warning = `
████████████████████████████████████████████████████████████████████████████████

                                **  WARNING  **

Fusing SoC OTPs is an **irreversible** action that permanently fuses values on
the device. This means that any errors in the process, or lost fused data such
as cryptographic key material, might result in a **bricked** device.

The use of this tool is therefore **at your own risk**.

████████████████████████████████████████████████████████████████████████████████
`

func init() {
	conf = &Config{}

	log.SetFlags(0)
	log.SetOutput(os.Stdout)

	flag.Usage = func() {
		if conf.syslog {
			return
		}

		tags := ""

		if Revision != "" && Build != "" {
			tags = fmt.Sprintf("%s (%s)", Revision, Build)
		}

		log.Printf("crucible - One-Time-Programmable (OTP) fusing tool %s", tags)
		log.Print(splash)
		log.Printf("Usage: crucible [options] [read|blow] [fuse/register name] [value]\n")
		flag.PrintDefaults()
	}

	flag.BoolVar(&conf.force, "Y", false, "do not prompt for confirmation (DANGEROUS)")
	flag.BoolVar(&conf.list, "l", false, "list fusemaps\nvisualize fusemap      (with -m and -r)\nvisualize read value   (with read operation on a register)\nvisualize read fusemap (with read operation and no register)")
	flag.BoolVar(&conf.syslog, "s", false, "use syslog, print only result value to stdout")
	flag.IntVar(&conf.base, "b", 0, "value base/format (2,10,16)")
	flag.StringVar(&conf.endianness, "e", "", "value endianness (big,little)")
	flag.StringVar(&conf.device, "n", "/sys/bus/nvmem/devices/imx-ocotp0/nvmem", "NVMEM device")
	flag.StringVar(&conf.fusemaps, "f", "", "reference fusemap directory")
	flag.StringVar(&conf.fusemap, "i", "", "vendor fusemap file")
	flag.StringVar(&conf.processor, "m", "", "processor model")
	flag.StringVar(&conf.reference, "r", "", "reference manual revision")

	flag.Parse()
}

func confirm() bool {
	reader := bufio.NewReader(os.Stdin)
	log.Print("Would you really like to blow this fuse? Type YES all uppercase to confirm: ")
	text, _ := reader.ReadString('\n')

	return text == "YES\n"
}

func checkArguments() error {
	switch conf.base {
	case 2, 10, 16:
	default:
		return errors.New("you must specify a valid base format")
	}

	if conf.device == "" {
		return errors.New("you must specify the target NVMEM device")
	}

	if conf.processor == "" {
		return errors.New("you must specify a processor model")
	}

	if conf.reference == "" {
		return errors.New("you must specify a reference manual revision")
	}

	if len(flag.Args()) < 2 {
		return errors.New("missing arguments")
	}

	return nil
}

func op(f *fusemap.FuseMap) {
	if err := checkArguments(); err != nil {
		flag.Usage()
		log.Fatalf("error: %v", err)
	}

	stat, err := os.Stat(conf.device)

	if err != nil || stat.IsDir() {
		log.Fatalf("error: could not open NVMEM device %s", conf.device)
	}

	op := flag.Arg(0)
	name := flag.Arg(1)
	tag := fmt.Sprintf("soc:%s ref:%s otp:%s op:%s", conf.processor, conf.reference, name, op)

	switch op {
	case "read":
		err = read(tag, f, name)
	case "blow":
		if len(flag.Args()) != 3 {
			log.Fatal("error: missing arguments")
		}

		if conf.syslog && !conf.force {
			log.Fatalf("error: forced operation is required when using syslog output")
		}

		err = blow(tag, f, name, flag.Arg(2))
	default:
		log.Fatal("error: invalid operation")
	}

	if err != nil {
		log.Fatalf("error: %v", err)
	}
}

func main() {
	var f *fusemap.FuseMap
	var v *fusemap.FuseMap
	var err error

	if conf.syslog {
		if logwriter, _ := syslog.New(syslog.LOG_INFO, "crucible"); logwriter != nil {
			log.SetOutput(logwriter)
		} else {
			log.SetOutput(os.Stderr)
		}
	} else {
		log.SetOutput(os.Stdout)
	}

	if len(conf.fusemaps) > 0 {
		stat, err := os.Stat(conf.fusemaps)

		if err != nil || !stat.IsDir() {
			log.Fatal("error: could not open fusemaps directory")
		}

		conf.fusemapDir = os.DirFS(conf.fusemaps)
	} else {
		conf.fusemapDir, err = fs.Sub(fusemaps, "fusemaps")

		if err != nil {
			log.Fatal("error: could not open fusemaps directory")
		}
	}

	if len(conf.fusemap) > 0 {
		if v, err = fusemap.Open(conf.fusemap); err != nil {
			log.Fatalf("error: could not open fusemap, %v", err)
		}

		conf.processor = v.Processor
		conf.reference = v.Reference
	}

	if conf.processor != "" && conf.reference != "" {
		if f, err = fusemap.Find(conf.fusemapDir, conf.processor, conf.reference); err != nil {
			log.Fatalf("error: could not open fusemap, %v", err)
		}
	}

	if v != nil {
		if err = f.Merge(v); err != nil {
			log.Fatalf("error: could not merge vendor and reference fusemaps, %v", err)
		}
	}

	if conf.list && len(flag.Args()) < 2 {
		if conf.processor != "" && conf.reference != "" {
			listFusemapRegisters(f)
		} else {
			listFusemaps()
		}

		return
	}

	op(f)
}
