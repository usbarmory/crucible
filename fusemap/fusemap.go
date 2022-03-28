// crucible
// One-Time-Programmable (OTP) fusing tool
//
// Copyright (c) WithSecure Corporation
//
// Use of this source code is governed by the license
// that can be found in the LICENSE file.

// Package fusemap implements a register definition format to describe
// One-Time-Programmable (OTP) registers and fuses.
package fusemap

import (
	"errors"
	"fmt"
)

// FuseMap represents a collection of One-Time-Programmable (OTP) registers and
// fuses for a given processor.
type FuseMap struct {
	Processor string               `json:"processor"`
	Reference string               `json:"reference"`
	Driver    string               `json:"driver"`
	BankSize  int                  `json:"bank_size"`
	Registers map[string]*Register `json:"registers"`
	Gaps      map[string]*Gap      `json:"gaps"`

	WordSize  int

	valid bool
}

// Gap represents a gap definition to account for addressing gap between OTP
// banks.
type Gap struct {
	Read   bool `json:"read"`
	Write  bool `json:"write"`
	Length int  `json:"len"`
}

// Register represents an OTP register definition.
type Register struct {
	Name         string
	ReadAddress  uint32
	WriteAddress uint32
	Length       int
	Bank         int              `json:"bank"`
	Word         int              `json:"word"`
	Fuses        map[string]*Fuse `json:"fuses"`
}

// Fuse is an OTP fuse definition, representing one or more bits within a
// register.
type Fuse struct {
	Name     string
	Offset   int `json:"offset"`
	Length   int `json:"len"`
	Register *Register
}

// SetAddress sets register addressing.
func (f *FuseMap) SetAddress(reg *Register) (err error) {
	if reg == nil {
		return
	}

	if reg.Word >= f.BankSize {
		return fmt.Errorf("register word cannot exceed %d", f.BankSize-1)
	}

	reg.ReadAddress = uint32((reg.Bank*f.BankSize + reg.Word) * f.WordSize)
	reg.WriteAddress = reg.ReadAddress

	return
}

// ApplyGaps applies gap information to register addressing.
func (f *FuseMap) ApplyGaps() (err error) {
	raddr := make(map[string]uint32)
	waddr := make(map[string]uint32)

	for name, reg := range f.Registers {
		raddr[name] = reg.ReadAddress
		waddr[name] = reg.WriteAddress

		for gapRegName, gap := range f.Gaps {
			if _, ok := f.Registers[gapRegName]; !ok {
				return fmt.Errorf("invalid gap register (%s)", gapRegName)
			}

			gapReg := f.Registers[gapRegName]

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
				raddr[name] += uint32(gap.Length / f.WordSize)
			}

			if gap.Write && reg.WriteAddress >= gapReg.WriteAddress {
				waddr[name] += uint32(gap.Length / f.WordSize)
			}
		}
	}

	for name, reg := range f.Registers {
		reg.ReadAddress = raddr[name]
		reg.WriteAddress = waddr[name]
	}

	return
}

// Valid returns whether the fusemap passed validation, see Validate().
func (f *FuseMap) Valid() bool {
	return f.valid
}

// Validate a fusemap and populate address values.
func (f *FuseMap) Validate() (err error) {
	names := make(map[string]bool)
	raddr := make(map[uint32]bool)
	waddr := make(map[uint32]bool)

	if f.Reference == "" {
		return errors.New("missing reference")
	}

	f.WordSize, err = f.driverParams()

	if err != nil {
		return
	}

	if f.BankSize <= 0 {
		return errors.New("missing bank_size")
	}

	for n1, reg := range f.Registers {
		if _, ok := names[n1]; ok {
			return fmt.Errorf("register/fuse names must be unique, double entry for %s", n1)
		}
		names[n1] = true

		if reg == nil {
			continue
		}

		reg.Name = n1
		reg.Length = 8 * f.WordSize

		err = f.SetAddress(reg)

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

	err = f.ApplyGaps()

	if err != nil {
		return
	}

	for n1, reg := range f.Registers {
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

	f.valid = true

	return
}

// Find a fusemap entry and return its corresponding Register or Fuse mapping.
func (f *FuseMap) Find(name string) (mapping interface{}, err error) {
	if !f.valid {
		err = errors.New("fusemap has not been validated yet")
		return
	}

	for n1, reg := range f.Registers {
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
