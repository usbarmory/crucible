One-Time-Programmable (OTP) fusing tool
=======================================

crucible | https://github.com/usbarmory/crucible  
Copyright (c) WithSecure Corporation

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
andrea.barisani@withsecure.com | andrea@inversepath.com  

Andrej Rosano  
andrej.rosano@withsecure.com | andrej@inversepath.com  

Introduction
============

The `crucible` utility provides user space support for reading, and writing,
One-Time-Programmable (OTP) fuses of System-on-Chip (SoC) application
processors through the [Linux NVMEM framework](https://github.com/torvalds/linux/blob/master/Documentation/driver-api/nvmem.rst).

The `habtool` utility provides support functions for NXP HABv4
[Secure Boot](https://github.com/usbarmory/usbarmory/wiki/Secure-boot-(Mk-II))
provisioning and executable signing.

The current support targets application processors from the NXP i.MX series
(see _Supported drivers_).

Warning
=======

Fusing SoC OTPs is an **irreversible** action that permanently fuses values on
the device. This means that any errors in the process, or lost fused data such
as cryptographic key material, might result in a **bricked** device.

The use of these tools is therefore **at your own risk**.

Installing
==========

Pre-compiled binaries for Linux and Windows are released
[here](https://github.com/usbarmory/crucible/releases).

You can also automatically download, compile and install the package, under
your GOPATH, as follows:

```
# crucible fusing tool
go install github.com/usbarmory/crucible/cmd/crucible@latest

# NXP HABv4 tool
go install github.com/usbarmory/crucible/cmd/habtool@latest
```

Alternatively you can manually compile it from source:

```
git clone https://github.com/usbarmory/crucible
cd crucible && make
```

All targets can be cross compiled for ARM as follows:

* 32-bit arm: `make GOARCH=arm`
* 64-bit arm: `make GOARCH=arm64`
* To just build the crucible executable for 64-bit arm: `make crucible GOARCH=arm64`

The `habtool` utility can be cross compiled Windows as follows:

```
make habtool.exe
```

Crucible
========

Operation
---------

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
    	visualize fusemap      (with -m and -r)
    	visualize read value   (with read operation on a register)
    	visualize read fusemap (with read operation and no register)
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
--------------

The `crucible` tool relies on register definition files in YAML format (see
`fusemaps` directory for examples) to map register/fuse names and address
information for read and write operations.

The definition format tries to adhere, as much as possible, to the information
contained in the relevant P/N reference manuals.

The syntax is the following:

```
processor: <string>       # processor model
reference: <string>       # reference manual number (for P/N revision match)
                          #
driver: <string>          # Linux driver name
bank_size: <int>          # bank size
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
crucible -m IMX6UL -r 1 -l
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

A bundle of [fusemaps](https://github.com/usbarmory/crucible/tree/master/fusemaps)
for all supported drivers is embedded in the `crucible` executable.

Supported drivers
-----------------

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
| NXP    | i.MX7D   | nvmem-imx-ocotp | yes   | yes   | yes     |
| NXP    | i.MX7ULP | nvmem-imx-ocotp | yes   | yes   | no      |
| NXP    | i.MX8M   | nvmem-imx-ocotp | yes   | yes   | yes     |
| NXP    | i.MX8MM  | nvmem-imx-ocotp | yes   | yes   | yes     |
| NXP    | i.MX8MP  | nvmem-imx-ocotp | yes   | yes   | yes     |

^ The nvmem-imx-ocotp driver does not handle addressing gaps between OTP banks,
the fusemap supports gap information specifically to work this problem around
and ensure correct reads (writes are unaffected). Such driver limitation
however does not allow for the entire fusemap to be read as its maximum size is
computed without accounting for the gaps, see comments within the fusemap for
affected registers.

HABv4 tool
==========

Operation
---------

```
Usage: habtool [OPTIONS]
  -h                  Show this help

SRK CA creation options:
  -C <output path>    SRK private key in PEM format
  -c <output path>    SRK public  key in PEM format

CSF/IMG certificates creation options:
  -C <input path>     SRK private key in PEM format
  -c <input path>     SRK public  key in PEM format

  -A <output path>    CSF private key in PEM format
  -a <output path>    CSF public  key in PEM format
  -B <output path>    IMG private key in PEM format
  -b <output path>    IMG public  key in PEM format

SRK table creation options:
  -1 <input path>     SRK public key 1 in PEM format
  -2 <input path>     SRK public key 2 in PEM format
  -3 <input path>     SRK public key 3 in PEM format
  -4 <input path>     SRK public key 4 in PEM format

  -o <output path>    Write SRK table hash to file
  -t <output path>    Write SRK table to file

Executable signing options:
  -A <input path>     CSF private key in PEM format
  -a <input path>     CSF public  key in PEM format
  -B <input path>     IMG private key in PEM format
  -b <input path>     IMG public  key in PEM format
  -t <input path>     Read SRK table from file
  -x <1-4>            Index for SRK key
  -e <id>             Crypto engine (e.g. 0x1b for HAB_ENG_DCP)
  -i <input path>     Image file w/ IVT header (e.g. boot.imx)

  -o <output path>    Write CSF to file

  -s                  Serial download mode
  -S <address>        Serial download DCD OCRAM address
                      (depends on mfg tool, default: 0x00910000)
```

The [USB armory](https://github.com/usbarmory/usbarmory/wiki) guide for
[Secure Boot](https://github.com/usbarmory/usbarmory/wiki/Secure-boot-(Mk-II))
provides an introduction on HABv4 using the USB armory Mk II as reference platform.

License
=======

crucible | https://github.com/usbarmory/crucible  
Copyright (c) WithSecure Corporation

This project is distributed under the BSD-style license found in the
[LICENSE](https://github.com/usbarmory/crucible/blob/master/LICENSE) file.
