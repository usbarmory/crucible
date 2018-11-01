// crucible
// One-Time-Programmable (OTP) fusing tool
//
// Copyright (c) F-Secure Corporation
//
// Use of this source code is governed by the license
// that can be found in the LICENSE file.

package crucible

import (
	"fmt"
	"io/ioutil"
	"path"

	"github.com/ghodss/yaml"
)

// Open a fuse map YAML file, validate it and convert it to a FuseMap
// structure.
func OpenFuseMap(dir string, processor string, reference string) (fusemap FuseMap, err error) {
	path := path.Join(dir, processor+".yaml")

	y, err := ioutil.ReadFile(path)

	if err != nil {
		return
	}

	fusemap, err = ParseFuseMap(y)

	if err != nil {
		return
	}

	if processor != fusemap.Processor {
		err = fmt.Errorf("fusemap file name must match its processor parameter (%s != %s)",
			processor, fusemap.Processor)
	}

	if reference != fusemap.Reference {
		err = fmt.Errorf("invalid reference")
	}

	return
}

// Parse a fuse map YAML file, validate it and convert it to a FuseMap
// structure.
func ParseFuseMap(y []byte) (fusemap FuseMap, err error) {
	fusemap = FuseMap{}
	err = yaml.Unmarshal(y, &fusemap)

	if err != nil {
		return
	}

	err = fusemap.Validate()

	return
}
