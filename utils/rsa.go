package utils

import (
	"crypto/rand"
	"crypto/rsa"
	"crypto/x509"
	"encoding/pem"
	"errors"
	"fmt"
)

const RSABitSize int = 2048

// ErrInvalidRSAData is the state that happens, when the rsa data
// is not valid PEM encoding.
var ErrInvalidPEMData = errors.New("data is not valid pem encoding")

var ErrInvalidPEMHeader = errors.New("invalid PEM header please use `RSA PRIVATE KEY` instead")

// ErrInvalidRSAKey is returned when there was an invalid
// parsing of the input data.
type ErrInvalidRSAKey struct {
	internal error
}

// Error satisfies the error interface
func (err ErrInvalidRSAKey) Error() string {
	return fmt.Sprintf("the pased data is not valid RSA, the following error happend: [%s]", err.internal.Error())
}

// Is satisfies the errors.Is interface
func (err ErrInvalidRSAKey) Is(target error) bool {
	_, ok := target.(ErrInvalidRSAKey)
	return ok
}

// Unwrap satisfies the errors.Unwrap interface
func (err ErrInvalidRSAKey) Unwrap() error {
	return err.internal
}

// ParseRSA will parse the rsa data package and
// return a private instance.
func ParseRSA(data []byte) (*rsa.PrivateKey, error) {
	// see https://stackoverflow.com/questions/44230634/how-to-read-an-rsa-key-from-file/44231740
	// for more information
	block, _ := pem.Decode(data)
	if block == nil {
		return nil, ErrInvalidRSAKey{
			internal: ErrInvalidPEMData,
		}
	}

	// we only support a single RSA scheme
	if block.Type != "RSA PRIVATE KEY" {
		return nil, ErrInvalidPEMHeader
	}

	pKey, err := x509.ParsePKCS1PrivateKey(block.Bytes)

	if err != nil {
		return nil, ErrInvalidRSAKey{
			internal: err,
		}
	}
	return pKey, nil
}

// GenerateRSA can generate a private rsa key of
// RSABitSize.
// Attention only use this function for testing
// do not use this line in production as it can
// not guarantie a long lasting private key.
func GenerateRSA() (*rsa.PrivateKey, error) {
	return rsa.GenerateKey(rand.Reader, RSABitSize)
}
