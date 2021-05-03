package utils

import (
	"crypto/rsa"
	"crypto/x509"
	"encoding/pem"
	"errors"
	"fmt"
)

// ErrInvalidRSAData is the state that happens, when the rsa data
// is not valid PEM encoding.
var ErrInvalidPEMData = errors.New("data is valid pem encoding")

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
	pKey, err := x509.ParsePKCS1PrivateKey(block.Bytes)

	if err != nil {
		return nil, ErrInvalidRSAKey{
			internal: err,
		}
	}
	return pKey, nil
}
