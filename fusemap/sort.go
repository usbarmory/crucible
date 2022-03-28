// crucible
// One-Time-Programmable (OTP) fusing tool
//
// Copyright (c) WithSecure Corporation
//
// Use of this source code is governed by the license
// that can be found in the LICENSE file.

package fusemap

import (
	"sort"
)

type regsByReadAddress []*Register
type regsByWriteAddress []*Register
type fusesByOffset []*Fuse

func (regs regsByReadAddress) Len() int {
	return len(regs)
}

func (regs regsByReadAddress) Swap(i, j int) {
	regs[i], regs[j] = regs[j], regs[i]
}

func (regs regsByReadAddress) Less(i, j int) bool {
	return regs[i].ReadAddress < regs[j].ReadAddress
}

// Return fusemap registers sorted by read address.
func (fusemap *FuseMap) RegistersByReadAddress() (regs regsByReadAddress) {
	for _, reg := range fusemap.Registers {
		regs = append(regs, reg)
	}

	sort.Sort(regs)

	return
}

func (regs regsByWriteAddress) Len() int {
	return len(regs)
}

func (regs regsByWriteAddress) Swap(i, j int) {
	regs[i], regs[j] = regs[j], regs[i]
}

func (regs regsByWriteAddress) Less(i, j int) bool {
	return regs[i].WriteAddress < regs[j].WriteAddress
}

// Return fusemap registers sorted by write address.
func (fusemap *FuseMap) RegistersByWriteAddress() (regs regsByWriteAddress) {
	for _, reg := range fusemap.Registers {
		regs = append(regs, reg)
	}

	sort.Sort(regs)

	return
}

func (fuses fusesByOffset) Len() int {
	return len(fuses)
}

func (fuses fusesByOffset) Swap(i, j int) {
	fuses[i], fuses[j] = fuses[j], fuses[i]
}

func (fuses fusesByOffset) Less(i, j int) bool {
	if fuses[i].Offset == fuses[j].Offset {
		// return longer fuse aliases first
		return (fuses[i].Length > fuses[j].Length)
	}

	return (fuses[i].Offset < fuses[j].Offset)
}

// Return register fuses sorted by offset.
func (reg *Register) FusesByOffset() (fuses fusesByOffset) {
	for _, fuse := range reg.Fuses {
		fuses = append(fuses, fuse)
	}

	sort.Sort(fuses)

	return
}
