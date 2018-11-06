// crucible
// One-Time-Programmable (OTP) fusing tool
//
// Copyright (c) F-Secure Corporation
//
// Use of this source code is governed by the license
// that can be found in the LICENSE file.

package crucible

import (
	"errors"
	"math/big"
	"os"
)

// Read a register or fuse through Linux NVMEM subsystem framework. The name
// argument could be a register or an individual OTP fuse.
func (fusemap *FuseMap) Read(devicePath string, name string) (res []byte, addr uint32, off uint32, size uint32, err error) {
	if devicePath == "" {
		err = errors.New("empty device path")
		return
	}

	mapping, err := fusemap.Find(name)

	if err != nil {
		return
	}

	switch mapping.(type) {
	case *Register:
		addr = mapping.(*Register).ReadAddress
		off = 0
		size = uint32(32)
	case *Fuse:
		addr = mapping.(*Fuse).ReadAddress
		off = mapping.(*Fuse).Offset
		size = mapping.(*Fuse).Length
	}

	device, err := os.OpenFile(devicePath, os.O_RDONLY|os.O_EXCL|os.O_SYNC, 0600)

	if err != nil {
		return
	}
	// make errcheck happy
	defer func() { _ = device.Close() }()

	_, err = device.Seek(int64(addr), 0)

	if err != nil {
		return
	}

	numRegisters := 1 + (off+size)/32

	// normalize
	if (off+size)%32 == 0 {
		numRegisters -= 1
	}

	val := make([]byte, numRegisters*4)
	_, err = device.Read(val)

	if err != nil {
		return
	}

	res = ConvertReadValue(off, size, val)

	return
}

// Convert read value, shifted accordingly to its register offset and size, to
// a big endian array of 32-bit registers.
func ConvertReadValue(off uint32, size uint32, val []byte) (res []byte) {
	// little-endian > big-endian
	res = SwitchEndianess(val)

	v := new(big.Int)
	v.SetBytes(res)
	v.Rsh(v, uint(off))

	// get only the bits that we care about
	mask := big.NewInt((1 << size) - 1)
	v.And(v, mask)

	res = PadBigInt(v, size)

	return
}
