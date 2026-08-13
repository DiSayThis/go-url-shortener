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

func NewArgon2idPasswordHasher() *Argon2idPasswordHasher {
	return &Argon2idPasswordHasher{
		params: Argon2idParams{
			Memory:      19 * 1024, // 19 MiB в KiB
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

	// Вычисляем Argon2id-хеш.
	hash := argon2.IDKey(
		[]byte(password),
		salt,
		hasher.params.Iterations,
		hasher.params.Memory,
		hasher.params.Parallelism,
		hasher.params.KeyLength,
	)

	// Сохраняем бинарные salt и hash как Base64.
	encodedSalt := base64.RawStdEncoding.EncodeToString(salt)
	encodedHash := base64.RawStdEncoding.EncodeToString(hash)

	// В строке сохраняются также алгоритм и его параметры.
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

	// Вычисляем хеш введённого пароля с параметрами и salt,
	// которые хранятся в encodedHash.
	actualHash := argon2.IDKey(
		[]byte(password),
		decoded.salt,
		decoded.iterations,
		decoded.memory,
		decoded.parallelism,
		uint32(len(decoded.hash)),
	)

	// Нельзя сравнивать security-sensitive значения обычным string == string.
	// ConstantTimeCompare уменьшает утечку информации через время выполнения.
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

	salt, err := base64.RawStdEncoding.DecodeString(parts[4])
	if err != nil {
		return decodedArgon2idHash{}, ErrInvalidPasswordHash
	}

	hash, err := base64.RawStdEncoding.DecodeString(parts[5])
	if err != nil {
		return decodedArgon2idHash{}, ErrInvalidPasswordHash
	}

	if len(salt) == 0 || len(hash) == 0 {
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
