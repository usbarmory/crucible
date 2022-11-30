// crucible
// One-Time-Programmable (OTP) fusing tool
//
// Copyright (c) WithSecure Corporation
//
// Use of this source code is governed by the license
// that can be found in the LICENSE file.

package fusemap

import (
	"bytes"
	"fmt"
	"sort"
	"strings"

	"github.com/usbarmory/crucible/util"
)

const bitSep = "┃"

// ┏━━┳━━┳━━┳━━┳━━┳━━┳━━┳━━┳━━┳ .. ┳━━┓
func genTopSep(size int) (s string) {
	s = "┏"

	for i := 1; i < size; i++ {
		s += "━━┳"
	}

	s += "━━┓"

	return
}

// ┗━━┻━━┻━━┻━━┻━━┻━━┻━━┻━━╋━━┻ .. ┻━━┛
func genLowSep(size int) (s string) {
	s = "┗"

	for i := 1; i < size; i++ {
		if i%8 == 0 {
			s += "━━╋"
		} else {
			s += "━━┻"
		}
	}

	s += "━━┛"

	return
}

// .. 07 06 05 04 03 02 01 00
func genRegMap(size int) (s string) {
	for i := 1; i <= size; i++ {
		s += fmt.Sprintf(" %.2d", size-i)
	}

	return
}

// Convert a byte array to a bit map compatible binary string representation
// (e.g. []byte{0x55, 0x55} => []byte("0  1  0  1  0  1  0  1  0  1  0  1  0  1 0  1")
func byteArrayToBitMap(val []byte, off int, size int) (m []byte) {
	m = []byte(fmt.Sprintf("%.8b", util.ConvertReadValue(off, size, util.SwitchEndianness(val))))
	m = bytes.Replace(m[1:len(m)-1], []byte(" "), nil, -1)
	m = bytes.Replace(m[len(m)-size:], []byte("0"), []byte("0  "), -1)
	m = bytes.Replace(m, []byte("1"), []byte("1  "), -1)

	return
}

// BitMap pretty prints a register bit map.
//
// The function operates on a single register, this means that fuses which
// start in other registers are not shown (to overcome this fusemaps can
// include fuse definitions to alias their register range).
//
// Additionally fuse definitions which overlap across each other (e.g.
// aliases) result in an overlapping bit map, individual fuse description
// remains accurate.
//
// An optional byte array can be passed to visualize read values, opposed to
// fuse names, within the bit map representation.
func (reg *Register) BitMap(res []byte) (m string) {
	if reg == nil {
		return
	}

	topSep := genTopSep(reg.Length)
	regMap := genRegMap(reg.Length)
	lowSep := genLowSep(reg.Length)

	// UTF-8 charlen (> 1 bytes) affects the math used to create the
	// bitMap, therefore we replace 8-bit chars with fancy UTF-8 ones only
	// as last step.
	bitSepFixed := []byte("|")
	bitBox := append([]byte("  "), bitSepFixed...)

	// line which maps register bits to fuse names
	bitMap := bytes.Repeat(bitBox, len(regMap)/len(bitBox))

	// list which maps fuses to register bits
	var lines []string

	if len(reg.Fuses) == 0 && res != nil {
		desc := byteArrayToBitMap(res, 0, reg.Length)
		copy(bitMap, util.SwitchEndianness(desc[0:len(desc)-1]))
	}

	for _, fuse := range reg.FusesByOffset() {
		var desc []byte
		size := fuse.Length

		// trim a fuse that falls outside the register
		if fuse.Offset+size > reg.Length {
			size = reg.Length - fuse.Offset
		}

		off := fuse.Offset * len(bitBox)
		descSize := size*2 + size - 1

		if res != nil {
			desc = byteArrayToBitMap(res, fuse.Offset, size)
		} else {
			desc = []byte(fuse.Name)
		}

		if len(desc) < descSize {
			desc = append(desc, bytes.Repeat([]byte(" "), descSize-len(desc))...)
		} else {
			desc = desc[0:descSize]
		}

		copy(bitMap[off:], util.SwitchEndianness(desc))

		if off > 0 {
			// Restore separator as it might have been overwritten
			// by a longer fuse alias displayed earlier (per
			// sorting rules).
			copy(bitMap[off-1:], bitSepFixed)
		}

		indent := len(regMap) - off - descSize

		// We track and increase line length separately to account for
		// UTF-8 charlen being > 1 bytes.
		line := strings.Repeat(" ", indent)
		lineLen := len(line)

		for i := fuse.Offset + size; i > fuse.Offset; i-- {
			if i == fuse.Offset+size || i == fuse.Offset+1 {
				bit := fmt.Sprintf("%.2d ", i-1)
				line += bit
				lineLen += len(bit)
			} else {
				// 5 bytes but we count them as 3
				line += "┄┄ "
				lineLen += 3
			}
		}

		if lineLen < len(regMap) {
			line += strings.Repeat("─", len(regMap)-lineLen)
			line += " "
		}

		lines = append(lines, fmt.Sprintf("%s %s\n", line, fuse.Name))
	}

	bitMap = util.SwitchEndianness(bitMap)
	bitMap = bytes.Replace(bitMap, bitSepFixed, []byte(bitSep), -1)
	bitMap = append(bitMap, []byte(bitSep)...)

	sort.Strings(lines)

	m += fmt.Sprintf("%s  %s\n", regMap, reg.Name)
	m += fmt.Sprintf("%s Bank:%d Word:%d\n", topSep, reg.Bank, reg.Word)
	m += fmt.Sprintf("%s R: 0x%.8x\n", bitMap, reg.ReadAddress)
	m += fmt.Sprintf("%s W: 0x%.8x\n", lowSep, reg.WriteAddress)

	for i := 0; i < len(lines); i++ {
		m += fmt.Sprint(lines[len(lines)-1-i])
	}

	return
}
