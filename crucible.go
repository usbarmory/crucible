// crucible
// One-Time-Programmable (OTP) fusing tool
//
// Copyright (c) F-Secure Corporation
//
// Use of this source code is governed by the license
// that can be found in the LICENSE file.

package main

import (
	"bufio"
	"errors"
	"flag"
	"fmt"
	"io/ioutil"
	"log"
	"log/syslog"
	"math/big"
	"os"
	"path/filepath"
	"strings"
	"text/tabwriter"

	"github.com/inversepath/crucible/lib"
)

// build information, initialized at compile time (see Makefile)
var Revision string
var Build string

type Config struct {
	force     bool
	list      bool
	syslog    bool
	base      int
	device    string
	fusemaps  string
	processor string
	reference string
}

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

The input value is assumed to have *big endian* order, make sure this is what
you expect.

                                ** IMPORTANT **

The value parameter endianness is always assumed to be big-endian, it is then
converted to little-endian before writing, as required by the driver. Please
note that certain tools, such as the ones creating the SRK_HASH for secure boot
purposes, typically already prepare their output in little-endian format.

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

		log.Printf("crucible - One-Time-Programmable (OTP) fusing tool %s (%s)", Revision, Build)
		log.Println(splash)
		log.Printf("Usage: crucible [options] [read|blow] [fuse/register name] [value]\n")
		flag.PrintDefaults()
	}

	flag.BoolVar(&conf.force, "Y", false, "do not prompt for confirmation (DANGEROUS)")
	flag.BoolVar(&conf.list, "l", false, "list fusemaps or fusemap registers (with -m and -r)")
	flag.BoolVar(&conf.syslog, "s", false, "use syslog, print only result value to stdout")
	flag.IntVar(&conf.base, "b", 0, "value base/format (2,10,16)")
	flag.StringVar(&conf.device, "n", "/sys/bus/nvmem/devices/imx-ocotp0/nvmem", "NVMEM device")
	flag.StringVar(&conf.fusemaps, "f", "fusemaps", "YAML fuse maps directory")
	flag.StringVar(&conf.processor, "m", "", "processor model")
	flag.StringVar(&conf.reference, "r", "", "reference manual revision")
}

func confirm() bool {
	reader := bufio.NewReader(os.Stdin)
	log.Print("Would you really like to blow this fuse? Type YES all uppercase to confirm: ")
	text, _ := reader.ReadString('\n')

	return text == "YES\n"
}

func listFusemapRegisters() {
	fusemap, err := crucible.OpenFuseMap(conf.fusemaps, conf.processor, conf.reference)

	if err != nil {
		log.Fatalf("error: could not open fuse map, %v", err)
	}

	for _, reg := range fusemap.RegistersByWriteAddress() {
		fmt.Print(reg.BitMap())
		fmt.Println()
	}
}

func listFusemaps() {
	t := tabwriter.NewWriter(os.Stdout, 16, 8, 0, '\t', tabwriter.TabIndent)

	_, _ = fmt.Printf("Model (-m)\tReference (-r)\tDriver\n")

	_ = filepath.Walk(conf.fusemaps, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		if info.IsDir() {
			return nil
		}

		if filepath.Ext(path) != ".yaml" {
			return nil
		}

		y, err := ioutil.ReadFile(path)

		if err != nil {
			return err
		}

		fusemap, err := crucible.ParseFuseMap(y)

		if err != nil {
			return err
		}

		_, _ = fmt.Fprintf(t, "%s\t%s\t%s\n", fusemap.Processor, fusemap.Reference, fusemap.Driver)

		return nil
	})

	_ = t.Flush()
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

func read(tag string, fusemap crucible.FuseMap, name string) (err error) {
	res, addr, off, size, err := fusemap.Read(conf.device, name)

	if err != nil {
		return
	}

	var value string
	var stdout string

	tag = fmt.Sprintf("%s addr:%#x off:%d len:%d", tag, addr, off, size)

	switch conf.base {
	case 2:
		value = fmt.Sprintf("0b%.8b", res)

		n := new(big.Int)
		n.SetBytes(res)

		stdout = fmt.Sprintf("%b", n)
	case 10:
		n := new(big.Int)
		n.SetBytes(res)

		value = fmt.Sprintf("%d", n)
		stdout = value
	case 16:
		value = fmt.Sprintf("%#x", res)
		stdout = fmt.Sprintf("%x", res)
	default:
		return errors.New("internal error, invalid base")
	}

	log.Printf("%s val:%s", tag, value)

	if conf.syslog {
		fmt.Println(stdout)
	}

	return
}

func blow(tag string, fusemap crucible.FuseMap, name string, val string) (err error) {
	n := new(big.Int)

	switch conf.base {
	case 2:
		val = strings.TrimPrefix(flag.Arg(2), "0b")
	case 10:
	case 16:
		val = strings.TrimPrefix(flag.Arg(2), "0x")
	default:
		return errors.New("internal error, invalid base")
	}

	n, ok := n.SetString(val, conf.base)

	if !ok {
		return errors.New("invalid value argument")
	}

	if !conf.force {
		log.Printf("%s reg:%s base:%d val:%s confirm?", tag, name, conf.base, val)
		log.Println(warning)

		if !confirm() {
			log.Fatal("you are not ready...")
		}
	}

	res, addr, off, size, err := fusemap.Blow(conf.device, name, n.Bytes())

	if err != nil {
		return err
	}

	log.Printf("%s addr:%#x off:%d len:%d val:%#x", tag, addr, off, size, res)

	if conf.syslog {
		fmt.Printf("%#x\n", res)
	}

	return
}

func main() {
	flag.Parse()

	if conf.syslog {
		logwriter, _ := syslog.New(syslog.LOG_INFO, "crucible")
		log.SetOutput(logwriter)
	} else {
		log.SetOutput(os.Stdout)
	}

	if conf.list {
		if conf.processor != "" && conf.reference != "" {
			listFusemapRegisters()
		} else {
			listFusemaps()
		}

		return
	}

	stat, err := os.Stat(conf.fusemaps)

	if err != nil || !stat.IsDir() {
		log.Fatalf("error: could not open fuse maps directory %s", conf.fusemaps)
	}

	err = checkArguments()

	if err != nil {
		flag.Usage()
		log.Fatalf("error: %v", err)
	}

	stat, err = os.Stat(conf.device)

	if err != nil || stat.IsDir() {
		log.Fatalf("error: could not open NVMEM device %s", conf.device)
	}

	fusemap, err := crucible.OpenFuseMap(conf.fusemaps, conf.processor, conf.reference)

	if err != nil {
		log.Fatalf("error: could not open fuse map, %v", err)
	}

	op := flag.Arg(0)
	name := flag.Arg(1)
	tag := fmt.Sprintf("%s ref:%s op:%s", conf.processor, conf.reference, op)

	switch op {
	case "read":
		err = read(tag, fusemap, name)
	case "blow":
		if len(flag.Args()) != 3 {
			log.Fatal("error: missing arguments")
		}

		if conf.syslog && !conf.force {
			log.Fatalf("error: forced operation is required when using syslog output")
		}

		err = blow(tag, fusemap, name, flag.Arg(2))
	default:
		log.Fatal("error: invalid operation")
	}

	if err != nil {
		log.Fatalf("error: %v", err)
	}
}
