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
	"bytes"
	"embed"
	"errors"
	"flag"
	"fmt"
	"io/fs"
	"log"
	"log/syslog"
	"math/big"
	"os"
	"path/filepath"
	"strings"
	"text/tabwriter"

	"github.com/usbarmory/crucible/fusemap"
	"github.com/usbarmory/crucible/otp"
	"github.com/usbarmory/crucible/util"
)

type Config struct {
	force      bool
	list       bool
	syslog     bool
	base       int
	endianness string
	device     string
	fusemaps   string
	processor  string
	reference  string

	fusemapDir fs.FS
}

// Bundled fusemaps
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
		log.Println(splash)
		log.Printf("Usage: crucible [options] [read|blow] [fuse/register name] [value]\n")
		flag.PrintDefaults()
	}

	flag.BoolVar(&conf.force, "Y", false, "do not prompt for confirmation (DANGEROUS)")
	flag.BoolVar(&conf.list, "l", false, "list fusemaps\nvisualize fusemap      (with -m and -r)\nvisualize read value   (with read operation on a register)\nvisualize read fusemap (with read operation and no register)")
	flag.BoolVar(&conf.syslog, "s", false, "use syslog, print only result value to stdout")
	flag.IntVar(&conf.base, "b", 0, "value base/format (2,10,16)")
	flag.StringVar(&conf.endianness, "e", "", "value endianness (big,little)")
	flag.StringVar(&conf.device, "n", "/sys/bus/nvmem/devices/imx-ocotp0/nvmem", "NVMEM device")
	flag.StringVar(&conf.fusemaps, "f", "", "YAML fusemaps directory")
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
	f, err := fusemap.Find(conf.fusemapDir, conf.processor, conf.reference)

	if err != nil {
		log.Fatalf("error: could not open fusemap, %v", err)
	}

	var res []byte

	for _, reg := range f.RegistersByWriteAddress() {
		if flag.Arg(0) == "read" {
			res, _, _, _, err = otp.ReadNVMEM(conf.device, f, reg.Name)

			if err != nil {
				log.Fatalf("error: could not read fusemap, %v", err)
			}

			n := new(big.Int)
			n.SetBytes(res)
		}

		fmt.Print(reg.BitMap(res))
		fmt.Println()
	}
}

func listFusemaps() {
	var list bytes.Buffer

	_, _ = fmt.Fprintf(&list, "Model (-m)\tReference (-r)\tDriver\n")

	t := tabwriter.NewWriter(&list, 16, 8, 0, '\t', tabwriter.TabIndent)

	_ = fs.WalkDir(conf.fusemapDir, ".", func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}

		if d.IsDir() {
			return nil
		}

		if filepath.Ext(path) != ".yaml" {
			return nil
		}

		y, err := fs.ReadFile(conf.fusemapDir, path)

		if err != nil {
			return err
		}

		f, err := fusemap.Parse(y)

		if err != nil {
			log.Printf("skipping %s (%v)", path, err)
			return nil
		}

		_, _ = fmt.Fprintf(t, "%s\t%s\t%s\n", f.Processor, f.Reference, f.Driver)

		return nil
	})

	_ = t.Flush()

	fmt.Printf(list.String())
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
		log.Println(warning)
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

func main() {
	var err error

	flag.Parse()

	if conf.syslog {
		logwriter, _ := syslog.New(syslog.LOG_INFO, "crucible")
		log.SetOutput(logwriter)
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

	if conf.list && len(flag.Args()) < 2 {
		if conf.processor != "" && conf.reference != "" {
			listFusemapRegisters()
		} else {
			listFusemaps()
		}

		return
	}

	if err = checkArguments(); err != nil {
		flag.Usage()
		log.Fatalf("error: %v", err)
	}

	stat, err := os.Stat(conf.device)

	if err != nil || stat.IsDir() {
		log.Fatalf("error: could not open NVMEM device %s", conf.device)
	}

	f, err := fusemap.Find(conf.fusemapDir, conf.processor, conf.reference)

	if err != nil {
		log.Fatalf("error: could not open fusemap, %v", err)
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
