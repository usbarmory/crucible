// crucible
// One-Time-Programmable (OTP) fusing tool
//
// Copyright (c) WithSecure Corporation
//
// Use of this source code is governed by the license
// that can be found in the LICENSE file.

package main

import (
	"bytes"
	"flag"
	"fmt"
	"io/fs"
	"log"
	"math/big"
	"path/filepath"
	"text/tabwriter"

	"github.com/usbarmory/crucible/fusemap"
	"github.com/usbarmory/crucible/otp"
)

func listFusemapRegisters(f *fusemap.FuseMap) {
	var res []byte

	for _, reg := range f.RegistersByWriteAddress() {
		if flag.Arg(0) == "read" {
			res, _, _, _, err := otp.ReadNVMEM(conf.device, f, reg.Name)

			if err != nil {
				log.Fatalf("error: could not read fusemap, %v", err)
			}

			n := new(big.Int)
			n.SetBytes(res)
		}

		fmt.Print(reg.BitMap(res))
		fmt.Println()
	}
}

func listFusemaps() {
	var list bytes.Buffer

	_, _ = fmt.Fprintf(&list, "Model (-m)\tReference (-r)\tDriver\n")

	t := tabwriter.NewWriter(&list, 16, 8, 0, '\t', tabwriter.TabIndent)

	_ = fs.WalkDir(conf.fusemapDir, ".", func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}

		if d.IsDir() {
			return nil
		}

		if filepath.Ext(path) != ".yaml" {
			return nil
		}

		y, err := fs.ReadFile(conf.fusemapDir, path)

		if err != nil {
			return err
		}

		f, err := fusemap.Parse(y)

		if err != nil {
			log.Printf("skipping %s (%v)", path, err)
			return nil
		}

		_, _ = fmt.Fprintf(t, "%s\t%s\t%s\n", f.Processor, f.Reference, f.Driver)

		return nil
	})

	_ = t.Flush()

	fmt.Print(list.String())
}
