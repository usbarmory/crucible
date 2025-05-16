// crucible
// One-Time-Programmable (OTP) fusing tool
//
// Copyright (c) WithSecure Corporation
//
// Use of this source code is governed by the license
// that can be found in the LICENSE file.

package fusemap

import (
	"testing"
)

func TestInvalidOverlay(t *testing.T) {
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
        offset: 0
        len: 4
...
`

	f, err := Parse([]byte(y))

	if err != nil {
		t.Fatal(err)
	}

	y = `
---
reference: test
driver: nvmem-imx-ocotp
bank_size: 8
registers:
  REG2:
    bank: 0
    word: 0
    fuses:
      OTP1A:
        offset: 2
        len: 2
...
`

	v, err := Parse([]byte(y))

	if err != nil {
		t.Fatal(err)
	}

	err = f.Overlay(v)

	if err == nil {
		t.Error("fusemap overlay with invalid register should raise an error")
	}

	y = `
---
reference: invalid
driver: nvmem-imx-ocotp
bank_size: 8
registers:
  REG1:
    bank: 0
    word: 0
    fuses:
      OTP1A:
        offset: 2
        len: 2
...
`

	v, err = Parse([]byte(y))

	if err != nil {
		t.Fatal(err)
	}

	err = f.Overlay(v)

	if err == nil {
		t.Error("fusemap overlay with invalid register should raise an error")
	}

	y = `
---
reference: invalid
driver: nvmem-imx-ocotp
bank_size: 8
registers:
  REG1:
    bank: 0
    word: 0
    fuses:
      OTP1:
        offset: 2
        len: 2
...
`

	v, err = Parse([]byte(y))

	if err != nil {
		t.Fatal(err)
	}

	err = f.Overlay(v)

	if err == nil {
		t.Error("fusemap overlay with duplicate fuse should raise an error")
	}
}

func TestValidOverlay(t *testing.T) {
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
        offset: 0
        len: 4
...
`

	f, err := Parse([]byte(y))

	if err != nil {
		t.Fatal(err)
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
      OTP1A:
        offset: 2
        len: 2
...
`

	v, err := Parse([]byte(y))

	if err != nil {
		t.Fatal(err)
	}

	err = f.Overlay(v)

	if err != nil {
		t.Error("fusemap overlay with valid fuse should not raise an error")
	}
}
