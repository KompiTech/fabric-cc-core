package engine

import (
	"crypto/sha256"
	"crypto/sha512"
	"crypto/x509"
	"encoding/hex"
	"strings"
)

// CertSHA512IDFunc is legacy function that was used before pluggable identity ID getting
var CertSHA512IDFunc = func(cert *x509.Certificate) (string, error) {
	hash := sha512.Sum512(cert.Raw)
	return strings.ToLower(hex.EncodeToString(hash[:])), nil
}

// CertSHA256IDFunc is example of different IDFunc for tests
var CertSHA256IDFunc = func(cert *x509.Certificate) (string, error) {
	hash := sha256.Sum256(cert.Raw)
	return strings.ToLower(hex.EncodeToString(hash[:])), nil
}
