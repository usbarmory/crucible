// crucible
// One-Time-Programmable (OTP) fusing tool
//
// Copyright (c) WithSecure Corporation
//
// Use of this source code is governed by the license
// that can be found in the LICENSE file.

package hab

import (
	"bytes"
	"crypto/rsa"
	"crypto/x509"
	"encoding/hex"
	"encoding/pem"
	"testing"
)

const srk1 = `
-----BEGIN CERTIFICATE-----
MIIDYDCCAkigAwIBAgIUVOfHUbVMUa6RtVhPdyvr8sZ33KMwDQYJKoZIhvcNAQEL
BQAwHDEaMBgGA1UEAxQRU1JLXzFfc2hhMjU2XzIwNDgwHhcNMjEwMzA4MDgyOTQx
WhcNMzEwMzA2MDgyOTQxWjAcMRowGAYDVQQDFBFTUktfMV9zaGEyNTZfMjA0ODCC
ASIwDQYJKoZIhvcNAQEBBQADggEPADCCAQoCggEBANT6413cZo/DhQ3bgz07b0ER
hz+TFhOqZwJBVDbRp6Oil+HwABJCW0ogqGOV4IDSPtRYe03unEoPLaHVsbiqTgTK
Lfo7GBnirr9nTdm06KIlh9W+x6hXVMniQbbpQuHsYZ5px2piau0YAAesRQyR//Tz
bQTDoq705FftiDCF7vHYv6KKCMq1cRSAA4DB+F0+tWeKaQYOkPrVKWC8rmDNaiVZ
o32wHr6YOxG0LwGinEgV+eR0Wv9t/Q3gZi3QkX/wKjINYCaN27Y132UT/adUgSod
ZDyjF/ux3yNFy6afhhe6dfrHfUwac2enJ/5UPJZBrQcZnNuy8RciTCV08YzRs0kC
AwEAAaOBmTCBljAdBgNVHQ4EFgQUcZUAq5M6mZroO+gRpk/30F+BEz8wVwYDVR0j
BFAwToAUcZUAq5M6mZroO+gRpk/30F+BEz+hIKQeMBwxGjAYBgNVBAMUEVNSS18x
X3NoYTI1Nl8yMDQ4ghRU58dRtUxRrpG1WE93K+vyxnfcozAPBgNVHRMBAf8EBTAD
AQH/MAsGA1UdDwQEAwICBDANBgkqhkiG9w0BAQsFAAOCAQEANQ5ywq3ty0J/hyQl
A1E7JUyDBj1CNSmOHVGb1dOY2VW40G3QtWu7o/y8iS+WmbH9dB1q8Y175SvU8668
F90YMUMvtU0pUHkpnfpwwIM25oV83REkt048IdVCffAbloySdwocpbPfIdyXTPvK
9WDIP1urtll0KbOZU+FNxIUSnqsBw4ovS3qNDCJh0P76zV5BSZVwnUn6od5yl+Lc
OwAWkAlaQFsQ8rhyl4SLwy5m2MXQMiM5iJlSYH4wsPh7E0ZELH5cCrq+aihdmWZk
ADCROzrO8BudLzrzgt2vXpnhTZ5x/r/esPlQCv/eB+9EhHyclKFRoxzmGfe1+Qcz
qMy3Wg==
-----END CERTIFICATE-----`

const srk2 = `
-----BEGIN CERTIFICATE-----
MIIDYDCCAkigAwIBAgIUPJnN3U7T1CCzf/VbwSuF+FrwKa0wDQYJKoZIhvcNAQEL
BQAwHDEaMBgGA1UEAxQRU1JLXzJfc2hhMjU2XzIwNDgwHhcNMjEwMzA4MDgyOTQx
WhcNMzEwMzA2MDgyOTQxWjAcMRowGAYDVQQDFBFTUktfMl9zaGEyNTZfMjA0ODCC
ASIwDQYJKoZIhvcNAQEBBQADggEPADCCAQoCggEBALk/PVvIqX+NTSwVPpo9q0KO
//9W8KTlOAAIZp/QbXo8rUJie9E84p7hxMs64tEYuDq5u6yp2Wtsk6VFrqfTj7fx
kohYdxSVd+MMDHiftoqFnJn7eBzkoEoD6VoZs+nD9CeEK0NocD3B2qJPS8rD4mwp
VoQBrtkKd6zyqgR8YO3Iau15RPAyQ/vhsVijCFSzOg8Chib9WR14Dnbq4KvOMVok
EWfEUmVnR/zIZCwpbCsTHxyMBQMl/Nccf60jRlLKE4dPxLshqLRslkpoGoxxErGK
f83iIlrmv+p41aCJpYbX4fr8NmWkteaWN8RCaR8dXOcMct5zM/59lRNEzNNCyzEC
AwEAAaOBmTCBljAdBgNVHQ4EFgQUQcqSj5PtbKIy9i/qZzeVGwXjpN0wVwYDVR0j
BFAwToAUQcqSj5PtbKIy9i/qZzeVGwXjpN2hIKQeMBwxGjAYBgNVBAMUEVNSS18y
X3NoYTI1Nl8yMDQ4ghQ8mc3dTtPUILN/9VvBK4X4WvAprTAPBgNVHRMBAf8EBTAD
AQH/MAsGA1UdDwQEAwICBDANBgkqhkiG9w0BAQsFAAOCAQEALvHYKkwDSx3h3y4j
MeGwOjpNBGdKQfSc9v2TBFSPMw7AcHJux9r9Lndvd5dhbrOkGMQgE27DLnkisgy/
0ee6cewY3CbxS00DLGb072zj70A60LQ9u1M3gJRkNFW7rS/717NJM0CsKFXRRGhI
Ia0YIKfa9AHVxULfaIjU9ltbHZWTq9h8Ars4ViUbK0lNVTtBV/JRb8rp0GdL7E+6
AntCB3YwKb+oOk/Jp2Z3IzIv4YqVMMIaHd9ERnL4FQuFu/N7q2zVRgrnHL4LSz6h
quEVBsI8EtVWefA1wKVvz0h3tkcyJfzCcpzl05FB1ioRrpPdZ2LbAMzR3lc2dlzu
yaHKaQ==
-----END CERTIFICATE-----`

const srk3 = `
-----BEGIN CERTIFICATE-----
MIIDYDCCAkigAwIBAgIUXtJnF4taNxCZ1s052o+1dFudTU4wDQYJKoZIhvcNAQEL
BQAwHDEaMBgGA1UEAxQRU1JLXzNfc2hhMjU2XzIwNDgwHhcNMjEwMzA4MDgyOTQy
WhcNMzEwMzA2MDgyOTQyWjAcMRowGAYDVQQDFBFTUktfM19zaGEyNTZfMjA0ODCC
ASIwDQYJKoZIhvcNAQEBBQADggEPADCCAQoCggEBAKHjbDZio7ls9/iKr+9zQu41
/FtZTSq8X2GgjjLda9B4rYdfgAL3L41si6EwnjLjV+Aq2oUU0t0UdRak5NvnEqA1
SjB0i5jDOBMNKlDAol+WBdmflnKAUj7A+BmcujSQ7q+OEglSn0aDd/5SN9C7X5yE
JTKfLkgvGNz9FIkBh5+I8InGQyc4aKhskc4/xHZN8QjMqeJ6kizoTCOxLITyDZIC
Vhtt2sYNiX/kARZe29TRxVCsB8EO3r+e9vP55s7k+cP4D3Z0EGf8UvF84ktXBgom
ChjqLNiFGchLmVqaJ/VBaGyoKSTWD2g/uWs5SmwA/SIYWG3XRG6TQuhi/L4kU8cC
AwEAAaOBmTCBljAdBgNVHQ4EFgQUzoIgNX/ss+Mwkt/myPxnVCKpArQwVwYDVR0j
BFAwToAUzoIgNX/ss+Mwkt/myPxnVCKpArShIKQeMBwxGjAYBgNVBAMUEVNSS18z
X3NoYTI1Nl8yMDQ4ghRe0mcXi1o3EJnWzTnaj7V0W51NTjAPBgNVHRMBAf8EBTAD
AQH/MAsGA1UdDwQEAwICBDANBgkqhkiG9w0BAQsFAAOCAQEAObzrSQYxoPNBOsI9
tbYeJXPCk5C4HU89g7cNGz8PebWWRfkBGTLRWHZrpIzeZHgiEI3ROf50bqE5197P
g8n0JHZdfPWbSbV8VmE+vSYKFp7BIm+yHREvlFqJZnSqiFpbvRjzSKQWgk8U/+Eh
5ETrk6/qgqEgXYNAxB+oK91S06q9BOmslWKKV/qtYre5MVTBHwGcXJFmCF2CDjcO
kuY3hyjK/fzctbXU7V2SW7AkMT9JKMeDaTNN+Qi5xP0+0eGhHWCH4+dLXi+ZONHW
6voYbJMKzwEKkSTaG1Ic5Ip9RKyUhRlOu8bsfhyKQ3D228uqEmjbeaKa9wsPm6D+
ZwI3PA==
-----END CERTIFICATE-----`

const srk4 = `
-----BEGIN CERTIFICATE-----
MIIDYDCCAkigAwIBAgIUCcZ15PcyT5YT20XqaBa+dTfHI/wwDQYJKoZIhvcNAQEL
BQAwHDEaMBgGA1UEAxQRU1JLXzRfc2hhMjU2XzIwNDgwHhcNMjEwMzA4MDgyOTQy
WhcNMzEwMzA2MDgyOTQyWjAcMRowGAYDVQQDFBFTUktfNF9zaGEyNTZfMjA0ODCC
ASIwDQYJKoZIhvcNAQEBBQADggEPADCCAQoCggEBANf2Jy9kHixK5zdZIu2m0DH0
0dAsn2WR9wAz1oYp59QiqSjEdN3GBK9iqV6PiQ1Qg+1kZJMX+t7dVW/D+hgJLjnU
UkhcLnvTs9iYQolKEUx4vDoM/pdwQCG+LuGhb6Nl0BnXg320nFwc+2Pef22JC1yS
kuhXzzg4PPxaM7Yhd5deWAuuny20OvUfGC4mQ9u/4nwuv2+ecusK0kdMXG95wmUh
a8hqUh295g1QvG8PDbLBcsqWDk7Ch8ZxDa2K6wXsC5XJoaFmoSPSCt1Swh0UaWzs
KcI1Tm/am113mR9yVTtLwaczlHWzvNZ5n8c8bLf7S8UZxNPOLQRCluqMHA5R5U0C
AwEAAaOBmTCBljAdBgNVHQ4EFgQUoLXp/tSVIyd8JzH1FNa4/egmm8YwVwYDVR0j
BFAwToAUoLXp/tSVIyd8JzH1FNa4/egmm8ahIKQeMBwxGjAYBgNVBAMUEVNSS180
X3NoYTI1Nl8yMDQ4ghQJxnXk9zJPlhPbRepoFr51N8cj/DAPBgNVHRMBAf8EBTAD
AQH/MAsGA1UdDwQEAwICBDANBgkqhkiG9w0BAQsFAAOCAQEAtsOT6qT4H3o5hUr8
kX1Y5HbJ5bMZP2bd8o7PxC9RkZ+uXkPsYMRBaJB2RrzdFIYPbsp/v+cnP7oWhjMs
frd4txjq03I8KKIGUrjLp2vbBTGtiOdg1pViQH8m/MZ+KaDpv6S17T7smWFYMsVQ
wywXmzOExpg4J9LfwTipu1xX9TZOIQnnOZmz2Zmik60orVBqp+vu9a+LU12Naxuy
/u2TGTuc1BMd/AdbmeABhu+cQtjefHSA/ndUYnUjXHV9ExpFYkJvDMOGTCZXP7cZ
kdxDfuuuNIsD68zTvDO2RtnXIgtTtsy1ppKIHBF7xf8Rp91AcF4BgMx8aNdCqlWW
BFzHew==
-----END CERTIFICATE-----`

const hashOne = "b78a0e67698057068ac2ebce06754951a3cbbbbf17b45e59fd135b8c4a772b81"
const hashAll = "a6bd4b05e61ccab70c636c0d851036cae97fa5b5d58fb8111b2e6f65c67c096a"

func TestPartialSRKTable(t *testing.T) {
	refHash, err := hex.DecodeString(hashOne)

	if err != nil {
		t.Fatal(err)
	}

	table, _ := NewSRKTable(nil)

	if err := table.AddKey(mustKeyFromPEM(t, srk1)); err != nil {
		t.Fatal(err)
	}

	srkHash := table.Hash()

	if !bytes.Equal(srkHash[:], refHash) {
		t.Errorf("SRK table with unexpected value, %x != %x", srkHash, refHash)
	}
}

func TestFullSRKTable(t *testing.T) {
	refHash, err := hex.DecodeString(hashAll)

	if err != nil {
		t.Fatal(err)
	}

	srks := []*rsa.PublicKey{
		mustKeyFromPEM(t, srk1),
		mustKeyFromPEM(t, srk2),
		mustKeyFromPEM(t, srk3),
		mustKeyFromPEM(t, srk4),
	}

	table, err := NewSRKTable(srks)

	if err != nil {
		t.Fatal(err)
	}

	srkHash := table.Hash()

	if !bytes.Equal(srkHash[:], refHash) {
		t.Errorf("SRK table with unexpected value, %x != %x", srkHash, refHash)
	}
}

func mustKeyFromPEM(t *testing.T, b string) *rsa.PublicKey {
	t.Helper()
	der, _ := pem.Decode([]byte(b))
	if der == nil {
		t.Fatal("Invalid testdata PEM")
	}
	c, err := x509.ParseCertificate(der.Bytes)
	if err != nil {
		t.Fatalf("Invalid testdata: %v", err)
	}
	r, ok := c.PublicKey.(*rsa.PublicKey)
	if !ok {
		t.Fatal("Invalid testdata: cert key must be RSA")
	}
	return r

}
