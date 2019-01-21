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
	"syscall"
)

func getSystemInformation() (kernel string, release string, err error) {
	var uname syscall.Utsname

	err = syscall.Uname(&uname)

	if err != nil {
		return
	}

	for _, c := range uname.Sysname {
		if c == 0 {
			break
		}

		kernel += string(byte(c))
	}

	for _, c := range uname.Release {
		if c == 0 {
			break
		}

		release += string(byte(c))
	}

	return
}

func driverParams(name string) (wordSize uint32, bankSize uint32, err error) {
	kernel, _, err := getSystemInformation()

	if err != nil {
		err = fmt.Errorf("could not get system information, %v", err)
		return
	}

	if kernel != "Linux" {
		err = errors.New("unsupported OS")
		return
	}

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
