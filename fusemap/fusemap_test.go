// crucible
// One-Time-Programmable (OTP) fusing tool
//
// Copyright (c) WithSecure Corporation
//
// Use of this source code is governed by the license
// that can be found in the LICENSE file.

package fusemap

import (
	"os"
	"testing"
)

var fusemaps = os.DirFS("../fusemaps")

func TestInvalidReference(t *testing.T) {
	y := `
---
driver: nvmem-imx-ocotp
...
`

	_, err := Parse([]byte(y))

	if err == nil || err.Error() != "missing reference" {
		t.Error("fusemap with missing reference should raise an error")
	}

	_, err = Find(fusemaps, "IMX53", "1")

	if err == nil || err.Error() != "invalid reference" {
		t.Error("fusemap with invalid reference should raise an error")
	}
}

func TestInvalidDriver(t *testing.T) {
	y := `
---
reference: test
...
`

	_, err := Parse([]byte(y))

	if err == nil || err.Error() != "missing driver" {
		t.Error("fusemap with missing driver should raise an error")
	}

	y = `
---
reference: test
driver: invalid
registers:
  NAME1:
    bank: 0
    word: 0
...
`

	_, err = Parse([]byte(y))

	if err == nil || err.Error() != "unsupported driver" {
		t.Error("fusemap with unsupported driver should raise an error")
	}
}

func TestDuplicateRegisterName(t *testing.T) {
	y := `
---
reference: test
driver: nvmem-imx-ocotp
bank_size: 8
registers:
  NAME1:
    fuses:
      NAME2:
  NAME2:
...
`

	_, err := Parse([]byte(y))

	if err == nil || !(err.Error() == "register/fuse names must be unique, double entry for NAME1" || err.Error() == "register/fuse names must be unique, double entry for NAME2") {
		t.Error("fusemap with duplicate register name should raise an error")
	}
}

func TestDuplicateAddress(t *testing.T) {
	y := `
---
reference: test
driver: nvmem-imx-ocotp
registers:
  REG1:
    bank: 0
    word: 0
  REG2:
    bank: 0
    word: 0
...
`

	_, err := Parse([]byte(y))

	if err == nil || !(err.Error() != "register address must be unique, double entry for 0 (REG1)" || err.Error() != "register address must be unique, double entry for 0 (REG2)") {
		t.Error("fusemap with duplicate register address should raise an error")
	}
}

func TestInvalidWordIndices(t *testing.T) {
	y := `
---
reference: test
driver: nvmem-imx-ocotp
bank_size: 8
registers:
  REG1:
    bank: 0
    word: 8
...
`

	_, err := Parse([]byte(y))

	if err == nil || err.Error() != "register word cannot exceed 7" {
		t.Error("fusemap with excessive word index should raise an error")
	}

	y = `
---
reference: test
driver: nvmem-imx-iim
bank_size: 32
registers:
  REG1:
    bank: 0
    word: 32
...
`

	_, err = Parse([]byte(y))

	if err == nil || err.Error() != "register word cannot exceed 31" {
		t.Error("fusemap with excessive word index should raise an error")
	}
}

func TestInvalidFuseIndices(t *testing.T) {
	y := `
---
reference: test
driver: nvmem-imx-ocotp
bank_size: 8
registers:
  REG1:
    bank: 0
    word: 0
    fuses:
      OTP1:
        offset: 32
        len: 4

...
`

	_, err := Parse([]byte(y))

	if err == nil || err.Error() != "fuse offset cannot exceed register length" {
		t.Error("fusemap with excessive offset index should raise an error")
	}

	y = `
---
reference: test
driver: nvmem-imx-ocotp
bank_size: 8
registers:
  REG1:
    bank: 0
    word: 0
    fuses:
      OTP1:
        offset: 0
        len: 513

...
`

	_, err = Parse([]byte(y))

	if err == nil || err.Error() != "fuse length cannot exceed 512" {
		t.Error("fusemap with excessive word index should raise an error")
	}
}

func TestValidMap(t *testing.T) {
	y := `
---
reference: test
driver: nvmem-imx-ocotp
bank_size: 8
registers:
  REG1:
    bank: 0
    word: 0
  REG2:
    bank: 0
    word: 1
...
`

	_, err := Parse([]byte(y))

	if err != nil {
		t.Errorf("valid fusemap should not raise an error (%v)", err)
	}
}

func TestFind(t *testing.T) {
	y := `
---
reference: test
driver: nvmem-imx-ocotp
bank_size: 8
registers:
  REG1:
    fuses:
      OTP1:
        offset: 0
        len: 4
...
`

	f, err := Parse([]byte(y))

	if err != nil {
		t.Fatal(err)
	}

	_, err = f.Find("REG2")

	if err == nil {
		t.Error("fusemap search with missing register name should raise an error")
	}

	m, err := f.Find("REG1")

	if err != nil {
		t.Fatal(err)
	}

	switch m.(type) {
	case *Register:
	default:
		t.Error("fusemap search with register name should return a register mapping")
	}

	m, err = f.Find("OTP1")

	if err != nil {
		t.Fatal(err)
	}

	switch m.(type) {
	case *Fuse:
	default:
		t.Error("fusemap search with OTP name should return a fuse mapping")
	}
}

func TestInvalidGap(t *testing.T) {
	y := `
---
reference: test
driver: nvmem-imx-ocotp
bank_size: 8
gaps:
  REG3:
    read: true
    len: 0x100
registers:
  REG1:
    bank: 0
    word: 0
  REG2:
    bank: 0
    word: 1
...
`

	_, err := Parse([]byte(y))

	if err == nil || err.Error() != "invalid gap register (REG3)" {
		t.Error("fusemap with invalid gap register should raise an error")
	}

	y = `
---
reference: test
driver: nvmem-imx-ocotp
bank_size: 8
gaps:
  REG2:
    len: 0x100
registers:
  REG1:
    bank: 0
    word: 0
  REG2:
    bank: 0
    word: 1
...
`

	_, err = Parse([]byte(y))

	if err == nil || err.Error() != "invalid gap, missing operation" {
		t.Error("fusemap with invalid gap operation should raise an error")
	}

	y = `
---
reference: test
driver: nvmem-imx-ocotp
bank_size: 8
gaps:
  REG2:
    read: true
registers:
  REG1:
    bank: 0
    word: 0
  REG2:
    bank: 0
    word: 1
...
`

	_, err = Parse([]byte(y))

	if err == nil || err.Error() != "invalid gap, missing length" {
		t.Error("fusemap with invalid gap length should raise an error")
	}
}

func TestGap(t *testing.T) {
	y := `
---
reference: test
driver: nvmem-imx-ocotp
bank_size: 8
gaps:
  REG2:
    read: true
    len: 0x100
  REG3:
    read: true
    len: 0x40
registers:
  REG1:
    bank: 0
    word: 0
  REG2:
    bank: 0
    word: 1
  REG3:
    bank: 0
    word: 2
...
`

	f, err := Parse([]byte(y))

	if err != nil {
		t.Fatal(err)
	}

	gap := f.Registers["REG2"].ReadAddress - f.Registers["REG1"].ReadAddress
	expGap := 4 + uint32(0x100)/4

	if gap != expGap {
		t.Errorf("unexpected gap, %x != %x", gap, expGap)
	}

	gap2 := f.Registers["REG3"].ReadAddress - f.Registers["REG2"].ReadAddress
	expGap2 := 4 + uint32(0x40)/4

	if gap2 != expGap2 {
		t.Errorf("unexpected gap, %x != %x", gap2, expGap2)
	}
}
