// crucible
// One-Time-Programmable (OTP) fusing tool
//
// Copyright (c) WithSecure Corporation
//
// Use of this source code is governed by the license
// that can be found in the LICENSE file.

//go:build tamago && arm
// +build tamago,arm

package otp

import (
	"encoding/binary"
	"time"

	"github.com/usbarmory/crucible/util"
	"github.com/usbarmory/tamago/soc/nxp/imx6ul"
	"github.com/usbarmory/tamago/soc/nxp/ocotp"
)

func init() {
	imx6ul.OCOTP.Init()
}

// Blow an OTP fuse using the NXP On-Chip OTP Controller.
//
// WARNING: Fusing SoC OTPs is an **irreversible** action that permanently
// fuses values on the device. This means that any errors in the process, or
// lost fused data such as cryptographic key material, might result in a
// **bricked** device.
//
// The use of this function is therefore **at your own risk**.
func BlowOCOTP(bank int, word int, off int, bitLen int, val []byte) (err error) {
	if len(val) == 0 {
		return
	}

	val, err = util.ConvertWriteValue(off, bitLen, val)

	if err != nil {
		return
	}

	val = util.Pad4(val)

	// write one complete OTP word write at the time
	for i := 0; i < len(val); i += ocotp.WordSize {
		w := word + (i / ocotp.WordSize)
		v := binary.LittleEndian.Uint32(val[i : i+ocotp.WordSize])

		if err = imx6ul.OCOTP.Blow(bank, w, v); err != nil {
			return
		}
		time.Sleep(10 * time.Millisecond)
	}

	return
}

// Read an OTP fuse using the NXP On-Chip OTP Controller.
func ReadOCOTP(bank int, word int, off int, bitLen int) (res []byte, err error) {
	regSize := ocotp.WordSize * 8
	numRegisters := 1 + (off+bitLen)/regSize

	// normalize
	if (off+bitLen)%regSize == 0 {
		numRegisters -= 1
	}

	res = make([]byte, numRegisters*ocotp.WordSize)

	// read one complete OTP word write at the time
	for i := 0; i < len(res); i += ocotp.WordSize {
		w := word + (i / ocotp.WordSize)

		val, err := imx6ul.OCOTP.Read(bank, w)

		if err != nil {
			return nil, err
		}

		buf := make([]byte, 4)
		binary.LittleEndian.PutUint32(buf, val)
		copy(res[i:i+ocotp.WordSize], buf)
	}

	res = util.ConvertReadValue(off, bitLen, res)

	return
}
