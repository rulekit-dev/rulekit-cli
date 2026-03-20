package wizard

import (
	"crypto/rand"
	"encoding/hex"
	"fmt"
)

// GenerateSecret generates a cryptographically random hex string of the given byte length.
func GenerateSecret(byteLen int) (string, error) {
	b := make([]byte, byteLen)
	if _, err := rand.Read(b); err != nil {
		return "", fmt.Errorf("generate secret: %w", err)
	}
	return hex.EncodeToString(b), nil
}

// GenerateAPIKey generates a 32-byte random hex API key.
func GenerateAPIKey() (string, error) {
	return GenerateSecret(32)
}
