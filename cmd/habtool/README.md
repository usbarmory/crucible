NXP HABv4 Secure Boot utility
=============================

crucible | https://github.com/usbarmory/crucible  
Copyright (c) The crucible authors. All Rights Reserved.

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
andrea@inversepath.com  

Andrej Rosano  
andrej@inversepath.com  

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
  -z <crypto backend> "file" (default) or "gcp"

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
  -1 <input path>     SRK public key 1 ('file': PEM format, 'gcp': resource ID)
  -2 <input path>     SRK public key 2 ('file': PEM format, 'gcp': resource ID)
  -3 <input path>     SRK public key 3 ('file': PEM format, 'gcp': resource ID)
  -4 <input path>     SRK public key 4 ('file': PEM format, 'gcp': resource ID)

  -o <output path>    Write SRK table hash to file
  -t <output path>    Write SRK table to file

Executable signing options:
  -A <input path>     CSF private key ('file': PEM format, 'gcp': resource ID)
  -a <input path>     CSF public  key ('file': PEM format, 'gcp': resource ID)
  -B <input path>     IMG private key ('file': PEM format, 'gcp': resource ID)
  -b <input path>     IMG public  key ('file': PEM format, 'gcp': resource ID)
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

Google Cloud support
====================

When setting the `-z` flag to `gcp`, `habtool` will use the Google Cloud APIs to fetch certificates
and perform signing operations. This backend requires that public and private keys are referenced
using [GCP Resource IDs](https://cloud.google.com/config-connector/docs/how-to/managing-resources-with-resource-ids)
rather than on-disk files.

Signing keys must be stored in [CloudHSM](https://cloud.google.com/kms/docs/hsm), and the particular
keys to use when signing the CSF and IMG payloads are passed as
[CloudHSM Key Resource IDs](https://cloud.google.com/kms/docs/getting-resource-ids) to the `-A` and `-B`flags, e.g:
`projects/myProject/locations/global/keyRings/myKeyRing/cryptoKeys/myKey/cryptoKeyVersions/1`.

Public key Resource IDs, passed via the `-1`, `-2`, `-3`, `-4`, `-A`, or `-B` flags, should reference either:

- a
[Certificate](https://cloud.google.com/certificate-authority-service/docs/reference/rest/v1/projects.locations.caPools.certificates#Certificate)
resource, e.g.:
`projects/myProject/locations/us-central1/caPools/myPool/certificates/myCertificate`
- a [CertificateAuthority](https://cloud.google.com/certificate-authority-service/docs/reference/rpc/google.cloud.security.privateca.v1#google.cloud.security.privateca.v1.CertificateAuthority)
resource, e.g.:
`projects/myProject/locations/us-central1/caPools/myPool/certificateAuthorities/myCertificateAuthority`

In the later case, the authoritie's public key certificate will be used.

License
=======

crucible | https://github.com/usbarmory/crucible  
Copyright (c) The crucible authors. All Rights Reserved.

This project is distributed under the BSD-style license found in the
[LICENSE](https://github.com/usbarmory/crucible/blob/master/LICENSE) file.
