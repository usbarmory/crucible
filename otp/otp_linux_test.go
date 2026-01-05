// crucible
// One-Time-Programmable (OTP) fusing tool
//
// Copyright (c) The crucible authors. All Rights Reserved.
//
// Use of this source code is governed by the license
// that can be found in the LICENSE file.

//go:build linux

package otp

import (
	"bytes"
	"os"
	"path/filepath"
	"testing"

	"github.com/usbarmory/crucible/fusemap"
)

var fusemaps = os.DirFS("../fusemaps")

func blowTest(t *testing.T, f *fusemap.FuseMap, path string, name string, val []byte, expRes []byte, expAddr uint32) {
	res, addr, _, _, err := BlowNVMEM(path, f, name, val)

	if err != nil {
		t.Fatal(err)
	}

	if !bytes.Equal(res, expRes) {
		t.Errorf("blown register %s with unexpected value, %x != %x", name, res, expRes)
	}

	if addr != expAddr {
		t.Errorf("blown register %s with unexpected address, %x != %x", name, addr, expAddr)
	}
}

func TestInvalidFuseMap(t *testing.T) {
	f := &fusemap.FuseMap{}

	_, _, _, _, err := BlowNVMEM("test", f, "test", []byte{0x00})

	if err == nil || err.Error() != "fusemap has not been validated yet" {
		t.Error("fusemap that has not been validated should raise an error")
	}

	_, _, _, _, err = ReadNVMEM("test", f, "test")

	if err == nil || err.Error() != "fusemap has not been validated yet" {
		t.Error("fusemap that has not been validated should raise an error")
	}
}

func TestBlowErrors(t *testing.T) {
	testYAML := `
---
reference: test
driver: nvmem-imx-ocotp
bank_size: 8
registers:
  REG1:
    fuses:
      OTP1:
        offset: 0
        len: 0
...
`

	f, err := fusemap.Parse([]byte(testYAML))

	if err != nil {
		t.Fatal(err)
	}

	_, _, _, _, err = BlowNVMEM("", f, "OTP1", []byte{})

	if err == nil || err.Error() != "null value" {
		t.Error("tripping a fuse with null length should raise an error")
	}

	_, _, _, _, err = BlowNVMEM("", f, "OTP2", []byte{0xff})

	if err == nil || err.Error() != "could not find any register/fuse named OTP2" {
		t.Error("tripping an invalid fuse should raise an error")
	}

	_, _, _, _, err = BlowNVMEM("invalid_file", f, "OTP1", []byte{0x00})

	if err == nil || err.Error() != "open invalid_file: no such file or directory" {
		t.Error("tripping a fuse with an invalid device should raise an error")
	}
}

func TestOverBlow(t *testing.T) {
	testYAML := `
---
reference: test
driver: nvmem-imx-ocotp
bank_size: 8
registers:
  REG1:
    fuses:
      OTP1:
        offset: 0
        len: 4
...
`

	f, err := fusemap.Parse([]byte(testYAML))

	if err != nil {
		t.Fatal(err)
	}

	_, _, _, _, err = BlowNVMEM("", f, "OTP1", []byte{0xff})

	if err == nil || err.Error() != "value bit length 8 exceeds 4" {
		t.Error("tripping a fuse with a value exceeding its size should raise an error")
	}

	_, _, _, _, err = BlowNVMEM("", f, "OTP1", []byte{0x02})

	if err != nil {
		t.Errorf("tripping a fuse with a value not exceeding its size should not raise an error (%v)", err)
	}

	_, _, _, _, err = BlowNVMEM("", f, "REG1", []byte{0xaa, 0xbb, 0xcc, 0xdd, 0xee})

	if err == nil || err.Error() != "value bit length 40 exceeds 32" {
		t.Error("tripping a register with a value exceeding its size should raise an error")
	}

	_, _, _, _, err = BlowNVMEM("", f, "REG1", []byte{0xaa, 0xbb, 0xcc, 0xdd})

	if err != nil {
		t.Errorf("tripping a register with a value not exceeding its size should not raise an error (%v)", err)
	}
}

func TestBlow(t *testing.T) {
	y := `
---
reference: test
driver: nvmem-imx-ocotp
bank_size: 8
registers:
  REG1:
    bank: 1
    word: 0
  REG2:
    bank: 1
    word: 1
    fuses:
      OTP1:
        offset: 1
        len: 3
      OTP2:
        offset: 4
        len: 48
  REG3:
    bank: 1
    word: 2
    fuses:
      OTP3:
        offset: 0
        len: 48
  REG4:
    bank: 1
    word: 3
    fuses:
      OTP4:
        offset: 4
        len: 60
...
`

	f, err := fusemap.Parse([]byte(y))

	if err != nil {
		t.Fatal(err)
	}

	blowTest(t, f, "", "REG1",
		[]byte{0x03},
		[]byte{0x03, 0x00, 0x00, 0x00},
		uint32(0x08*4))

	blowTest(t, f, "", "OTP1",
		[]byte{0x03},
		[]byte{0x06, 0x00, 0x00, 0x00},
		uint32(0x09*4))

	blowTest(t, f, "", "OTP2",
		[]byte{0xaa, 0xbb, 0xcc, 0xdd, 0xee, 0xff},
		[]byte{0xf0, 0xef, 0xde, 0xcd, 0xbc, 0xab, 0x0a, 0x00},
		uint32(0x09*4))

	blowTest(t, f, "", "OTP3",
		[]byte{0xaa, 0xbb, 0xcc, 0xdd, 0xee, 0xff},
		[]byte{0xff, 0xee, 0xdd, 0xcc, 0xbb, 0xaa, 0x00, 0x00},
		uint32(0x0a*4))

	blowTest(t, f, "", "OTP4",
		[]byte{0x0f, 0xaa, 0xbb, 0xcc, 0xdd, 0xee, 0xff, 0xaa},
		[]byte{0xa0, 0xfa, 0xef, 0xde, 0xcd, 0xbc, 0xab, 0xfa},
		uint32(0x0b*4))
}

func TestBlowIMX6UL(t *testing.T) {
	f, err := fusemap.Find(fusemaps, "IMX6UL", "1")

	if err != nil {
		t.Fatal(err)
	}

	blowTest(t, f, "", "SRK_LOCK",
		[]byte{0x01},
		[]byte{0x00, 0x40, 0x00, 0x00},
		uint32(0x00))

	blowTest(t, f, "", "MAC1_ADDR",
		[]byte{0xaa, 0xbb, 0xcc, 0xdd, 0xee, 0xff},
		[]byte{0xff, 0xee, 0xdd, 0xcc, 0xbb, 0xaa, 0x00, 0x00},
		uint32(0x22*4))
}

func TestBlowAndRead(t *testing.T) {
	y := `
---
reference: test
driver: nvmem-imx-ocotp
bank_size: 8
registers:
  REG1:
    bank: 0
    word: 0
    fuses:
      OTP1:
        offset: 8
        len: 256
...
`

	var nvram []byte

	for range 33 {
		nvram = append(nvram, 0xaa)
	}

	f, err := fusemap.Parse([]byte(y))

	if err != nil {
		t.Fatal(err)
	}

	tempDir, err := os.MkdirTemp("", "crucible_test-")

	defer func() {
		_ = os.RemoveAll(tempDir)
	}()

	if err != nil {
		t.Fatal(err)
	}

	tempFile := filepath.Join(tempDir, "nvram")
	err = os.WriteFile(tempFile, nvram, 0600)

	if err != nil {
		t.Fatal(err)
	}

	val := []byte{
		0xed, 0x58, 0x44, 0x66, 0xa8, 0xbc, 0x74, 0x89,
		0x29, 0x83, 0x46, 0x20, 0xee, 0x78, 0x82, 0x56,
		0xa0, 0x09, 0xf9, 0x23, 0xd9, 0xe2, 0xcc, 0x79,
		0xef, 0xc3, 0x59, 0xba, 0xaa, 0x22, 0x70, 0x7c,
	}

	expRes := []byte{
		0x00,
		0x7c, 0x70, 0x22, 0xaa, 0xba, 0x59, 0xc3, 0xef,
		0x79, 0xcc, 0xe2, 0xd9, 0x23, 0xf9, 0x09, 0xa0,
		0x56, 0x82, 0x78, 0xee, 0x20, 0x46, 0x83, 0x29,
		0x89, 0x74, 0xbc, 0xa8, 0x66, 0x44, 0x58, 0xed,
		0x00, 0x00, 0x00,
	}

	expAddr := uint32(0x00)

	blowTest(t, f, tempFile, "OTP1", val, expRes, expAddr)
	readTest(t, tempFile, f, "OTP1", val, expAddr)
}

func TestBlowIMX53(t *testing.T) {
	f, err := fusemap.Find(fusemaps, "IMX53", "2.1")

	if err != nil {
		t.Fatal(err)
	}

	_, _, _, _, err = BlowNVMEM("", f, "SRK_LOCK", []byte{0xff})

	if err == nil || err.Error() != "driver does not support blow operation" {
		t.Errorf("tripping a fuse on a read/only driver should raise an error")
	}
}

func readTest(t *testing.T, devicePath string, f *fusemap.FuseMap, name string, expRes []byte, expAddr uint32) {
	res, addr, _, _, err := ReadNVMEM(devicePath, f, name)

	if err != nil {
		t.Fatal(err)
	}

	if !bytes.Equal(res, expRes) {
		t.Errorf("read register %s with unexpected value, %x != %x", name, res, expRes)
	}

	if addr != expAddr {
		t.Errorf("read register %s with unexpected address, %x != %x", name, addr, expAddr)
	}
}

func TestReadErrors(t *testing.T) {
	f, err := fusemap.Find(fusemaps, "IMX53", "2.1")

	if err != nil {
		t.Fatal(err)
	}

	_, _, _, _, err = ReadNVMEM("", f, "OTP1")

	if err == nil || err.Error() != "empty device path" {
		t.Error("reading a fuse with an invalid device should raise an error")
	}
}

func TestReadIMX53(t *testing.T) {
	f, err := fusemap.Find(fusemaps, "IMX53", "2.1")

	if err != nil {
		t.Fatal(err)
	}

	devicePath := "../test/nvmem.IMX53"

	// register
	readTest(t, devicePath, f, "BANK0_WORD0", []byte{0x10}, uint32(0x00))

	// fuses
	readTest(t, devicePath, f, "SRK_LOCK", []byte{0x01}, uint32(0x20))
	readTest(t, devicePath, f, "SRK_HASH[255:248]", []byte{0x5d}, uint32(0x21))
	readTest(t, devicePath, f, "SRK_HASH[247:0]",
		[]byte{
			0x85, 0xe5, 0xaf, 0x63, 0xd0, 0xb6, 0x6c, 0xb4,
			0x6e, 0x18, 0x09, 0x3e, 0x94, 0xad, 0x70, 0x94,
			0x51, 0x54, 0xd7, 0xbc, 0xc5, 0xa6, 0x26, 0x77,
			0xe7, 0x11, 0x21, 0x8e, 0x0a, 0xb4, 0xa9,
		},
		uint32(0x61))
	readTest(t, devicePath, f, "SJC_CHALL",
		[]byte{
			0x80, 0x41, 0x00, 0x51, 0x06, 0x38, 0x05, 0x1b,
		},
		uint32(0x08))
}

func TestReadIMX6UL(t *testing.T) {
	f, err := fusemap.Find(fusemaps, "IMX6UL", "1")

	if err != nil {
		t.Fatal(err)
	}

	devicePath := "../test/nvmem.IMX6UL"

	// register
	readTest(t, devicePath, f, "OCOTP_OTPMK0", []byte{0xba, 0xda, 0xba, 0xda}, uint32(0x40))

	// fuses
	readTest(t, devicePath, f, "SRK_LOCK", []byte{0x01}, uint32(0x00))
	readTest(t, devicePath, f, "SRK_HASH",
		[]byte{
			0xed, 0x58, 0x44, 0x66, 0xa8, 0xbc, 0x74, 0x89,
			0x29, 0x83, 0x46, 0x20, 0xee, 0x78, 0x82, 0x56,
			0xa0, 0x09, 0xf9, 0x23, 0xd9, 0xe2, 0xcc, 0x79,
			0xef, 0xc3, 0x59, 0xba, 0xaa, 0x22, 0x70, 0x7c,
		},
		uint32(0x18*4))
	readTest(t, devicePath, f, "MAC1_ADDR",
		[]byte{
			0x00, 0x1f, 0x7b, 0x10, 0x07, 0xe3,
		},
		uint32(0x22*4))
	// test post gap addressing
	readTest(t, devicePath, f, "OCOTP_GP30",
		[]byte{
			0x00, 0x00, 0x00, 0x00,
		},
		uint32(0x140))
	readTest(t, devicePath, f, "GP3",
		[]byte{
			0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00,
			0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00,
			0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00,
			0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00,
			0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00,
			0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00,
			0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00,
			0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00,
		},
		uint32(0x140))
	readTest(t, devicePath, f, "GP3[511:480]",
		[]byte{
			0x00, 0x00, 0x00, 0x00,
		},
		uint32(0x140))
}

func TestReadBitMap8(t *testing.T) {
	exp := ` 07 06 05 04 03 02 01 00  BANK0_WORD0
┏━━┳━━┳━━┳━━┳━━┳━━┳━━┳━━┓ Bank:0 Word:0
┃0 ┃0 ┃0 ┃1 ┃0 ┃0 ┃0 ┃0 ┃ R: 0x00000000
┗━━┻━━┻━━┻━━┻━━┻━━┻━━┻━━┛ W: 0x00000000
 07 ────────────────────  FBWP
    06 ─────────────────  FBOP
       05 ──────────────  FBRP
          04 ───────────  TESTER_LOCK
             03 ────────  FBESP
                02 ─────  TESTER_LOCK2
                   01 ──  GP_LOCK
                      00  BOOT_LOCK
`

	f, err := fusemap.Find(fusemaps, "IMX53", "2.1")

	if err != nil {
		t.Fatal(err)
	}

	devicePath := "../test/nvmem.IMX53"
	name := "BANK0_WORD0"

	res, _, _, _, err := ReadNVMEM(devicePath, f, name)

	if err != nil {
		t.Fatal(err)
	}

	m := f.Registers[name].BitMap(res)

	if m != exp {
		t.Errorf("unexpected map\n%s\n  !=\n%s", m, exp)
	}
}

func TestReadBitMap32(t *testing.T) {
	exp := ` 31 30 29 28 27 26 25 24 23 22 21 20 19 18 17 16 15 14 13 12 11 10 09 08 07 06 05 04 03 02 01 00  OCOTP_CFG1
┏━━┳━━┳━━┳━━┳━━┳━━┳━━┳━━┳━━┳━━┳━━┳━━┳━━┳━━┳━━┳━━┳━━┳━━┳━━┳━━┳━━┳━━┳━━┳━━┳━━┳━━┳━━┳━━┳━━┳━━┳━━┳━━┓ Bank:0 Word:2
┃0  0  1  0  0  1  1  1 ┃0  0  0  1  0  0  0  0 ┃0  1  0  0  0 ┃0  0  1  1  1  0  1  0  1  0  0 ┃ R: 0x00000008
┗━━┻━━┻━━┻━━┻━━┻━━┻━━┻━━╋━━┻━━┻━━┻━━┻━━┻━━┻━━┻━━╋━━┻━━┻━━┻━━┻━━┻━━┻━━┻━━╋━━┻━━┻━━┻━━┻━━┻━━┻━━┻━━┛ W: 0x00000008
 31 ┄┄ ┄┄ ┄┄ ┄┄ ┄┄ ┄┄ ┄┄ ┄┄ ┄┄ ┄┄ ┄┄ ┄┄ ┄┄ ┄┄ ┄┄ ┄┄ ┄┄ ┄┄ ┄┄ ┄┄ ┄┄ ┄┄ ┄┄ ┄┄ ┄┄ ┄┄ ┄┄ ┄┄ ┄┄ ┄┄ 00  UNIQUE_ID[63:32]
 31 ┄┄ ┄┄ ┄┄ ┄┄ ┄┄ ┄┄ ┄┄ ┄┄ ┄┄ ┄┄ ┄┄ ┄┄ ┄┄ ┄┄ ┄┄ ┄┄ ┄┄ ┄┄ ┄┄ ┄┄ ┄┄ ┄┄ ┄┄ ┄┄ ┄┄ ┄┄ ┄┄ ┄┄ ┄┄ ┄┄ 00  SJC_CHALLENGE[63:32]
 31 ┄┄ ┄┄ ┄┄ ┄┄ ┄┄ ┄┄ 24 ───────────────────────────────────────────────────────────────────────  DIE-X-CORDINATE
                         23 ┄┄ ┄┄ ┄┄ ┄┄ ┄┄ ┄┄ 16 ───────────────────────────────────────────────  DIE-Y-CORDINATE
                                                 15 ┄┄ ┄┄ ┄┄ 11 ────────────────────────────────  WAFER_NO
                                                                10 ┄┄ ┄┄ ┄┄ ┄┄ ┄┄ ┄┄ ┄┄ ┄┄ ┄┄ 00  LOT_NO_ENC[42:32]
`

	f, err := fusemap.Find(fusemaps, "IMX6UL", "1")

	if err != nil {
		t.Fatal(err)
	}

	devicePath := "../test/nvmem.IMX6UL"
	name := "OCOTP_CFG1"

	res, _, _, _, err := ReadNVMEM(devicePath, f, name)

	if err != nil {
		t.Fatal(err)
	}

	m := f.Registers[name].BitMap(res)

	if m != exp {
		t.Errorf("unexpected map\n%s\n  !=\n%s", m, exp)
	}
}
