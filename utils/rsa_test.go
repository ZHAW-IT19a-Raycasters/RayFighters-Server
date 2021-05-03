package utils_test

import (
	"crypto/rsa"
	"crypto/x509"
	"encoding/pem"
	"testing"

	"github.zhaw.ch/raycasters/server/utils"
)

func TestParseRSA(t *testing.T) {
	const (
		pubKey string = `-----BEGIN PUBLIC KEY-----
MIGeMA0GCSqGSIb3DQEBAQUAA4GMADCBiAKBgG6OF8DB0i8AunYrbgRg/S9Ip/0U
B36IE+poSqkFga6MCJfpjl4gzPbwRST9nd5Nludg+G9iKxy4AsQZ7WN/ZmLTMbb5
LIkMwDFh2WAamftcW2u9GnbFrBfRxzFT4oBTJeX8I9lx0kmpJTeGrcMXNHKNYTDz
VSjdNAw36Jz/OsoPAgMBAAE=
-----END PUBLIC KEY-----
`

		privKey string = `-----BEGIN RSA PRIVATE KEY-----
MIICWwIBAAKBgG6OF8DB0i8AunYrbgRg/S9Ip/0UB36IE+poSqkFga6MCJfpjl4g
zPbwRST9nd5Nludg+G9iKxy4AsQZ7WN/ZmLTMbb5LIkMwDFh2WAamftcW2u9GnbF
rBfRxzFT4oBTJeX8I9lx0kmpJTeGrcMXNHKNYTDzVSjdNAw36Jz/OsoPAgMBAAEC
gYBPfuM7muNdIwPxNvdsNKW6G1+YTo+4/3Z6prWQ5kHC1jTTr4DM1H8tzcWafLfJ
AzfQdt0NkqR3iSWMByy40AI5d7r67lDl83lqcgbXbUXoWv01KNqT8Oteybve542I
1ASyHBtG2KKI7ouIKApI4bg4BSuYbrYkLV2p9El3i6XwQQJBAL9V9pa8iG/2d7By
7yzicYLiKH2ySxOyWz+UEDJ6gp9HsE4h+TkS4s0myZEVNK7LZxzYFM6RtC4c9PzM
PhVK7W8CQQCT6x8Z/3bS87qTd18PmMWHrPIDBxRHbzO+41+bE+25GLXIAotXbXT9
DX1Z//YxbcnebzBZrrRsrM4L4SzcAt1hAkEAn634O1hgbUA+Jy3kWYspvFdN/4ja
izegW3EQC3Cerps/khVcTrtqOFgUhFY0P2RVS4cx5itHOezEsG0rqs17VQJAV5qi
2uTf3OQYlgnWtQgHxOj6lCx3KC+F+AH1t4TAACXQU1qeZ/MdnFmG6eycRH3/VsuH
WTYNCWf2yDZeORdBQQJAQeAzob5no0rMhRy9a3lGggRfxeUQd7g/C9Pfbz/dNj7M
NeNcxDYgME1X7ruvj6BnjgMErInFIr6D5MCDe0vwCw==
-----END RSA PRIVATE KEY-----
`
	)

	privParsed, err := utils.ParseRSA([]byte(privKey))

	if err != nil {
		t.Error("there should not be an error here")
		return
	}

	// check private key
	pemDataPriv := pem.EncodeToMemory(&pem.Block{
		Type:  "RSA PRIVATE KEY",
		Bytes: x509.MarshalPKCS1PrivateKey(privParsed),
	})
	s := string(pemDataPriv)
	if s != privKey {
		t.Errorf("should be parsed correctly got: \n%s", s)
	}

	// check pub key
	pubParsed := privParsed.Public()
	decodedPub, _ := pem.Decode([]byte(pubKey))

	pubKeyParsedInt, err := x509.ParsePKIXPublicKey(decodedPub.Bytes)
	if err != nil {
		t.Errorf("unable to parse the public key -> [%s]", err.Error())
		return
	}

	pubKeyParsed := pubKeyParsedInt.(*rsa.PublicKey)

	if !pubKeyParsed.Equal(pubParsed) {
		t.Error("keys should be equals")
	}

}
