// Package signing verifies code signatures made by the CLI.
// The server only ever holds public certificates, never private keys.
package signing

import (
	"crypto/rsa"
	"errors"
)

// PublicKeyFromCertificate parses the app's PEM certificate.
func PublicKeyFromCertificate(certPEM []byte) (*rsa.PublicKey, error) {
	// TODO: pem.Decode → x509.ParseCertificate → *rsa.PublicKey
	return nil, errors.New("not implemented")
}

// Verify checks an expo-signature header value (sig="...", keyid="...") against body.
func Verify(pub *rsa.PublicKey, body []byte, signature string) error {
	// TODO: parse sig="..." → base64 decode → rsa.VerifyPKCS1v15(sha256(body))
	return errors.New("not implemented")
}
