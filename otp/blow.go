// crucible
// One-Time-Programmable (OTP) fusing tool
//
// Copyright (c) F-Secure Corporation
//
// Use of this source code is governed by the license
// that can be found in the LICENSE file.

// Package otp provides support for One-Time-Programmable (OTP) fuses read and
// write operations.
package otp

import (
	"errors"
	"os"

	"github.com/f-secure-foundry/crucible/fusemap"
	"github.com/f-secure-foundry/crucible/util"
)

// Blow a fuse through Linux NVMEM subsystem framework, returns the input value
// converted as required for the fusing operation as well as the written
// address. The name argument could be a register or an individual OTP fuse.
//
// An empty NVMEM device path is allowed to simulate the operation and test
// returned values.
//
// The value parameter is interpreted as a big-endian value, please note that
// certain tools, such as the ones creating the `SRK_HASH` for secure boot
// purposes, typically prepare their output in little-endian format.
func Blow(devicePath string, f *fusemap.FuseMap, name string, val []byte) (res []byte, addr uint32, off int, size int, err error) {
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
		size = reg.Length
	case *fusemap.Fuse:
		fuse := m
		addr = fuse.Register.WriteAddress
		off = fuse.Offset
		size = fuse.Length
	}

	res, err = util.ConvertWriteValue(off, size, val)

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
