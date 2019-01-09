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
)

type FuseMap struct {
	Processor string               `json:"processor"`
	Reference string               `json:"reference"`
	Driver    string               `json:"driver"`
	Registers map[string]*Register `json:"registers"`
	Gaps      map[string]*Gap      `json:"gaps"`

	valid bool
}

type Gap struct {
	Read   bool   `json:"read"`
	Write  bool   `json:"write"`
	Length uint32 `json:"len"`
}

type Register struct {
	Name         string
	ReadAddress  uint32
	WriteAddress uint32
	Length       uint32
	Bank         uint32           `json:"bank"`
	Word         uint32           `json:"word"`
	Fuses        map[string]*Fuse `json:"fuses"`
}

type Fuse struct {
	Name         string
	Offset       uint32 `json:"offset"`
	Length       uint32 `json:"len"`
	Register     *Register
}

// Set register addressing.
func (fusemap *FuseMap) SetAddress(reg *Register) (err error) {
	if reg == nil {
		return
	}

	wordSize, bankSize, err := driverParams(fusemap.Driver)

	if err != nil {
		return
	}

	if reg.Word >= bankSize {
		return fmt.Errorf("register word cannot exceed %d", bankSize-1)
	}

	reg.ReadAddress = (reg.Bank*bankSize + reg.Word) * wordSize
	reg.WriteAddress = reg.ReadAddress

	return
}

// Apply gap information to register addressing.
func (fusemap *FuseMap) ApplyGaps() (err error) {
	wordSize, _, err := driverParams(fusemap.Driver)

	if err != nil {
		return
	}

	raddr := make(map[string]uint32)
	waddr := make(map[string]uint32)

	for name, reg := range fusemap.Registers {
		raddr[name] = reg.ReadAddress
		waddr[name] = reg.WriteAddress

		for gapRegName, gap := range fusemap.Gaps {
			if _, ok := fusemap.Registers[gapRegName]; !ok {
				return fmt.Errorf("invalid gap register (%s)", gapRegName)
			}

			gapReg := fusemap.Registers[gapRegName]

			if gap == nil {
				continue
			}

			if !gap.Read && !gap.Write {
				return errors.New("invalid gap, missing operation")
			}

			if gap.Length == 0 {
				return errors.New("invalid gap, missing length")
			}

			if gap.Read && reg.ReadAddress >= gapReg.ReadAddress {
				raddr[name] += gap.Length / wordSize
			}

			if gap.Write && reg.WriteAddress >= gapReg.WriteAddress {
				waddr[name] += gap.Length / wordSize
			}
		}
	}

	for name, reg := range fusemap.Registers {
		reg.ReadAddress = raddr[name]
		reg.WriteAddress = waddr[name]
	}

	return
}

// Validate a fuse map and populate address values.
func (fusemap *FuseMap) Validate() (err error) {
	names := make(map[string]bool)
	raddr := make(map[uint32]bool)
	waddr := make(map[uint32]bool)

	if fusemap.Reference == "" {
		return errors.New("missing reference")
	}

	if fusemap.Driver == "" {
		return errors.New("missing driver")
	}

	wordSize, _, err := driverParams(fusemap.Driver)

	if err != nil {
		return
	}

	for n1, reg := range fusemap.Registers {
		if _, ok := names[n1]; ok {
			return fmt.Errorf("register/fuse names must be unique, double entry for %s", n1)
		}
		names[n1] = true

		if reg == nil {
			continue
		}

		reg.Name = n1
		reg.Length = 8 * wordSize

		err = fusemap.SetAddress(reg)

		if err != nil {
			return
		}

		for n2, fuse := range reg.Fuses {
			if _, ok := names[n2]; ok {
				return fmt.Errorf("register/fuse names must be unique, double entry for %s", n2)
			}
			names[n2] = true

			if fuse == nil {
				continue
			}

			if fuse.Offset > 31 {
				return fmt.Errorf("fuse offset cannot exceed register length")
			}

			if fuse.Length > 512 {
				return fmt.Errorf("fuse length cannot exceed 512")
			}

			fuse.Name = n2
			fuse.Register = reg
		}
	}

	err = fusemap.ApplyGaps()

	if err != nil {
		return
	}

	for n1, reg := range fusemap.Registers {
		if reg == nil {
			continue
		}

		if _, ok := raddr[reg.ReadAddress]; ok {
			return fmt.Errorf("register read address must be unique, double entry for %d (%s)", reg.ReadAddress, n1)
		}
		raddr[reg.ReadAddress] = true

		if _, ok := waddr[reg.WriteAddress]; ok {
			return fmt.Errorf("register write address must be unique, double entry for %d (%s)", reg.WriteAddress, n1)
		}
		waddr[reg.WriteAddress] = true
	}

	fusemap.valid = true

	return
}

// Find a fusemap entry and return its corresponding Register or Fuse mapping.
func (fusemap *FuseMap) Find(name string) (mapping interface{}, err error) {
	if !fusemap.valid {
		err = errors.New("fusemap has not been validated yet")
		return
	}

	for n1, reg := range fusemap.Registers {
		if n1 == name {
			return reg, nil
		}

		if reg == nil {
			continue
		}

		for n2, otp := range reg.Fuses {
			if n2 == name {
				return otp, nil
			}
		}
	}

	err = fmt.Errorf("could not find any register/fuse named %s", name)

	return
}
