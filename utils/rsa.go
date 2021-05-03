package utils

import (
	"crypto/rsa"
	"crypto/x509"
	"encoding/pem"
)

// ParseRSA will parse the rsa data package anmd
// return a private instance.
func ParseRSA(data []byte) (*rsa.PrivateKey, error) {
	// see https://stackoverflow.com/questions/44230634/how-to-read-an-rsa-key-from-file/44231740
	// for more information
	block, _ := pem.Decode(data)
	return x509.ParsePKCS1PrivateKey(block.Bytes)
}
