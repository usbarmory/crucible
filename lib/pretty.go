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
	"fmt"
	"sort"
	"strings"
)

const regMap = " 31 30 29 28 27 26 25 24 23 22 21 20 19 18 17 16 15 14 13 12 11 10 09 08 07 06 05 04 03 02 01 00"
const topSep = "┏━━┳━━┳━━┳━━┳━━┳━━┳━━┳━━┳━━┳━━┳━━┳━━┳━━┳━━┳━━┳━━┳━━┳━━┳━━┳━━┳━━┳━━┳━━┳━━┳━━┳━━┳━━┳━━┳━━┳━━┳━━┳━━┓"
const bitSep = "┃"
const lowSep = "┗━━┻━━┻━━┻━━┻━━┻━━┻━━┻━━┻━━┻━━┻━━┻━━┻━━┻━━┻━━┻━━┻━━┻━━┻━━┻━━┻━━┻━━┻━━┻━━┻━━┻━━┻━━┻━━┻━━┻━━┻━━┻━━┛"
const octSep = "                      b3                      b2                      b1                      b0"

// Pretty print register bit map.
//
// The function operates on a single register, this means that fuses which
// start in other registers are not shown.
//
// "Alias" fuses which overlaps across each other result in an overlapping bit
// map, individual fuse description remains accurate.
func (reg *Register) BitMap() (m string) {
	if reg == nil {
		return
	}

	bitSepFixed := []byte("|")
	bitBox := append([]byte("  "), bitSepFixed...)

	// line which maps register bits to fuse names
	bitMap := bytes.Repeat(bitBox, len(regMap)/len(bitBox))

	// list which maps fuses to register bits
	var lines []string

	for fuseName, fuse := range reg.Fuses {
		size := fuse.Length

		// trim a fuse that falls outside the register
		if fuse.Offset+size > 32 {
			size = 32 - fuse.Offset
		}

		off := int(fuse.Offset) * len(bitBox)
		fillCount := int(size)*2 + int(size) - 1

		fill := []byte(fuseName)

		if len(fill) < fillCount {
			fill = append(fill, bytes.Repeat([]byte(" "), fillCount-len(fill))...)
		} else {
			fill = fill[0:fillCount]
		}

		copy(bitMap[off:], SwitchEndianess(fill))

		indent := len(regMap) - int(off) - fillCount
		line := strings.Repeat(" ", indent)

		for i := fuse.Offset + size; i > fuse.Offset; i-- {
			line += fmt.Sprintf("%.2d ", i-1)
		}

		if len(line) < len(regMap) {
			line += " "
			line += strings.Repeat("─", len(regMap)-len(line))
			line += " "
		}

		lines = append(lines, fmt.Sprintf("%s%s\n", line, fuseName))
	}

	bitMap = SwitchEndianess(bitMap)
	bitMap = bytes.Replace(bitMap, bitSepFixed, []byte(bitSep), -1)
	bitMap = append(bitMap, []byte(bitSep)...)

	sort.Strings(lines)

	m += fmt.Sprintf("%s  %s\n", regMap, reg.Name)
	m += fmt.Sprintf("%s Bank:%d Word:%d\n", topSep, reg.Bank, reg.Word)
	m += fmt.Sprintf("%s R: 0x%.8x\n", bitMap, reg.ReadAddress)
	m += fmt.Sprintf("%s W: 0x%.8x\n", lowSep, reg.WriteAddress)
	m += octSep + "\n"

	for i := 0; i < len(lines); i++ {
		m += fmt.Sprint(lines[len(lines)-1-i])
	}

	return
}
