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
	"testing"
)

func readTest(t *testing.T, devicePath string, fusemap FuseMap, name string, expRes []byte, expAddr uint32) {
	res, addr, _, _, err := fusemap.Read(devicePath, name)

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
	fusemap, err := OpenFuseMap("../fusemaps", "IMX53", "2.1")

	if err != nil {
		t.Fatal(err)
	}

	_, _, _, _, err = fusemap.Read("", "OTP1")

	if err == nil || err.Error() != "empty device path" {
		t.Error("reading a fuse with an invalid device should raise an error")
	}
}

func TestReadIMX53(t *testing.T) {
	fusemap, err := OpenFuseMap("../fusemaps", "IMX53", "2.1")

	if err != nil {
		t.Fatal(err)
	}

	devicePath := "../test/nvmem.IMX53"

	// register
	readTest(t, devicePath, fusemap, "BANK0_WORD0", []byte{0x00, 0x10, 0x00, 0x10}, uint32(0x00))

	// fuses
	readTest(t, devicePath, fusemap, "SRK_LOCK", []byte{0x01}, uint32(0x20))
	readTest(t, devicePath, fusemap, "SRK_HASH0", []byte{0x5d}, uint32(0x21))
	readTest(t, devicePath, fusemap, "SRK_HASH1",
		[]byte{
			0x85, 0xe5, 0xaf, 0x63, 0xd0, 0xb6, 0x6c, 0xb4,
			0x6e, 0x18, 0x09, 0x3e, 0x94, 0xad, 0x70, 0x94,
			0x51, 0x54, 0xd7, 0xbc, 0xc5, 0xa6, 0x26, 0x77,
			0xe7, 0x11, 0x21, 0x8e, 0x0a, 0xb4, 0xa9,
		},
		uint32(0x61))
}

func TestReadIMX6UL(t *testing.T) {
	fusemap, err := OpenFuseMap("../fusemaps", "IMX6UL", "1")

	if err != nil {
		t.Fatal(err)
	}

	devicePath := "../test/nvmem.IMX6UL"

	// register
	readTest(t, devicePath, fusemap, "OCOTP_OTPMK0", []byte{0xba, 0xda, 0xba, 0xda}, uint32(0x40))

	// fuses
	readTest(t, devicePath, fusemap, "SRK_LOCK", []byte{0x01}, uint32(0x00))
	readTest(t, devicePath, fusemap, "SRK_HASH",
		[]byte{
			0xed, 0x58, 0x44, 0x66, 0xa8, 0xbc, 0x74, 0x89,
			0x29, 0x83, 0x46, 0x20, 0xee, 0x78, 0x82, 0x56,
			0xa0, 0x09, 0xf9, 0x23, 0xd9, 0xe2, 0xcc, 0x79,
			0xef, 0xc3, 0x59, 0xba, 0xaa, 0x22, 0x70, 0x7c,
		},
		uint32(0x18*4))
	readTest(t, devicePath, fusemap, "MAC1_ADDR",
		[]byte{
			0x00, 0x1f, 0x7b, 0x10, 0x07, 0xe3,
		},
		uint32(0x22*4))
}
