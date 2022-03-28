NXP HABv4 Secure Boot utility
=============================

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

The `habtool` utility provides support functions for NXP HABv4
[Secure Boot](https://github.com/usbarmory/usbarmory/wiki/Secure-boot-(Mk-II))
provisioning and executable signing.

Warning
=======

Fusing SoC OTPs is an **irreversible** action that permanently fuses values on
the device. This means that any errors in the process, or lost fused data such
as cryptographic key material, might result in a **bricked** device.

The use of this tool is therefore **at your own risk**.

Installing
==========

Pre-compiled binaries for Linux and Windows are released
[here](https://github.com/usbarmory/crucible/releases).

You can also automatically download, compile and install the package, under
your GOPATH, as follows:

```
go install github.com/usbarmory/crucible/cmd/habtool@latest
```

Alternatively you can manually compile it from source:

```
git clone https://github.com/usbarmory/crucible
cd crucible && make
```

The utility can be cross compiled Windows as follows:

```
make habtool.exe
```

Operation
=========

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

This program is free software: you can redistribute it and/or modify it under
the terms of the GNU General Public License as published by the Free Software
Foundation under version 3 of the License.

This program is distributed in the hope that it will be useful, but WITHOUT ANY
WARRANTY; without even the implied warranty of MERCHANTABILITY or FITNESS FOR A
PARTICULAR PURPOSE. See the GNU General Public License for more details.

See accompanying LICENSE file for full details.
