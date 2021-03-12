One-Time-Programmable (OTP) fusing tool
=======================================

crucible | https://github.com/f-secure-foundry/crucible  
Copyright (c) F-Secure Corporation

```
 ▄████▄   ██▀███   █    ██  ▄████▄   ██▓ ▄▄▄▄    ██▓    ▓█████
▒██▀ ▀█  ▓██ ▒ ██▒ ██  ▓██▒▒██▀ ▀█  ▓██▒▓█████▄ ▓██▒    ▓█   ▀
▒▓█    ▄ ▓██ ░▄█ ▒▓██  ▒██░▒▓█    ▄ ▒██▒▒██▒ ▄██▒██░    ▒███
▒▓▓▄ ▄██▒▒██▀▀█▄  ▓▓█  ░██░▒▓▓▄ ▄██▒░██░▒██░█▀  ▒██░    ▒▓█  ▄
▒ ▓███▀ ░░██▓ ▒██▒▒▒█████▓ ▒ ▓███▀ ░░██░░▓█  ▀█▓░██████▒░▒████▒
░ ░▒ ▒  ░░ ▒▓ ░▒▓░░▒▓▒ ▒ ▒ ░ ░▒ ▒  ░░▓  ░▒▓███▀▒░ ▒░▓  ░░░ ▒░ ░
  ░  ▒     ░▒ ░ ▒░░░▒░ ░ ░   ░  ▒    ▒ ░▒░▒   ░ ░ ░ ▒  ░ ░ ░  ░
░          ░░   ░  ░░░ ░ ░ ░         ▒ ░ ░    ░   ░ ░      ░
░ ░         ░        ░     ░ ░       ░   ░          ░  ░   ░  ░
░                          ░                  ░

                  Where SoCs meet their fate.
```

Authors
=======

Andrea Barisani  
andrea.barisani@f-secure.com | andrea@inversepath.com  

Introduction
============

The `crucible` utility provides user space support for reading, and writing,
One-Time-Programmable (OTP) fuses of System-on-Chip (SoC) application
processors through the [Linux NVMEM framework](https://github.com/torvalds/linux/blob/master/Documentation/nvmem/nvmem.txt).

The current support targets application processors from the NXP i.MX series
(see _Supported drivers_).

Warning
=======

Fusing SoC OTPs is an **irreversible** action that permanently fuses values on
the device. This means that any errors in the process, or lost fused data such
as cryptographic key material, might result in a **bricked** device.

The use of this tool is therefore **at your own risk**.

Installing
==========

Pre-compiled binaries for Linux are released
[here](https://github.com/f-secure-foundry/crucible/releases).

You can also automatically download, compile and install the package, under
your GOPATH, as follows:

```
go install github.com/f-secure-foundry/crucible/cmd/crucible@latest
```

Alternatively you can manually compile it from source:

```
git clone https://github.com/f-secure-foundry/crucible
cd crucible && make
```

All targets can be cross compiled for ARM as follows:

```
make GOARCH=arm
```

Operation
=========

```
Usage: crucible [options] [read|blow] [fuse/register name] [value]
  -Y	do not prompt for confirmation (DANGEROUS)
  -b int
    	value base/format (2,10,16)
  -e string
    	value endianness (big,little)
  -f string
    	YAML fusemaps directory (default "fusemaps")
  -l	list fusemaps
    	visualize fusemap registers (with -m and -r)
    	visualize read value (with read operation on a register)
  -m string
    	processor model
  -n string
    	NVMEM device (default "/sys/bus/nvmem/devices/imx-ocotp0/nvmem")
  -r string
    	reference manual revision
  -s	use syslog, print only result value to stdout
```

The `-b` option controls value argument base/format and must be explicitly set
for all operations. For instance binary values like 0b10 or 10 are treated as
binary with `-b 2`, while values such as 0x0a or 0a are treated as hexadecimal
values with `-b 16`. Similarly the option controls the output value format
when reading.

The `-e` option controls value argument endianness and must be explicitly set
for blow operations. Typically most values should remain big-endian, however
certain tools, such as the ones creating the `SRK_HASH` for secure boot
purposes, may prepare their output in little-endian format. On read operations
the endianness is set to big-endian by default, however the option can be used
to force big-endian interpretation.

The syslog flag (`-s`) can be used to ease batch processing and limiting
standard output to solely read or blown values while redirecting all logs to
syslog, this mode requires to force all operations (`-Y`).

Example use:

```
# blow hex value (note: confirmation prompt not shown)
crucible -m IMX6UL -r 1 -b 16 -e big blow MAC1_ADDR 0x001f7b1007e3
soc:IMX6UL ref:1 otp:MAC1_ADDR op:blow addr:0x88 off:0 len:48 val:0x001f7b1007e3 res:0xe307107b1f000000

# read hex value
crucible -m IMX6UL -r 1 -b 16 read MAC1_ADDR
soc:IMX6UL ref:1 otp:MAC1_ADDR op:read addr:0x88 off:0 len:48 val:0x001f7b1007e3

# read hex value with minimal standard output
crucible -s -m IMX6UL -r 1 -b 16 read MAC1_ADDR
001f7b1007e3

# read binary value
crucible -m IMX6UL -r 1 -b 2 read SI_REV
soc:IMX6UL ref:1 otp:SI_REV op:read addr:0xc off:16 len:4 val:0b0001

# read binary value with minimal standard output
crucible -s -m IMX6UL -r 1 -b 2 read SI_REV
0001

# read register value with bit map visualization
crucible -l -m IMX6UL -r 1 -b 16 read OCOTP_CFG2
soc:IMX6UL ref:1 otp:OCOTP_CFG2 op:read addr:0xc off:0 len:32 val:0x703100ec

 31 30 29 28 27 26 25 24 23 22 21 20 19 18 17 16 15 14 13 12 11 10 09 08 07 06 05 04 03 02 01 00  OCOTP_CFG2
┏━━┳━━┳━━┳━━┳━━┳━━┳━━┳━━┳━━┳━━┳━━┳━━┳━━┳━━┳━━┳━━┳━━┳━━┳━━┳━━┳━━┳━━┳━━┳━━┳━━┳━━┳━━┳━━┳━━┳━━┳━━┳━━┓ Bank:0 Word:3
┃  ┃  ┃  ┃  ┃  ┃  ┃  ┃  ┃  ┃  ┃1  1 ┃0  0  0  1 ┃  ┃  ┃  ┃  ┃  ┃  ┃  ┃  ┃  ┃  ┃  ┃  ┃  ┃  ┃  ┃  ┃ R: 0x0000000c
┗━━┻━━┻━━┻━━┻━━┻━━┻━━┻━━╋━━┻━━┻━━┻━━┻━━┻━━┻━━┻━━╋━━┻━━┻━━┻━━┻━━┻━━┻━━┻━━╋━━┻━━┻━━┻━━┻━━┻━━┻━━┻━━┛ W: 0x0000000c
                               21 20 ───────────────────────────────────────────────────────────  TAMPER_PIN_DISABLE
                                     19 ┄┄ ┄┄ 16 ───────────────────────────────────────────────  SI_REV
```

Fusemap format
==============

The `crucible` tool relies on register definition files in YAML format (see
`fusemaps` directory for examples) to map register/fuse names and address
information for read and write operations.

The definition format tries to adhere, as much as possible, to the information
contained in the relevant P/N reference manuals.

The syntax is the following:

```
processor: <string>       # processor model
reference: <string>       # reference manual number (for P/N revision match)
driver: <string>          # Linux driver name
                          #
gaps:                     # gap definitions
  <string>:               #   name of first register after gap
    read: <bool>          #     applies to read operation
    write: <bool>         #     applies to write operation
    len: <uint32>         #     gap length in bytes
                          #
registers:                # register definitions
  <string>:               #   register name
    bank: <uint32>        #     bank index
    word: <uint32>        #     word index
    fuses:                #     individual OTP fuse definitions
      <string>:           #       fuse name
        offset: <uint32>  #         fuse offset within register word
        len: <uint32>     #         fuse length in bits
```

When loaded, the fusemap undergoes some basic sanity checks to ensure unique
names, unique register addresses, bank and word indices compatible with the
specified driver.

Development of new fusemaps can be facilitated with the `-l` flag, in
combination with the fusemap selection (`-m` and `-r` flags), to visualize bit
allocation and ease reference manual table comparison.

```
crucible -s -m IMX6UL -r 1 -l
...
 31 30 29 28 27 26 25 24 23 22 21 20 19 18 17 16 15 14 13 12 11 10 09 08 07 06 05 04 03 02 01 00  OCOTP_CFG1
┏━━┳━━┳━━┳━━┳━━┳━━┳━━┳━━┳━━┳━━┳━━┳━━┳━━┳━━┳━━┳━━┳━━┳━━┳━━┳━━┳━━┳━━┳━━┳━━┳━━┳━━┳━━┳━━┳━━┳━━┳━━┳━━┓ Bank:0 Word:2
┃DIE-X-CORDINATE        ┃DIE-Y-CORDINATE        ┃WAFER_NO      ┃LOT_NO_ENC[42:32]               ┃ R: 0x00000008
┗━━┻━━┻━━┻━━┻━━┻━━┻━━┻━━╋━━┻━━┻━━┻━━┻━━┻━━┻━━┻━━╋━━┻━━┻━━┻━━┻━━┻━━┻━━┻━━╋━━┻━━┻━━┻━━┻━━┻━━┻━━┻━━┛ W: 0x00000008
 31 ┄┄ ┄┄ ┄┄ ┄┄ ┄┄ ┄┄ ┄┄ ┄┄ ┄┄ ┄┄ ┄┄ ┄┄ ┄┄ ┄┄ ┄┄ ┄┄ ┄┄ ┄┄ ┄┄ ┄┄ ┄┄ ┄┄ ┄┄ ┄┄ ┄┄ ┄┄ ┄┄ ┄┄ ┄┄ ┄┄ 00  UNIQUE_ID[63:32]
 31 ┄┄ ┄┄ ┄┄ ┄┄ ┄┄ ┄┄ 24 ───────────────────────────────────────────────────────────────────────  DIE-X-CORDINATE
                         23 ┄┄ ┄┄ ┄┄ ┄┄ ┄┄ ┄┄ 16 ───────────────────────────────────────────────  DIE-Y-CORDINATE
                                                 15 ┄┄ ┄┄ ┄┄ 11 ────────────────────────────────  WAFER_NO
                                                                10 ┄┄ ┄┄ ┄┄ ┄┄ ┄┄ ┄┄ ┄┄ ┄┄ ┄┄ 00  LOT_NO_ENC[42:32]
...
```

A bundle of [fusemaps](https://github.com/f-secure-foundry/crucible/tree/master/fusemaps)
for all supported drivers is embedded in the `crucible` executable.

Supported drivers
=================

The following table summarizes the currently supported hardware in terms of
driver and fusemap availability.

| Vendor | Model    | Linux driver    | Read  | Write | Fusemap |
|--------|----------|-----------------|-------|-------|---------|
| NXP    | i.MX53   | nvmem-imx-iim   | yes   | no    | yes     |
| NXP    | i.MX6DL  | nvmem-imx-ocotp | yes   | yes   | yes     |
| NXP    | i.MX6DQ  | nvmem-imx-ocotp | yes   | yes   | yes     |
| NXP    | i.MX6SL  | nvmem-imx-ocotp | yes   | yes   | no      |
| NXP    | i.MX6SLL | nvmem-imx-ocotp | yes   | yes   | no      |
| NXP    | i.MX6SX  | nvmem-imx-ocotp | yes^  | yes   | no      |
| NXP    | i.MX6UL  | nvmem-imx-ocotp | yes^  | yes   | yes     |
| NXP    | i.MX6ULL | nvmem-imx-ocotp | yes   | yes   | yes     |
| NXP    | i.MX6ULZ | nvmem-imx-ocotp | yes   | yes   | yes     |
| NXP    | i.MX7D   | nvmem-imx-ocotp | yes   | yes   | no      |
| NXP    | i.MX7ULP | nvmem-imx-ocotp | yes   | yes   | no      |

^ The nvmem-imx-ocotp driver does not handle addressing gaps between OTP banks,
the fusemap supports gap information specifically to work this problem around
and ensure correct reads (writes are unaffected). Such driver limitation
however does not allow for the entire fusemap to be read as its maximum size is
computed without accounting for the gaps, see comments within the fusemap for
affected registers.

License
=======

crucible | https://github.com/f-secure-foundry/crucible  
Copyright (c) F-Secure Corporation

This program is free software: you can redistribute it and/or modify it under
the terms of the GNU General Public License as published by the Free Software
Foundation under version 3 of the License.

This program is distributed in the hope that it will be useful, but WITHOUT ANY
WARRANTY; without even the implied warranty of MERCHANTABILITY or FITNESS FOR A
PARTICULAR PURPOSE. See the GNU General Public License for more details.

See accompanying LICENSE file for full details.
