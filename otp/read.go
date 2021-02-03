// crucible
// One-Time-Programmable (OTP) fusing tool
//
// Copyright (c) F-Secure Corporation
//
// Use of this source code is governed by the license
// that can be found in the LICENSE file.

package otp

import (
	"errors"
	"os"

	"github.com/f-secure-foundry/crucible/fusemap"
	"github.com/f-secure-foundry/crucible/util"
)

// Read a register or fuse through Linux NVMEM subsystem framework. The name
// argument could be a register or an individual OTP fuse.
func Read(devicePath string, f *fusemap.FuseMap, name string) (res []byte, addr uint32, off uint32, size uint32, err error) {
	if devicePath == "" {
		err = errors.New("empty device path")
		return
	}

	if !f.Valid() {
		err = errors.New("fusemap has not been validated yet")
		return
	}

	mapping, err := f.Find(name)

	if err != nil {
		return
	}

	regSize := 8 * f.WordSize

	switch m := mapping.(type) {
	case *fusemap.Register:
		reg := m
		addr = reg.ReadAddress
		off = 0
		size = regSize
	case *fusemap.Fuse:
		fuse := m
		addr = fuse.Register.ReadAddress
		off = fuse.Offset
		size = fuse.Length
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

	numRegisters := 1 + (off+size)/regSize

	// normalize
	if (off+size)%regSize == 0 {
		numRegisters -= 1
	}

	val := make([]byte, numRegisters*f.WordSize)
	_, err = device.Read(val)

	if err != nil {
		return
	}

	res = util.ConvertReadValue(off, size, val)

	return
}
