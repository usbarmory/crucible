// crucible
// One-Time-Programmable (OTP) fusing tool
//
// Copyright (c) WithSecure Corporation
//
// Use of this source code is governed by the license
// that can be found in the LICENSE file.

package fusemap

import (
	"errors"
)

func (f *FuseMap) driverParams() (wordSize int, err error) {
	switch f.Driver {
	case "nvmem-imx-iim":
		wordSize = 1
	case "nvmem-imx-ocotp":
		wordSize = 4
	case "":
		err = errors.New("missing driver")
	default:
		err = errors.New("unsupported driver")
	}

	return
}
