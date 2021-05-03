// crucible
// One-Time-Programmable (OTP) fusing tool
//
// Copyright (c) F-Secure Corporation
//
// Use of this source code is governed by the license
// that can be found in the LICENSE file.

package fusemap

import (
	"testing"
)

func TestFuseBitMap8(t *testing.T) {
	y := `
---
reference: test
driver: nvmem-imx-iim
bank_size: 8
registers:
  REG1:
    bank: 0
    word: 0
    fuses:
      OTP1:
        offset: 0
        len: 1
      OTP2:
        offset: 1
        len: 1
      OTP3:
        offset: 2
        len: 2
      OTP4:
        offset: 4
        len: 4
...
`

	exp := ` 07 06 05 04 03 02 01 00  REG1
┏━━┳━━┳━━┳━━┳━━┳━━┳━━┳━━┓ Bank:0 Word:0
┃OTP4       ┃OTP3 ┃OT┃OT┃ R: 0x00000000
┗━━┻━━┻━━┻━━┻━━┻━━┻━━┻━━┛ W: 0x00000000
 07 ┄┄ ┄┄ 04 ───────────  OTP4
             03 02 ─────  OTP3
                   01 ──  OTP2
                      00  OTP1
`

	f, err := Parse([]byte(y))

	if err != nil {
		t.Fatal(err)
	}

	m := f.Registers["REG1"].BitMap(nil)

	if m != exp {
		t.Errorf("unexpected map\n%s\n  !=\n%s", m, exp)
	}
}

func TestFuseBitMap32(t *testing.T) {
	y := `
---
reference: test
driver: nvmem-imx-ocotp
bank_size: 8
registers:
  REG1:
    bank: 0
    word: 0
    fuses:
      OTP1:
        offset: 0
        len: 1
      OTP2:
        offset: 1
        len: 1
      OTP3:
        offset: 2
        len: 2
      OTP4:
        offset: 4
        len: 28
...
`

	exp := ` 31 30 29 28 27 26 25 24 23 22 21 20 19 18 17 16 15 14 13 12 11 10 09 08 07 06 05 04 03 02 01 00  REG1
┏━━┳━━┳━━┳━━┳━━┳━━┳━━┳━━┳━━┳━━┳━━┳━━┳━━┳━━┳━━┳━━┳━━┳━━┳━━┳━━┳━━┳━━┳━━┳━━┳━━┳━━┳━━┳━━┳━━┳━━┳━━┳━━┓ Bank:0 Word:0
┃OTP4                                                                               ┃OTP3 ┃OT┃OT┃ R: 0x00000000
┗━━┻━━┻━━┻━━┻━━┻━━┻━━┻━━╋━━┻━━┻━━┻━━┻━━┻━━┻━━┻━━╋━━┻━━┻━━┻━━┻━━┻━━┻━━┻━━╋━━┻━━┻━━┻━━┻━━┻━━┻━━┻━━┛ W: 0x00000000
 31 ┄┄ ┄┄ ┄┄ ┄┄ ┄┄ ┄┄ ┄┄ ┄┄ ┄┄ ┄┄ ┄┄ ┄┄ ┄┄ ┄┄ ┄┄ ┄┄ ┄┄ ┄┄ ┄┄ ┄┄ ┄┄ ┄┄ ┄┄ ┄┄ ┄┄ ┄┄ 04 ───────────  OTP4
                                                                                     03 02 ─────  OTP3
                                                                                           01 ──  OTP2
                                                                                              00  OTP1
`

	f, err := Parse([]byte(y))

	if err != nil {
		t.Fatal(err)
	}

	m := f.Registers["REG1"].BitMap(nil)

	if m != exp {
		t.Errorf("unexpected map\n%s\n  !=\n%s", m, exp)
	}

	exp = ` 31 30 29 28 27 26 25 24 23 22 21 20 19 18 17 16 15 14 13 12 11 10 09 08 07 06 05 04 03 02 01 00  OCOTP_LOCK
┏━━┳━━┳━━┳━━┳━━┳━━┳━━┳━━┳━━┳━━┳━━┳━━┳━━┳━━┳━━┳━━┳━━┳━━┳━━┳━━┳━━┳━━┳━━┳━━┳━━┳━━┳━━┳━━┳━━┳━━┳━━┳━━┓ Bank:0 Word:0
┃GP6_L┃GP8_L┃GP7_L┃PI┃GP┃GP┃MI┃RO┃OT┃ANALO┃OT┃SW┃GP┃SR┃GP2_L┃GP1_L┃MAC_A┃  ┃SJ┃MEM_T┃BOOT_┃TESTE┃ R: 0x00000000
┗━━┻━━┻━━┻━━┻━━┻━━┻━━┻━━╋━━┻━━┻━━┻━━┻━━┻━━┻━━┻━━╋━━┻━━┻━━┻━━┻━━┻━━┻━━┻━━╋━━┻━━┻━━┻━━┻━━┻━━┻━━┻━━┛ W: 0x00000000
 31 30 ─────────────────────────────────────────────────────────────────────────────────────────  GP6_LOCK
       29 28 ───────────────────────────────────────────────────────────────────────────────────  GP8_LOCK
             27 26 ─────────────────────────────────────────────────────────────────────────────  GP7_LOCK
                   25 ──────────────────────────────────────────────────────────────────────────  PIN_LOCK
                      24 ───────────────────────────────────────────────────────────────────────  GP5_LOCK
                         23 ────────────────────────────────────────────────────────────────────  GP4_LOCK
                            22 ─────────────────────────────────────────────────────────────────  MISC_CONF_LOCK
                               21 ──────────────────────────────────────────────────────────────  ROM_PATCH_LOCK
                                  20 ───────────────────────────────────────────────────────────  OTPMK_CRC_LOCK
                                     19 18 ─────────────────────────────────────────────────────  ANALOG_LOCK
                                           17 ──────────────────────────────────────────────────  OTPMK_LOCK
                                              16 ───────────────────────────────────────────────  SW_GP_LOCK
                                                 15 ────────────────────────────────────────────  GP3_LOCK
                                                    14 ─────────────────────────────────────────  SRK_LOCK
                                                       13 12 ───────────────────────────────────  GP2_LOCK
                                                             11 10 ─────────────────────────────  GP1_LOCK
                                                                   09 08 ───────────────────────  MAC_ADDR_LOCK
                                                                            06 ─────────────────  SJC_RESP_LOCK
                                                                               05 04 ───────────  MEM_TRIM_LOCK
                                                                                     03 02 ─────  BOOT_CFG_LOCK
                                                                                           01 00  TESTER_LOCK
`

	f, err = Find(fusemaps, "IMX6UL", "1")

	if err != nil {
		t.Fatal(err)
	}

	m = f.Registers["OCOTP_LOCK"].BitMap(nil)

	if m != exp {
		t.Errorf("unexpected map\n%s\n  !=\n%s", m, exp)
	}

	exp = ` 31 30 29 28 27 26 25 24 23 22 21 20 19 18 17 16 15 14 13 12 11 10 09 08 07 06 05 04 03 02 01 00  OCOTP_CFG1
┏━━┳━━┳━━┳━━┳━━┳━━┳━━┳━━┳━━┳━━┳━━┳━━┳━━┳━━┳━━┳━━┳━━┳━━┳━━┳━━┳━━┳━━┳━━┳━━┳━━┳━━┳━━┳━━┳━━┳━━┳━━┳━━┓ Bank:0 Word:2
┃DIE-X-CORDINATE        ┃DIE-Y-CORDINATE        ┃WAFER_NO      ┃LOT_NO_ENC[42:32]               ┃ R: 0x00000008
┗━━┻━━┻━━┻━━┻━━┻━━┻━━┻━━╋━━┻━━┻━━┻━━┻━━┻━━┻━━┻━━╋━━┻━━┻━━┻━━┻━━┻━━┻━━┻━━╋━━┻━━┻━━┻━━┻━━┻━━┻━━┻━━┛ W: 0x00000008
 31 ┄┄ ┄┄ ┄┄ ┄┄ ┄┄ ┄┄ ┄┄ ┄┄ ┄┄ ┄┄ ┄┄ ┄┄ ┄┄ ┄┄ ┄┄ ┄┄ ┄┄ ┄┄ ┄┄ ┄┄ ┄┄ ┄┄ ┄┄ ┄┄ ┄┄ ┄┄ ┄┄ ┄┄ ┄┄ ┄┄ 00  UNIQUE_ID[63:32]
 31 ┄┄ ┄┄ ┄┄ ┄┄ ┄┄ ┄┄ ┄┄ ┄┄ ┄┄ ┄┄ ┄┄ ┄┄ ┄┄ ┄┄ ┄┄ ┄┄ ┄┄ ┄┄ ┄┄ ┄┄ ┄┄ ┄┄ ┄┄ ┄┄ ┄┄ ┄┄ ┄┄ ┄┄ ┄┄ ┄┄ 00  SJC_CHALLENGE[63:32]
 31 ┄┄ ┄┄ ┄┄ ┄┄ ┄┄ ┄┄ 24 ───────────────────────────────────────────────────────────────────────  DIE-X-CORDINATE
                         23 ┄┄ ┄┄ ┄┄ ┄┄ ┄┄ ┄┄ 16 ───────────────────────────────────────────────  DIE-Y-CORDINATE
                                                 15 ┄┄ ┄┄ ┄┄ 11 ────────────────────────────────  WAFER_NO
                                                                10 ┄┄ ┄┄ ┄┄ ┄┄ ┄┄ ┄┄ ┄┄ ┄┄ ┄┄ 00  LOT_NO_ENC[42:32]
`

	m = f.Registers["OCOTP_CFG1"].BitMap(nil)

	if m != exp {
		t.Errorf("unexpected map\n%s\n  !=\n%s", m, exp)
	}
}
