package security

import (
	"crypto/ed25519"
	"encoding/base64"
	"errors"
)

const PublicKeyStr = "3yUKKkTZq2wXqstDTiUo91+ahuvSkYCL9F5xfdlYTlY="

// VerifyLicense checks if the license (signature) is valid for the given machineID.
func VerifyLicense(licenseStr string, machineID string) bool {
	// Decode Public Key
	pubKeyBytes, err := base64.StdEncoding.DecodeString(PublicKeyStr)
	if err != nil {
		return false // Should not happen if constant is correct
	}
	if len(pubKeyBytes) != ed25519.PublicKeySize {
		return false
	}

	// Decode License (Signature)
	sigBytes, err := base64.StdEncoding.DecodeString(licenseStr)
	if err != nil {
		return false
	}

	// Verify
	return ed25519.Verify(ed25519.PublicKey(pubKeyBytes), []byte(machineID), sigBytes)
}

// SignLicense is a helper for the pay server (though pay server is separate,
// having it here might be useful if we share code, but we likely won't link this package in pay server
// to avoid bloating it, but it's fine for reference).
// The pay server will likely implement its own signing using the private key.
func SignLicense(privateKeyBase64 string, machineID string) (string, error) {
	privKeyBytes, err := base64.StdEncoding.DecodeString(privateKeyBase64)
	if err != nil {
		return "", err
	}

	if len(privKeyBytes) != ed25519.PrivateKeySize {
		return "", errors.New("invalid private key size")
	}

	signature := ed25519.Sign(ed25519.PrivateKey(privKeyBytes), []byte(machineID))
	return base64.StdEncoding.EncodeToString(signature), nil
}
