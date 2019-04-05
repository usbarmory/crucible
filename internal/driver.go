// crucible
// One-Time-Programmable (OTP) fusing tool
//
// Copyright (c) F-Secure Corporation
//
// Use of this source code is governed by the license
// that can be found in the LICENSE file.
//
// +build linux

package crucible

import (
	"errors"
)

func driverParams(name string) (wordSize uint32, bankSize uint32, err error) {
	switch name {
	case "nvmem-imx-iim":
		wordSize = 1
		bankSize = 32
	case "nvmem-imx-ocotp":
		wordSize = 4
		bankSize = 8
	default:
		err = errors.New("unsupported driver")
	}

	return
}
