// crucible
// One-Time-Programmable (OTP) fusing tool
//
// Copyright (c) F-Secure Corporation
//
// Use of this source code is governed by the license
// that can be found in the LICENSE file.

package hab

import (
	"bytes"
	"encoding/binary"
	"errors"
	"fmt"
)

// Header represents a data structure header
// (p22, 4.3 Command Sequence File, HABv4 API RM).
type Header struct {
	Tag uint8
	Len uint16
	Ver uint8
}

// Bytes converts the header to byte array format.
func (hdr *Header) Bytes() []byte {
	buf := new(bytes.Buffer)
	binary.Write(buf, binary.BigEndian, hdr)
	return buf.Bytes()
}

// NewHeader returns a tagged data structure header with default values.
func NewHeader(tag uint8) *Header {
	return &Header{
		Tag: tag,
		Len: 4,
		Ver: HAB_VER,
	}
}

// CSF represents a Command Sequence File
// (p23, 4.3 Command Sequence File, HABv4 API RM).
type CSF struct {
	Header *Header
	Data   []byte
}

// Set sets the CSF data.
func (csf *CSF) Set(buf []byte) {
	csf.Data = buf
	csf.Header.Len += uint16(len(buf))
}

// Read parses a buffer as Command Sequence File.
func (csf *CSF) Read(buf []byte) (err error) {
	if len(buf) < 4 {
		return errors.New("CSF header not found")
	}

	if csf.Header == nil {
		csf.Header = NewHeader(0)
	}

	err = binary.Read(bytes.NewReader(buf[0:4]), binary.BigEndian, csf.Header)

	if err != nil {
		return
	}

	csf.Data = buf[4:csf.Header.Len]

	return
}

// Bytes converts the CSF to byte array format.
func (csf *CSF) Bytes() []byte {
	buf := new(bytes.Buffer)

	buf.Write(csf.Header.Bytes())
	buf.Write(csf.Data)

	return buf.Bytes()
}

// NewCSF returns a tagged Command Sequence File with default values.
func NewCSF(tag uint8) *CSF {
	return &CSF{
		Header: NewHeader(tag),
	}
}

// IVT represents an Image Vector Table
// (p161, 8.7.1.1 Image vector table structure, IMX6ULLRM).
type IVT struct {
	Header    Header
	Entry     uint32 // Absolute address of the first instruction to execute from the image
	Reserved1 uint32 // Reserved and should be zero
	DCD       uint32 // Absolute address of the image DCD
	BootData  uint32 // Absolute address of the boot data
	Self      uint32 // Absolute address of the IVT
	CSF       uint32 // Absolute address of the Command Sequence File (CSF)
	Reserved2 uint32 // Reserved and should be zero
}

// Read parses a buffer as Image Vector Table.
func (ivt *IVT) Read(buf []byte) (err error) {
	err = binary.Read(bytes.NewReader(buf), binary.LittleEndian, ivt)

	if err != nil {
		return
	}

	err = binary.Read(bytes.NewReader(buf[0:4]), binary.BigEndian, &ivt.Header)

	if err != nil {
		return
	}

	if ivt.Header.Tag != HAB_TAG_IVT {
		return errors.New("IVT header not found")
	}

	return
}

// Bytes converts the IVT to byte array format.
func (ivt *IVT) Bytes() []byte {
	buf := new(bytes.Buffer)

	buf.Write(ivt.Header.Bytes())
	binary.Write(buf, binary.LittleEndian, ivt.Entry)
	binary.Write(buf, binary.LittleEndian, ivt.Reserved1)
	binary.Write(buf, binary.LittleEndian, ivt.DCD)
	binary.Write(buf, binary.LittleEndian, ivt.BootData)
	binary.Write(buf, binary.LittleEndian, ivt.Self)
	binary.Write(buf, binary.LittleEndian, ivt.CSF)
	binary.Write(buf, binary.LittleEndian, ivt.Reserved2)

	return buf.Bytes()
}

// BootData represents a Boot Data structure
// (p162, 8.7.1.2 Boot data structure, IMX6ULLRM).
type BootData struct {
	Start  uint32 // Absolute address of the image
	Length uint32 // Size of the program image
	Plugin uint32 // Plugin flag
}

// NewBootData returns the Boot Data included in the passed IMX image.
func NewBootData(imx []byte, ivt *IVT) (data *BootData, err error) {
	if ivt == nil {
		return nil, errors.New("invalid IVT header")
	}

	off := ivt.BootData - ivt.Self

	if int(off) > len(imx) {
		return nil, fmt.Errorf("invalid boot data offset (%d/%d)", off, len(imx))
	}

	data = &BootData{}
	err = binary.Read(bytes.NewReader(imx[off:]), binary.LittleEndian, data)

	return
}

// InstallKey represents an Install Key command
// (p33, 4.3.7 Install Key, HABv4 API RM).
type InstallKey struct {
	Tag    uint8
	Len    uint16
	Flg    uint8
	Pcl    uint8  // Key authentication protocol
	Alg    uint8  // Hash algorithm
	Src    uint8  // Source key index
	Tgt    uint8  // Target key index
	KeyDat uint32 // Start address of key data to install
}

// NewInstallKey returns an Install Key command with default values.
func NewInstallKey() *InstallKey {
	return &InstallKey{
		Tag: HAB_CMD_INS_KEY,
		Len: 12,
		Alg: HAB_ALG_ANY,
	}
}

// Bytes converts the Install Key command to byte array format.
func (cmd *InstallKey) Bytes() []byte {
	buf := new(bytes.Buffer)
	binary.Write(buf, binary.BigEndian, cmd)
	return buf.Bytes()
}

// AuthenticateData represents an Authenticate Data command
// (p39, 4.3.8 Authenticate Data, HABv4 API RM).
type AuthenticateData struct {
	Tag      uint8
	Len      uint16
	Flg      uint8
	Key      uint8  // Verification key index
	Pcl      uint8  // Authentication protocol
	Eng      uint8  // Engine used to process data blocks
	Cfg      uint8  // Engine configuration flags
	AutStart uint32 // Address of authentication data
	blk      DataBlock
}

// DataBlock represents an optional data block for Authenticate Data commands.
type DataBlock struct {
	Start uint32 // Absolute address of data block
	Bytes uint32 // Size in bytes of data block
}

// SetDataBlock adds a data block to the command.
func (cmd *AuthenticateData) SetDataBlock(start uint32, size uint32) {
	cmd.blk.Start = start
	cmd.blk.Bytes = size
	cmd.Len += 8
}

// Bytes converts the Authenticate Data command to byte array format.
func (cmd *AuthenticateData) Bytes() []byte {
	buf := new(bytes.Buffer)

	binary.Write(buf, binary.BigEndian, cmd)

	if cmd.Len <= 12 {
		buf.Truncate(int(cmd.Len))
	}

	return buf.Bytes()
}

// NewAuthenticateData returns an Authenticate Data command with default
// values.
func NewAuthenticateData() *AuthenticateData {
	return &AuthenticateData{
		Tag: HAB_CMD_AUT_DAT,
		Len: 12,
		Flg: HAB_CMD_AUT_DAT_CLR,
		// HAB_ENG_SW used due to ERR010449, if BT_MMU_DISABLE fuse is
		// blown then hardware engines can also be used.
		Eng: HAB_ENG_SW,
	}
}
