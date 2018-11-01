// crucible
// One-Time-Programmable (OTP) fusing tool
//
// Copyright (c) F-Secure Corporation
//
// Use of this source code is governed by the license
// that can be found in the LICENSE file.

package crucible

import (
	"bytes"
	"io/ioutil"
	"os"
	"path/filepath"
	"testing"
)

func blowTest(t *testing.T, fusemap FuseMap, path string, name string, val []byte, expRes []byte, expAddr uint32) {
	res, addr, _, _, err := fusemap.Blow(path, name, val)

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

func TestBlowErrors(t *testing.T) {
	testYAML := `
---
reference: test
driver: nvmem-imx-ocotp
registers:
  REG1:
    fuses:
      OTP1:
        offset: 0
        len: 0
...
`

	fusemap, err := ParseFuseMap([]byte(testYAML))

	if err != nil {
		t.Fatal(err)
	}

	_, _, _, _, err = fusemap.Blow("", "OTP1", []byte{})

	if err == nil || err.Error() != "null value" {
		t.Error("tripping a fuse with null length should raise an error")
	}

	_, _, _, _, err = fusemap.Blow("", "OTP2", []byte{0xff})

	if err == nil || err.Error() != "could not find any register/fuse named OTP2" {
		t.Error("tripping an invalid fuse should raise an error")
	}

	_, _, _, _, err = fusemap.Blow("invalid_file", "OTP1", []byte{0x00})

	if err == nil || err.Error() != "open invalid_file: no such file or directory" {
		t.Error("tripping a fuse with an invalid device should raise an error")
	}
}

func TestOverBlow(t *testing.T) {
	testYAML := `
---
reference: test
driver: nvmem-imx-ocotp
registers:
  REG1:
    fuses:
      OTP1:
        offset: 0
        len: 4
...
`

	fusemap, err := ParseFuseMap([]byte(testYAML))

	if err != nil {
		t.Fatal(err)
	}

	_, _, _, _, err = fusemap.Blow("", "OTP1", []byte{0xff})

	if err == nil || err.Error() != "value bit size 8 exceeds 4" {
		t.Error("tripping a fuse with a value exceeding its size should raise an error")
	}

	_, _, _, _, err = fusemap.Blow("", "OTP1", []byte{0x02})

	if err != nil {
		t.Errorf("tripping a fuse with a value not exceeding its size should not raise an error (%v)", err)
	}

	_, _, _, _, err = fusemap.Blow("", "REG1", []byte{0xaa, 0xbb, 0xcc, 0xdd, 0xee})

	if err == nil || err.Error() != "value bit size 40 exceeds 32" {
		t.Error("tripping a register with a value exceeding its size should raise an error")
	}

	_, _, _, _, err = fusemap.Blow("", "REG1", []byte{0xaa, 0xbb, 0xcc, 0xdd})

	if err != nil {
		t.Errorf("tripping a register with a value not exceeding its size should not raise an error (%v)", err)
	}
}

func TestBlow(t *testing.T) {
	y := `
---
reference: test
driver: nvmem-imx-ocotp
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

	fusemap, err := ParseFuseMap([]byte(y))

	if err != nil {
		t.Fatal(err)
	}

	blowTest(t, fusemap, "", "REG1",
		[]byte{0x03},
		[]byte{0x03, 0x00, 0x00, 0x00},
		uint32(0x08*4))

	blowTest(t, fusemap, "", "OTP1",
		[]byte{0x03},
		[]byte{0x06, 0x00, 0x00, 0x00},
		uint32(0x09*4))

	blowTest(t, fusemap, "", "OTP2",
		[]byte{0xaa, 0xbb, 0xcc, 0xdd, 0xee, 0xff},
		[]byte{0xf0, 0xef, 0xde, 0xcd, 0xbc, 0xab, 0x0a, 0x00},
		uint32(0x09*4))

	blowTest(t, fusemap, "", "OTP3",
		[]byte{0xaa, 0xbb, 0xcc, 0xdd, 0xee, 0xff},
		[]byte{0xff, 0xee, 0xdd, 0xcc, 0xbb, 0xaa, 0x00, 0x00},
		uint32(0x0a*4))

	blowTest(t, fusemap, "", "OTP4",
		[]byte{0x0f, 0xaa, 0xbb, 0xcc, 0xdd, 0xee, 0xff, 0xaa},
		[]byte{0xa0, 0xfa, 0xef, 0xde, 0xcd, 0xbc, 0xab, 0xfa},
		uint32(0x0b*4))
}

func TestBlowIMX6UL(t *testing.T) {
	fusemap, err := OpenFuseMap("../../fusemaps", "IMX6UL", "1")

	if err != nil {
		t.Fatal(err)
	}

	blowTest(t, fusemap, "", "SRK_LOCK",
		[]byte{0x01},
		[]byte{0x00, 0x40, 0x00, 0x00},
		uint32(0x00))

	blowTest(t, fusemap, "", "MAC1_ADDR",
		[]byte{0xaa, 0xbb, 0xcc, 0xdd, 0xee, 0xff},
		[]byte{0xff, 0xee, 0xdd, 0xcc, 0xbb, 0xaa, 0x00, 0x00},
		uint32(0x22*4))
}

func TestBlowAndRead(t *testing.T) {
	y := `
---
reference: test
driver: nvmem-imx-ocotp
registers:
  REG1:
    bank: 0
    word: 0
    fuses:
      OTP1:
        offset: 0
        len: 256
...
`

	var nvram []byte

	for i := 0; i < 33; i++ {
		nvram = append(nvram, 0xaa)
	}

	fusemap, err := ParseFuseMap([]byte(y))

	if err != nil {
		t.Fatal(err)
	}

	tempDir, err := ioutil.TempDir("", "crucible_test-")

	defer func() {
		_ = os.RemoveAll(tempDir)
	}()

	if err != nil {
		t.Fatal(err)
	}

	tempFile := filepath.Join(tempDir, "nvram")
	err = ioutil.WriteFile(tempFile, nvram, 0600)

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
		0x7c, 0x70, 0x22, 0xaa, 0xba, 0x59, 0xc3, 0xef,
		0x79, 0xcc, 0xe2, 0xd9, 0x23, 0xf9, 0x09, 0xa0,
		0x56, 0x82, 0x78, 0xee, 0x20, 0x46, 0x83, 0x29,
		0x89, 0x74, 0xbc, 0xa8, 0x66, 0x44, 0x58, 0xed,
	}

	expAddr := uint32(0x00)

	blowTest(t, fusemap, tempFile, "OTP1", val, expRes, expAddr)
	readTest(t, tempFile, fusemap, "OTP1", val, expAddr)
}

func TestBlowIMX53(t *testing.T) {
	fusemap, err := OpenFuseMap("../../fusemaps", "IMX53", "2.1")

	if err != nil {
		t.Fatal(err)
	}

	_, _, _, _, err = fusemap.Blow("", "SRK_LOCK", []byte{0xff})

	if err == nil || err.Error() != "driver does not support blow operation" {
		t.Errorf("tripping a fuse on a read/only driver should raise an error")
	}
}
