// crucible
// One-Time-Programmable (OTP) fusing tool
//
// Copyright (c) F-Secure Corporation
//
// Use of this source code is governed by the license
// that can be found in the LICENSE file.

package crucible

import (
	"testing"
)

func TestSort(t *testing.T) {
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
    word: 1
  REG3:
    bank: 0
    word: 2
...
`

	fusemap, err := ParseFuseMap([]byte(y))

	if err != nil {
		t.Fatal(err)
	}

	exp := []string{"REG1", "REG2", "REG3"}

	for i, reg := range fusemap.RegistersByReadAddress() {
		if exp[i] != reg.Name {
			t.Errorf("unexpected order, %s != %s", reg.Name, exp[i])
		}
	}

	for i, reg := range fusemap.RegistersByWriteAddress() {
		if exp[i] != reg.Name {
			t.Errorf("unexpected order, %s != %s", reg.Name, exp[i])
		}
	}
}
