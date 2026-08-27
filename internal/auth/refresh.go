package auth

import (
	"crypto/rand"
	"crypto/sha256"
	"encoding/base64"
	"fmt"
	"strings"
)

const refreshTokenSize = 32

type GeneratedRefreshToken struct {
	Raw  string
	Hash []byte
}

func GenerateRefreshToken() (GeneratedRefreshToken, error) {
	bytes := make([]byte, 32)

	if _, err := rand.Read(bytes); err != nil {
		return GeneratedRefreshToken{}, fmt.Errorf("generate refresh token: %w", err)
	}

	raw := base64.RawURLEncoding.EncodeToString(bytes)
	hash := sha256.Sum256([]byte(raw))

	return GeneratedRefreshToken{
		Raw:  raw,
		Hash: hash[:],
	}, nil
}

func HashRefreshToken(raw string) ([]byte, error) {
	raw = strings.TrimSpace(raw)

	expectedLength := base64.RawURLEncoding.EncodedLen(refreshTokenSize)
	if len(raw) != expectedLength {
		return nil, ErrInvalidRefreshToken
	}

	decoded, err := base64.RawURLEncoding.DecodeString(raw)
	if err != nil || len(decoded) != refreshTokenSize {
		return nil, ErrInvalidRefreshToken
	}

	hash := sha256.Sum256([]byte(raw))

	return hash[:], nil
}
