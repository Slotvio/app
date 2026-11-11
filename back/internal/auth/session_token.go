package auth

import (
	"crypto/rand"
	"crypto/sha256"
	"encoding/base64"
	"fmt"
)

func GenerateSessionToken() (token string, hash [32]byte, err error) {
	var randomBytes [32]byte
	if _, err := rand.Read(randomBytes[:]); err != nil {
		return "", hash, fmt.Errorf("generate session token: %w", err)
	}

	token = base64.RawURLEncoding.EncodeToString(randomBytes[:])
	hash = HashSessionToken(token)
	return token, hash, nil
}

func HashSessionToken(token string) [32]byte {
	return sha256.Sum256([]byte(token))
}
