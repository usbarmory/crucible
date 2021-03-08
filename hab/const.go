// crucible
// One-Time-Programmable (OTP) fusing tool
//
// Copyright (c) F-Secure Corporation
//
// Use of this source code is governed by the license
// that can be found in the LICENSE file.

package hab

// HAB_VER specifies HABv4
const HAB_VER = 0x40

// HABv4 key constants
// (p37, 4.3.7 Install Key, HABv4 API RM).
const (
	// Public Key types
	HAB_KEY_PUBLIC = 0xe1
	HAB_KEY_HASH   = 0xee

	// Public Key store indices
	HAB_IDX_SRK  = 0
	HAB_IDX_CSFK = 1

	// Flags for Install Key commands
	HAB_CMD_INS_KEY_CLR = 0
	HAB_CMD_INS_KEY_ABS = 1
	HAB_CMD_INS_KEY_CSF = 2
	HAB_CMD_INS_KEY_DAT = 4
	HAB_CMD_INS_KEY_CFG = 8
	HAB_CMD_INS_KEY_FID = 16
	HAB_CMD_INS_KEY_MID = 32
	HAB_CMD_INS_KEY_CID = 64
	HAB_CMD_INS_KEY_HSH = 128
)

// HABv4 tags
// (p61, 6.2 Structure, HABv4 API RM).
const (
	HAB_TAG_IVT = 0xd1
	HAB_TAG_DCD = 0xd2
	HAB_TAG_CSF = 0xd4
	HAB_TAG_CRT = 0xd7
	HAB_TAG_SIG = 0xd8
	HAB_TAG_EVT = 0xdb
	HAB_TAG_RVT = 0xdd
	HAB_TAG_WRP = 0x81
	HAB_TAG_MAC = 0xac
)

// HABv4 tags
// (p63, 6.5 Algorithms, HABv4 API RM).
const (
	// Algorithm types
	HAB_ALG_ANY    = 0x00
	HAB_ALG_HASH   = 0x01
	HAB_ALG_SIG    = 0x02
	HAB_ALG_F      = 0x03
	HAB_ALG_EC     = 0x04
	HAB_ALG_CIPHER = 0x05
	HAB_ALG_MODE   = 0x06
	HAB_ALG_WRAP   = 0x07

	// Hash algorithms
	HAB_ALG_SHA1   = 0x11
	HAB_ALG_SHA256 = 0x17
	HAB_ALG_SHA512 = 0x1b

	// Signature algorithms
	HAB_ALG_PKCS1 = 0x21

	// Cipher algorithms
	HAB_ALG_AES = 0x55

	// Cipher or hash modes
	HAB_MODE_CCM = 0x66

	// Key wrap algorithms
	HAB_ALG_BLOB = 0x71
)
