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

// IVT_OFFSET represents the image vector table offset on disk images
const IVT_OFFSET = 1024

// DCD represents the default DCD location in OCRAM for Serial Download Mode
const DCD = 0x00910000

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
	HAB_CMD_INS_KEY_CLR = 0   // No flags set
	HAB_CMD_INS_KEY_ABS = 1   // Absolute certificate address
	HAB_CMD_INS_KEY_CSF = 2   // Install CSF key
	HAB_CMD_INS_KEY_DAT = 4   // Key binds to Data Type
	HAB_CMD_INS_KEY_CFG = 8   // Key binds to Configuration
	HAB_CMD_INS_KEY_FID = 16  // Key binds to Fabrication UID
	HAB_CMD_INS_KEY_MID = 32  // Key binds to Manufacturing ID
	HAB_CMD_INS_KEY_CID = 64  // Key binds to Caller ID
	HAB_CMD_INS_KEY_HSH = 128 // Certificate hash present
)

// HABv4 flags
// (p62, 4.3.8 Authenticate Data, HABv4 API RM).
const (
	HAB_CMD_AUT_DAT_CLR = 0x00 // No flags set
	HAB_CMD_AUT_DAT_ABS = 0x01 // Absolute signature address
)

// HABv4 data structures
// (p61, 6.2 Structure, HABv4 API RM).
const (
	HAB_TAG_IVT = 0xd1 // Image Vector Table
	HAB_TAG_DCD = 0xd2 // Device Configuration Data
	HAB_TAG_CSF = 0xd4 // Command Sequence File
	HAB_TAG_CRT = 0xd7 // Certificate
	HAB_TAG_SIG = 0xd8 // Signature
	HAB_TAG_EVT = 0xdb // Event
	HAB_TAG_RVT = 0xdd // ROM Vector Table
	HAB_TAG_WRP = 0x81 // Wrapped Key
	HAB_TAG_MAC = 0xac // Message Authentication Code
)

// HABv4 commands
// (p62, 6.3 Commands, HABv4 API RM).
const (
	HAB_CMD_SET     = 0xb1 // Set
	HAB_CMD_INS_KEY = 0xbe // Install Key
	HAB_CMD_AUT_DAT = 0xca // Authenticate Data
	HAB_CMD_WRT_DAT = 0xcc // Write Data
	HAB_CMD_CHK_DAT = 0xcf // Check Data
	HAB_CMD_NOP     = 0xc0 // No Operation
	HAB_CMD_INIT    = 0xb4 // Initialize
	HAB_CMD_UNLK    = 0xb2 // Unlock
)

// HABv4 protocol
// (p62, 6.4 Protocol, HABv4 API RM).
const (
	HAB_PCL_SRK  = 0x03 // SRK certificate format
	HAB_PCL_X509 = 0x09 // X.509v3 certificate format
	HAB_PCL_CMS  = 0xc5 // CMS/PKCS#7 signature format
	HAB_PCL_BLOB = 0xbb // SHW-specific wrapped key format
	HAB_PCL_AEAD = 0xa3 // Proprietary AEAD MAC format
)

// HABv4 tags
// (p63, 6.5 Algorithms, HABv4 API RM).
const (
	// Algorithm types
	HAB_ALG_ANY    = 0x00 // Algorithm type ANY
	HAB_ALG_HASH   = 0x01 // Hash algorithm type
	HAB_ALG_SIG    = 0x02 // Signature algorithm type
	HAB_ALG_F      = 0x03 // Finite field arithmetic
	HAB_ALG_EC     = 0x04 // Elliptic curve arithmetic
	HAB_ALG_CIPHER = 0x05 // Cipher algorithm type
	HAB_ALG_MODE   = 0x06 // Cipher/Hash modes
	HAB_ALG_WRAP   = 0x07 // Key wrap algorithm type

	// Hash algorithms
	HAB_ALG_SHA1   = 0x11 // SHA-1 algorithm ID
	HAB_ALG_SHA256 = 0x17 // SHA-256 algorithm ID
	HAB_ALG_SHA512 = 0x1b // SHA-512 algorithm ID

	// Signature algorithms
	HAB_ALG_PKCS1 = 0x21 // PKCS#1 RSA signature algorithm

	// Cipher algorithms
	HAB_ALG_AES = 0x55 // AES algorithm ID

	// Cipher or hash modes
	HAB_MODE_CCM = 0x66 // Counter with CBC-MAC

	// Key wrap algorithms
	HAB_ALG_BLOB = 0x71 // SHW-specific key wrap
)

// HABv4 engines
// (p64, 6.5 Engine, HABv4 API RM).
const (
	HAB_ENG_ANY    = 0x00 // First compatible engine
	HAB_ENG_SCC    = 0x03 // Security controller
	HAB_ENG_RTIC   = 0x05 // Run-time integrity checker
	HAB_ENG_SAHARA = 0x06 // Crypto accelerator
	HAB_ENG_CSU    = 0x0a // Central Security Unit
	HAB_ENG_SRTC   = 0x0c // Secure clock
	HAB_ENG_DCP    = 0x1b // Data Co-Processor
	HAB_ENG_CAAM   = 0x1d // Cryptographic Acceleration and Assurance Module
	HAB_ENG_SNVS   = 0x1e // Secure Non-Volatile Storage
	HAB_ENG_OCOTP  = 0x21 // Fuse controller
	HAB_ENG_DTCP   = 0x22 //DTCP co-processor
	HAB_ENG_ROM    = 0x36 // Protected ROM area
	HAB_ENG_HDCP   = 0x24 // HDCP co-processor
	HAB_ENG_SW     = 0xff // Software engine
)
