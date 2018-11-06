// crucible
// One-Time-Programmable (OTP) fusing tool
//
// Copyright (c) F-Secure Corporation
//
// Use of this source code is governed by the license
// that can be found in the LICENSE file.

package crucible

import (
	"sort"
)

type regsByReadAddress []*Register
type regsByWriteAddress []*Register

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
