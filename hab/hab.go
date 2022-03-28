// crucible
// One-Time-Programmable (OTP) fusing tool
//
// Copyright (c) WithSecure Corporation
//
// Use of this source code is governed by the license
// that can be found in the LICENSE file.

// Package hab provides support functions for NXP HABv4 Secure Boot
// provisioning and executable signing.
//
// WARNING: Enabling NXP HABv4 secure boot is an irreversible action that
// permanently fuses verification keys hashes on the device. Any errors in the
// process or loss of the signing PKI will result in a bricked device incapable
// of executing unsigned code. This is a security feature, not a bug.
//
// The use of this package is therefore **at your own risk**.
package hab

import (
	"crypto/rsa"
	"fmt"
)

// SignOptions describes options for HABv4 executable image signing.
type SignOptions struct {
	// CSFKeyPEMBlock specifies the Command Sequence File signing key in
	// PEM format.
	CSFKeyPEMBlock []byte

	// CSFCertPEMBlock specifies the Command Sequence File signing
	// certificate in PEM format.
	CSFCertPEMBlock []byte

	// IMGKeyPEMBlock specifies the IMX executable signing key in PEM
	// format.
	IMGKeyPEMBlock []byte

	// IMGCertPEMBlock specifies the IMX executable signing certificate in
	// PEM format.
	IMGCertPEMBlock []byte

	// Table specifies the Super Root Key (SRK) table.
	Table []byte
	// Index specifies the SRK key index.
	Index int
	// Engine specifies the crypto engine.
	Engine int

	// SDP specifies whether the signed image is meant for Serial Download
	// Protocol execution.
	SDP bool
	DCD uint32
}

// Sign generates an HABv4 compliant authentication and configuration script
// (CSF) based on passed keys and against a target IMX executable image. The
// resulting CSF can be concatenated to the IMX image and used as signed
// bootable payload on NXP HABv4 supported processors.
func Sign(imx []byte, opts SignOptions) (out []byte, err error) {
	var csfKey, imgKey *rsa.PrivateKey
	var csfDER, imgDER []byte
	var sig []byte

	ivt := &IVT{}
	dcd := &DCD{}

	if err = ivt.Read(imx); err != nil {
		return
	}

	bootData, err := NewBootData(imx, ivt)

	if err != nil {
		return
	}

	if csfKey, err = parseKey(opts.CSFKeyPEMBlock); err != nil {
		return
	}

	if _, csfDER, err = parseCert(opts.CSFCertPEMBlock); err != nil {
		return
	}

	if imgKey, err = parseKey(opts.IMGKeyPEMBlock); err != nil {
		return
	}

	if _, imgDER, err = parseCert(opts.IMGCertPEMBlock); err != nil {
		return
	}

	// [Header]
	csf := &CSF{
		Header: NewHeader(HAB_TAG_CSF),
	}

	// [Install SRK]
	installSRK := NewInstallKey()
	installSRK.Flg = HAB_CMD_INS_KEY_CLR
	installSRK.Pcl = HAB_PCL_SRK
	installSRK.Alg = HAB_ALG_SHA256
	installSRK.Src = uint8(opts.Index) - 1
	csf.Header.Len += installSRK.Len

	// [Install CSFK]
	installCSFK := NewInstallKey()
	installCSFK.Flg = HAB_CMD_INS_KEY_CSF
	installCSFK.Pcl = HAB_PCL_X509
	installCSFK.Tgt = 1
	csf.Header.Len += installCSFK.Len

	// [Authenticate CSF]
	authenticateCSF := NewAuthenticateData()
	authenticateCSF.Key = 1
	authenticateCSF.Pcl = HAB_PCL_CMS
	authenticateCSF.Eng = uint8(opts.Engine)
	csf.Header.Len += authenticateCSF.Len

	// [Install Key]
	installIMG := NewInstallKey()
	installIMG.Pcl = HAB_PCL_X509
	installIMG.Tgt = 2
	csf.Header.Len += installIMG.Len

	// [Authenticate Data]
	authenticateData := NewAuthenticateData()
	authenticateData.Key = 2
	authenticateData.Pcl = HAB_PCL_CMS
	authenticateData.Eng = uint8(opts.Engine)
	authenticateData.SetDataBlock(ivt.Self, uint32(len(imx)))
	csf.Header.Len += authenticateData.Len

	// [Authenticate Data] for DCD block
	authenticateDCD := NewAuthenticateData()

	if opts.SDP {
		dcdStart := ivt.DCD - ivt.Self

		if err = dcd.Read(imx[dcdStart:]); err != nil {
			return nil, err
		}

		authenticateDCD.Key = 2
		authenticateDCD.Pcl = HAB_PCL_CMS
		authenticateDCD.Eng = uint8(opts.Engine)
		authenticateDCD.SetDataBlock(opts.DCD, uint32(dcd.Header.Len))
		csf.Header.Len += authenticateDCD.Len

		// Clear DCD pointer in IVT since Serial download mode will do
		// the same on loading the image.
		ivt.DCD = 0

		// make a copy to leave input slice untouched
		imx = append(ivt.Bytes(), imx[len(ivt.Bytes()):]...)
	}

	// Prepare CSF body

	csfCrt := NewCSF(HAB_TAG_CRT)
	csfCrt.Set(padCert(csfDER))

	imgCrt := NewCSF(HAB_TAG_CRT)
	imgCrt.Set(padCert(imgDER))

	// Sign IMX executable image
	if sig, err = sign(imx, opts.IMGCertPEMBlock, imgKey); err != nil {
		return nil, err
	}

	imgSig := NewCSF(HAB_TAG_SIG)
	imgSig.Set(padCert(sig))

	dcdSig := NewCSF(HAB_TAG_SIG)

	if opts.SDP {
		// Sign DCD
		if sig, err = sign(dcd.Bytes(), opts.IMGCertPEMBlock, imgKey); err != nil {
			return nil, err
		}

		dcdSig.Set(padCert(sig))
	}

	installSRK.KeyDat = uint32(csf.Header.Len)
	csf.Data = append(csf.Data, installSRK.Bytes()...)

	installCSFK.KeyDat = uint32(csf.Header.Len) + uint32(len(opts.Table))
	csf.Data = append(csf.Data, installCSFK.Bytes()...)

	authenticateCSF.AutStart = installCSFK.KeyDat + uint32(csfCrt.Header.Len)
	csf.Data = append(csf.Data, authenticateCSF.Bytes()...)

	installIMG.KeyDat = authenticateCSF.AutStart + uint32(imgSig.Header.Len)
	csf.Data = append(csf.Data, installIMG.Bytes()...)

	authenticateData.AutStart = installIMG.KeyDat + uint32(imgCrt.Header.Len)
	csf.Data = append(csf.Data, authenticateData.Bytes()...)

	if opts.SDP {
		authenticateDCD.AutStart = authenticateData.AutStart + uint32(imgSig.Header.Len)
		csf.Data = append(csf.Data, authenticateDCD.Bytes()...)
	}

	// Sign CSF commands
	if sig, err = sign(csf.Bytes(), opts.CSFCertPEMBlock, csfKey); err != nil {
		return nil, err
	}

	csfSig := &CSF{
		Header: NewHeader(HAB_TAG_SIG),
	}
	csfSig.Set(padCert(sig))

	csf.Data = append(csf.Data, opts.Table...)
	csf.Data = append(csf.Data, csfCrt.Bytes()...)
	csf.Data = append(csf.Data, csfSig.Bytes()...)
	csf.Data = append(csf.Data, imgCrt.Bytes()...)
	csf.Data = append(csf.Data, imgSig.Bytes()...)

	if opts.SDP {
		csf.Data = append(csf.Data, dcdSig.Bytes()...)
	}

	out = csf.Bytes()
	csfPadTo := int(bootData.Length) - len(imx) - IVT_OFFSET

	if len(out) > csfPadTo {
		return nil, fmt.Errorf("unexpected CSF length (%d > %d)", len(out), csfPadTo)
	} else if len(out) < csfPadTo {
		out = append(out, make([]byte, (csfPadTo-len(out)))...)
	}

	return
}
