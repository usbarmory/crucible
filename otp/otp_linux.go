// crucible
// One-Time-Programmable (OTP) fusing tool
//
// Copyright (c) WithSecure Corporation
//
// Use of this source code is governed by the license
// that can be found in the LICENSE file.
//
// +build linux

package otp

import (
	"errors"
	"os"

	"github.com/usbarmory/crucible/fusemap"
	"github.com/usbarmory/crucible/util"
)

// BlowNVMEM a fuse through Linux NVMEM subsystem framework, returns the input
// value converted as required for the fusing operation as well as the written
// address. The name argument could be a register or an individual OTP fuse.
//
// An empty NVMEM device path is allowed to simulate the operation and test
// returned values.
//
// The value parameter is interpreted as a big-endian value, please note that
// certain tools, such as the ones creating the `SRK_HASH` for secure boot
// purposes, typically prepare their output in little-endian format.
//
// WARNING: Fusing SoC OTPs is an **irreversible** action that permanently
// fuses values on the device. This means that any errors in the process, or
// lost fused data such as cryptographic key material, might result in a
// **bricked** device.
//
// The use of this function is therefore **at your own risk**.
func BlowNVMEM(devicePath string, f *fusemap.FuseMap, name string, val []byte) (res []byte, addr uint32, off int, bitLen int, err error) {
	if len(val) == 0 {
		err = errors.New("null value")
		return
	}

	if !f.Valid() {
		err = errors.New("fusemap has not been validated yet")
		return
	}

	if f.Driver != "nvmem-imx-ocotp" {
		err = errors.New("driver does not support blow operation")
		return
	}

	mapping, err := f.Find(name)

	if err != nil {
		return
	}

	if err != nil {
		return
	}

	switch m := mapping.(type) {
	case *fusemap.Register:
		reg := m
		addr = reg.WriteAddress
		off = 0
		bitLen = reg.Length
	case *fusemap.Fuse:
		fuse := m
		addr = fuse.Register.WriteAddress
		off = fuse.Offset
		bitLen = fuse.Length
	}

	res, err = util.ConvertWriteValue(off, bitLen, val)

	if err != nil {
		return
	}

	res = util.Pad4(res)

	if devicePath == "" {
		return
	}

	device, err := os.OpenFile(devicePath, os.O_WRONLY|os.O_EXCL|os.O_SYNC, 0600)

	if err != nil {
		return
	}

	// nvmem-imx-ocotp allows only one complete OTP word write at a time
	for i := 0; i < len(res); i += f.WordSize {
		_, err = device.Seek(int64(addr)+int64(i), 0)

		if err != nil {
			_ = device.Close()
			return
		}

		_, err = device.Write(res[i : i+f.WordSize])
	}

	_ = device.Close()

	return
}

// ReadNVMEM reads a register or fuse through Linux NVMEM subsystem framework.
// The name argument could be a register or an individual OTP fuse.
func ReadNVMEM(devicePath string, f *fusemap.FuseMap, name string) (res []byte, addr uint32, off int, bitLen int, err error) {
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
		bitLen = regSize
	case *fusemap.Fuse:
		fuse := m
		addr = fuse.Register.ReadAddress
		off = fuse.Offset
		bitLen = fuse.Length
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

	numRegisters := 1 + (off+bitLen)/regSize

	// normalize
	if (off+bitLen)%regSize == 0 {
		numRegisters -= 1
	}

	val := make([]byte, numRegisters*f.WordSize)
	_, err = device.Read(val)

	if err != nil {
		return
	}

	res = util.ConvertReadValue(off, bitLen, val)

	return
}
