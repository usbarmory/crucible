// crucible
// One-Time-Programmable (OTP) fusing tool
//
// Copyright (c) The crucible authors. All Rights Reserved.
//
// Use of this source code is governed by the license
// that can be found in the LICENSE file.

package hab

import (
	"bytes"
	"crypto/rsa"
	"crypto/sha256"
	"encoding/binary"
	"errors"
	"math/big"
)

// PublicKey represents a Super Root Key publick key entry, suitable for SRK
// table generation.
type PublicKey struct {
	Tag1   uint8
	KeyLen uint16
	Tag2   uint8
	_      uint16
	Tag3   uint16
	ModLen uint16
	ExpLen uint16
	Mod    []byte
	Exp    []byte
}

// Set imports an RSA public key in a Super Root Key public key.
func (pk *PublicKey) Set(pubKey *rsa.PublicKey) (err error) {
	mod := pubKey.N.Bytes()
	exp := big.NewInt(int64(pubKey.E)).Bytes()

	if len(mod) > 0xffff {
		return errors.New("unexpected modulus size")
	}

	if len(exp) > 3 {
		return errors.New("unexpected exponent size")
	}

	pk.Mod = mod
	pk.Exp = exp

	pk.ModLen = uint16(len(pk.Mod))
	pk.ExpLen = uint16(len(pk.Exp))
	pk.KeyLen = uint16(len(pk.Bytes()))

	return
}

// Bytes converts the SRK public key to byte array format.
func (pk *PublicKey) Bytes() []byte {
	buf := new(bytes.Buffer)

	binary.Write(buf, binary.BigEndian, pk.Tag1)
	binary.Write(buf, binary.BigEndian, pk.KeyLen)
	binary.Write(buf, binary.BigEndian, pk.Tag2)
	binary.Write(buf, binary.BigEndian, uint16(0))
	binary.Write(buf, binary.BigEndian, pk.Tag3)
	binary.Write(buf, binary.BigEndian, pk.ModLen)
	binary.Write(buf, binary.BigEndian, pk.ExpLen)
	binary.Write(buf, binary.BigEndian, pk.Mod)
	binary.Write(buf, binary.BigEndian, pk.Exp)

	return buf.Bytes()
}

// Hash returns the Super Root Key public key hash (SHA256), suitable for SRK
// table generation.
func (pk *PublicKey) Hash() [32]byte {
	return sha256.Sum256(pk.Bytes())
}

// SRKTable represents a Super Root Key table entry, suitable for fusing hash
// generation. To apply appropriate defaults NewSRKTable() should be used to
// instantiate it.
type SRKTable struct {
	Tag uint8
	Len uint16
	Ver uint8
	SRK []PublicKey
}

// AddKey imports an RSA public key in a Super Root Key table.
func (table *SRKTable) AddKey(srk *rsa.PublicKey) (err error) {
	if len(table.SRK) > 4 {
		return errors.New("no more than 4 SRKs can be added")
	}

	pk := PublicKey{
		Tag1: HAB_KEY_PUBLIC,
		Tag2: HAB_ALG_PKCS1,
		Tag3: HAB_CMD_INS_KEY_HSH,
	}
	pk.Set(srk)

	table.SRK = append(table.SRK, pk)
	table.Len = uint16(len(table.Bytes()))

	return
}

// Bytes converts the SRK table to byte array format.
func (table *SRKTable) Bytes() []byte {
	buf := new(bytes.Buffer)

	binary.Write(buf, binary.BigEndian, table.Tag)
	binary.Write(buf, binary.BigEndian, table.Len)
	binary.Write(buf, binary.BigEndian, table.Ver)

	for _, pk := range table.SRK {
		buf.Write(pk.Bytes())
	}

	return buf.Bytes()
}

// Hash returns the Super Root Key public key hash (SHA256), suitable for OTP
// fusing.
func (table *SRKTable) Hash() [32]byte {
	var pkHashes []byte

	if len(table.SRK) == 0 || len(table.SRK) > 4 {
		panic("invalid table")
	}

	for _, pk := range table.SRK {
		hash := pk.Hash()
		pkHashes = append(pkHashes, hash[:]...)
	}

	return sha256.Sum256(pkHashes)
}

// NewSRKTable returns a Super Root Key table entry with appropriate defaults.
// The argument takes an array of Super Root Keys for addition, it can be set
// to nil (or an empty list) as keys can also be individually added with
// AddKey().
func NewSRKTable(srks []*rsa.PublicKey) (table *SRKTable, err error) {
	table = &SRKTable{
		Tag: HAB_TAG_CRT,
		Ver: HAB_VER,
	}

	for _, srk := range srks {
		err = table.AddKey(srk)

		if err != nil {
			break
		}
	}

	return
}
