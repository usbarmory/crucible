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
	"fmt"
	"math/big"
	"math/bits"
	"os"
)

// Blow a fuse through Linux NVMEM subsystem framework, returns the input value
// converted as required for the fusing operation as well as the written
// address. The name argument could be a register or an individual OTP fuse.
//
// An empty NVMEM device path is allowed to simulate the operation and test
// returned values.
func (fusemap *FuseMap) Blow(devicePath string, name string, val []byte) (res []byte, addr uint32, off uint32, size uint32, err error) {
	if len(val) == 0 {
		err = errors.New("null value")
		return
	}

	if fusemap.Driver != "nvmem-imx-ocotp" {
		err = errors.New("driver does not support blow operation")
		return
	}

	mapping, err := fusemap.Find(name)

	if err != nil {
		return
	}

	switch mapping.(type) {
	case *Register:
		reg := mapping.(*Register)
		addr = reg.WriteAddress
		off = 0
		size = reg.Length
	case *Fuse:
		fuse := mapping.(*Fuse)
		addr = fuse.WriteAddress
		off = fuse.Offset
		size = fuse.Length
	}

	res, err = ConvertWriteValue(off, size, val)

	if err != nil {
		return
	}

	res = Pad4(res)

	if devicePath == "" {
		return
	}

	device, err := os.OpenFile(devicePath, os.O_WRONLY|os.O_EXCL|os.O_SYNC, 0600)

	if err != nil {
		return
	}

	_, err = device.Seek(int64(addr), 0)

	if err != nil {
		_ = device.Close()
		return
	}

	_, err = device.Write(res)
	_ = device.Close()

	return
}

// Convert value to be written, shifted accordingly to its register offset and
// size, to a little endian array of 32-bit registers.
func ConvertWriteValue(off uint32, size uint32, val []byte) (res []byte, err error) {
	bitLen := bits.Len(uint(val[0])) + (len(val)-1)*8

	if bitLen > int(size) {
		err = fmt.Errorf("value bit size %d exceeds %d", bitLen, size)
		return
	}

	v := new(big.Int)
	v.SetBytes(val)
	v.Lsh(v, uint(off))

	res = PadBigInt(v, size)
	// big-endian > little-endian
	res = SwitchEndianess(res)

	return
}
