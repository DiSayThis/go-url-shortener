package auth

import (
	"crypto/rand"
	"crypto/subtle"
	"encoding/base64"
	"fmt"
	"strings"

	"golang.org/x/crypto/argon2"
)

type Argon2idParams struct {
	Memory      uint32
	Iterations  uint32
	Parallelism uint8
	SaltLength  uint32
	KeyLength   uint32
}

type Argon2idPasswordHasher struct {
	params Argon2idParams
}

const maxEncodedArgon2idHashLength = 1024
const (
	minArgon2Memory      uint32 = 8 * 1024
	maxArgon2Memory      uint32 = 256 * 1024
	minArgon2Iterations  uint32 = 1
	maxArgon2Iterations  uint32 = 10
	minArgon2Parallelism uint8  = 1
	maxArgon2Parallelism uint8  = 16

	minArgon2SaltLength = 16
	maxArgon2SaltLength = 64
	minArgon2HashLength = 16
	maxArgon2HashLength = 64
)

func NewArgon2idPasswordHasher() *Argon2idPasswordHasher {
	return &Argon2idPasswordHasher{
		params: Argon2idParams{
			Memory:      19 * 1024,
			Iterations:  2,
			Parallelism: 1,
			SaltLength:  16,
			KeyLength:   32,
		},
	}
}

func (hasher *Argon2idPasswordHasher) Hash(
	password string,
) (string, error) {
	salt := make([]byte, hasher.params.SaltLength)

	if _, err := rand.Read(salt); err != nil {
		return "", fmt.Errorf("generate password salt: %w", err)
	}

	hash := argon2.IDKey(
		[]byte(password),
		salt,
		hasher.params.Iterations,
		hasher.params.Memory,
		hasher.params.Parallelism,
		hasher.params.KeyLength,
	)

	encodedSalt := base64.RawStdEncoding.EncodeToString(salt)
	encodedHash := base64.RawStdEncoding.EncodeToString(hash)

	encoded := fmt.Sprintf(
		"$argon2id$v=%d$m=%d,t=%d,p=%d$%s$%s",
		argon2.Version,
		hasher.params.Memory,
		hasher.params.Iterations,
		hasher.params.Parallelism,
		encodedSalt,
		encodedHash,
	)

	return encoded, nil
}

func (hasher *Argon2idPasswordHasher) Compare(
	password string,
	encodedHash string,
) error {
	decoded, err := decodeArgon2idHash(encodedHash)
	if err != nil {
		return err
	}

	actualHash := argon2.IDKey(
		[]byte(password),
		decoded.salt,
		decoded.iterations,
		decoded.memory,
		decoded.parallelism,
		uint32(len(decoded.hash)),
	)

	if subtle.ConstantTimeCompare(
		actualHash,
		decoded.hash,
	) != 1 {
		return ErrPasswordMismatch
	}

	return nil
}

type decodedArgon2idHash struct {
	memory      uint32
	iterations  uint32
	parallelism uint8
	salt        []byte
	hash        []byte
}

func decodeArgon2idHash(
	encodedHash string,
) (decodedArgon2idHash, error) {
	if len(encodedHash) == 0 ||
		len(encodedHash) > maxEncodedArgon2idHashLength {
		return decodedArgon2idHash{}, ErrInvalidPasswordHash
	}

	parts := strings.Split(encodedHash, "$")

	// Строка начинается с $, поэтому parts[0] будет пустым.
	//
	// $argon2id$v=19$m=19456,t=2,p=1$salt$hash
	//
	// parts[0] = ""
	// parts[1] = "argon2id"
	// parts[2] = "v=19"
	// parts[3] = "m=19456,t=2,p=1"
	// parts[4] = salt
	// parts[5] = hash
	if len(parts) != 6 {
		return decodedArgon2idHash{}, ErrInvalidPasswordHash
	}

	if parts[1] != "argon2id" {
		return decodedArgon2idHash{}, ErrInvalidPasswordHash
	}

	var version int

	if _, err := fmt.Sscanf(
		parts[2],
		"v=%d",
		&version,
	); err != nil {
		return decodedArgon2idHash{}, ErrInvalidPasswordHash
	}

	if version != argon2.Version {
		return decodedArgon2idHash{}, ErrInvalidPasswordHash
	}

	var (
		memory      uint32
		iterations  uint32
		parallelism uint8
	)

	if _, err := fmt.Sscanf(
		parts[3],
		"m=%d,t=%d,p=%d",
		&memory,
		&iterations,
		&parallelism,
	); err != nil {
		return decodedArgon2idHash{}, ErrInvalidPasswordHash
	}
	if memory < minArgon2Memory ||
		memory > maxArgon2Memory ||
		memory < 8*uint32(parallelism) ||
		iterations < minArgon2Iterations ||
		iterations > maxArgon2Iterations ||
		parallelism < minArgon2Parallelism ||
		parallelism > maxArgon2Parallelism {
		return decodedArgon2idHash{}, ErrInvalidPasswordHash
	}

	salt, err := base64.RawStdEncoding.DecodeString(parts[4])
	if err != nil {
		return decodedArgon2idHash{}, ErrInvalidPasswordHash
	}

	hash, err := base64.RawStdEncoding.DecodeString(parts[5])
	if err != nil {
		return decodedArgon2idHash{}, ErrInvalidPasswordHash
	}
	if len(salt) < minArgon2SaltLength ||
		len(salt) > maxArgon2SaltLength ||
		len(hash) < minArgon2HashLength ||
		len(hash) > maxArgon2HashLength {
		return decodedArgon2idHash{}, ErrInvalidPasswordHash
	}

	return decodedArgon2idHash{
		memory:      memory,
		iterations:  iterations,
		parallelism: parallelism,
		salt:        salt,
		hash:        hash,
	}, nil
}
