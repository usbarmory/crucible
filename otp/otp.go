// crucible
// One-Time-Programmable (OTP) fusing tool
//
// Copyright (c) WithSecure Corporation
//
// Use of this source code is governed by the license
// that can be found in the LICENSE file.

// Package otp provides support for One-Time-Programmable (OTP) fuses read and
// write operations.
//
// WARNING: Fusing SoC OTPs is an **irreversible** action that permanently
// fuses values on the device. This means that any errors in the process, or
// lost fused data such as cryptographic key material, might result in a
// **bricked** device.
//
// The use of this package is therefore **at your own risk**.
package otp
