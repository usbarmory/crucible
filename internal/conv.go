// crucible
// One-Time-Programmable (OTP) fusing tool
//
// Copyright (c) F-Secure Corporation
//
// Use of this source code is governed by the license
// that can be found in the LICENSE file.

package crucible

import (
	"math/big"
)

// Pad a byte array to ensure that it always represents one or more 4-byte
// register.
func Pad4(val []byte) (res []byte) {
	numRegisters := 1 + len(val)/4

	// normalize
	if len(val)%4 == 0 {
		numRegisters -= 1
	}

	pad := numRegisters*4 - len(val)

	for i := 0; i < pad; i++ {
		val = append(val, 0x00)
	}

	return val
}

// Pad a bit.Int value to account for the fact that big.Bytes() returns the
// absolute value, therefore leading 0x00 bytes are not returned and 0x00
// values are empty.
func PadBigInt(val *big.Int, size uint32) (res []byte) {
	numBytes := 1 + int(size/8)

	// normalize
	if size%8 == 0 {
		numBytes -= 1
	}

	pad := numBytes - len(val.Bytes())
	res = val.Bytes()

	for i := 0; i < pad; i++ {
		res = append([]byte{0x00}, res...)
	}

	return
}

// Reverse a byte array to switch between big <> little endianess.
func SwitchEndianness(val []byte) []byte {
	for i := len(val)/2 - 1; i >= 0; i-- {
		rev := len(val) - 1 - i
		val[i], val[rev] = val[rev], val[i]
	}

	return val
}
