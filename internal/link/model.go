package link

import (
	"crypto/rand"
	"fmt"
	"go-api/internal/database"
	"math/big"
)

func NewLink(url string, hash string) *database.Link {
	return &database.Link{
		Url:  url,
		Hash: hash,
	}
}

var hashAlphabet = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"

func GenerateHash(length int) (string, error) {
	if length <= 0 {
		return "", fmt.Errorf("invalid hash length")
	}
	result := make([]byte, length)
	alphabetLength := big.NewInt(int64(len(hashAlphabet)))
	for i := range result {
		index, err := rand.Int(rand.Reader, alphabetLength)
		if err != nil {
			return "", fmt.Errorf("generate random hash: %w", err)
		}
		result[i] = hashAlphabet[index.Int64()]
	}

	return string(result), nil
}
