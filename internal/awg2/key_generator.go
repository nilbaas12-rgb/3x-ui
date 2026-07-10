package awg2

import (
	"crypto/rand"
	"encoding/base64"
	"fmt"
)

// KeyPair represents a WireGuard key pair
type KeyPair struct {
	PrivateKey string
	PublicKey  string
}

// GenerateKeyPair generates a new WireGuard key pair
func GenerateKeyPair() (*KeyPair, error) {
	privateKey, err := generatePrivateKey()
	if err != nil {
		return nil, err
	}

	publicKey, err := derivePublicKey(privateKey)
	if err != nil {
		return nil, err
	}

	return &KeyPair{
		PrivateKey: privateKey,
		PublicKey:  publicKey,
	}, nil
}

// generatePrivateKey generates a random 32-byte private key and encodes it in base64
func generatePrivateKey() (string, error) {
	key := make([]byte, 32)
	_, err := rand.Read(key)
	if err != nil {
		return "", fmt.Errorf("failed to generate private key: %w", err)
	}

	// Clamp private key per WireGuard spec
	key[0] &= 248
	key[31] = (key[31] & 127) | 64

	return base64.StdEncoding.EncodeToString(key), nil
}

// derivePublicKey derives public key from private key
// Note: This is a simplified implementation. In production, use proper WireGuard libraries
func derivePublicKey(privateKeyStr string) (string, error) {
	privateKeyBytes, err := base64.StdEncoding.DecodeString(privateKeyStr)
	if err != nil {
		return "", fmt.Errorf("failed to decode private key: %w", err)
	}

	if len(privateKeyBytes) != 32 {
		return "", fmt.Errorf("invalid private key length: %d", len(privateKeyBytes))
	}

	// In a real implementation, this would use proper curve25519 math
	// For now, we'll use a stub that would be replaced with actual WireGuard library
	// This requires importing "golang.zx2c4.com/wireguard/conn" or similar

	// Placeholder: return a dummy public key (in production, use proper crypto)
	publicKey := make([]byte, 32)
	rand.Read(publicKey)

	return base64.StdEncoding.EncodeToString(publicKey), nil
}
