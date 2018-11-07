One-Time-Programmable (OTP) fusing tool
=======================================

crucible | https://github.com/inversepath/crucible  
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

The `crucible` tool provides user space support for reading, and writing,
One-Time-Programmable (OTP) fuses of System-on-Chip (SoC) application
processors.

The current support targets application processors from the NXP i.MX series
(see _Supported drivers_).

The tool consists of a client for SoC drivers which leverage on the
[Linux NVMEM framework](https://github.com/torvalds/linux/blob/master/Documentation/nvmem/nvmem.txt)
to provide read and write access to the processor non-volatile memory.

Warning
=======

Fusing SoC OTPs is an **irreversible** action that permanently fuses values on
the device. This means that any errors in the process, or lost fused data such
as cryptographic key material, might result in a **bricked** device.

The use of this tool is therefore **at your own risk**.

Operation
=========

```
Usage: crucible [options] [read|blow] [fuse/register name] [value]
  -Y	do not prompt for confirmation (DANGEROUS)
  -b int
    	value base/format (2,10,16)
  -f string
    	YAML fuse maps directory (default "fusemaps")
  -l	list fusemaps or fusemap contents (with -m and -r)
  -m string
    	processor model
  -n string
    	NVMEM device (default "/sys/bus/nvmem/devices/imx-ocotp0/nvmem")
  -r string
    	reference manual revision
  -s	use syslog, print ony result value to stdout
```

The value parameter format depends on the passed base argument (`-b`). For
instance with base 2 value arguments of 0b10 or 10 are treated as binary while
base 16 means hexadecimal values such as 0x0a or 0a). The base argument also
controls the output value format when reading.

**IMPORTANT**: The value parameter endianness is always assumed to be
big-endian, it is then converted to little-endian before writing, as required
by the driver. Please note that certain tools, such as the ones creating the
`SRK_HASH` for secure boot purposes, typically already prepare their output in
little-endian format.

The syslog flag (`-s`) can be used to ease batch processing and limiting
standard output to solely read or blown values while redirecting all logs to
syslog, this mode requires to force all operations (`-Y`).

Example use:

```
# blow hex value (note: confirmation prompt not shown)
crucible -m IMX6UL -r 1 -b 16 blow MAC1_ADDR 0x001f7b1007e3
IMX6UL ref:1 op:blow addr:0x88 off:0 len:48 val:0xe307107b1f000000

# read hex value
crucible -m IMX6UL -r 1 -b 16 read MAC1_ADDR
IMX6UL ref:1 op:read addr:0x88 off:0 len:48 val:0x001f7b1007e3

# read hex value with minimal standard output
crucible -s -m IMX6UL -r 1 -b 16 read MAC1_ADDR
001f7b1007e3

# read binary value
crucible -m IMX6UL -r 1 -b 2 read MAC1_ADDR
IMX6UL ref:1 op:read addr:0x88 off:0 len:48 val:0b[00000000 00011111 01111011 00010000 00000111 11100011]

# read binary value with minimal standard output
crucible -s -m IMX6UL -r 1 -b 2 read MAC1_ADDR
1111101111011000100000000011111100011
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
┃DIE-X-CORDINATE        ┃DIE-Y-CORDINATE        ┃WAFER_NO      ┃  ┃  ┃  ┃  ┃  ┃  ┃  ┃  ┃  ┃  ┃  ┃ R: 0x00000008
┗━━┻━━┻━━┻━━┻━━┻━━┻━━┻━━╋━━┻━━┻━━┻━━┻━━┻━━┻━━┻━━╋━━┻━━┻━━┻━━┻━━┻━━┻━━┻━━╋━━┻━━┻━━┻━━┻━━┻━━┻━━┻━━┛ W: 0x00000008
 31 ┄┄ ┄┄ ┄┄ ┄┄ ┄┄ ┄┄ 24 ───────────────────────────────────────────────────────────────────────  DIE-X-CORDINATE
                         23 ┄┄ ┄┄ ┄┄ ┄┄ ┄┄ ┄┄ 16 ───────────────────────────────────────────────  DIE-Y-CORDINATE
...                                              15 ┄┄ ┄┄ ┄┄ 11 ────────────────────────────────  WAFER_NO
```

Supported drivers
=================

The following table summarizes the currently supported hardware in terms of
driver and fusemap availability.

| Vendor | Model   | Linux driver    | Read  | Write | Fusemap |
|--------|---------|-----------------|-------|-------|---------|
| NXP    | i.MX53  | nvmem-imx-iim   | yes   | no    | Yes     |
| NXP    | i.MX6Q  | nvmem-imx-ocotp | yes^  | yes   | No      |
| NXP    | i.MX6SL | nvmem-imx-ocotp | yes^  | yes   | No      |
| NXP    | i.MX6SX | nvmem-imx-ocotp | yes^  | yes   | No      |
| NXP    | i.MX6UL | nvmem-imx-ocotp | yes^  | yes   | Yes     |
| NXP    | i.MX7D  | nvmem-imx-ocotp | yes^  | yes   | No      |

^ The nvmem-imx-ocotp driver does not handle addressing gaps between OTP banks,
the fusemap supports gap information specifically to work this problem around
and ensure correct reads (writes are unaffected). Such driver limitation
however does not allow for the entire fusemap to be read as its maximum size is
computed without accounting for the gaps, see comments within the fusemap for
affected registers.

Installing
==========

You can automatically download, compile and install the package, under your
GOPATH, as follows:

```
go get -u github.com/inversepath/crucible
```

Alternatively you can manually compile it from source:

```
go get -u github.com/ghodss/yaml
git clone https://github.com/inversepath/crucible
cd crucible && make
```

To cross compile for an ARM target it is sufficient to pass `GOARCH=arm` when
compiling.

The default compilation target automatically runs all available unit tests.

License
=======

crucible | https://github.com/inversepath/crucible  
Copyright (c) F-Secure Corporation

This program is free software: you can redistribute it and/or modify it under
the terms of the GNU General Public License as published by the Free Software
Foundation under version 3 of the License.

This program is distributed in the hope that it will be useful, but WITHOUT ANY
WARRANTY; without even the implied warranty of MERCHANTABILITY or FITNESS FOR A
PARTICULAR PURPOSE. See the GNU General Public License for more details.

See accompanying LICENSE file for full details.
