package auth

import (
	"crypto/rand"
	"crypto/subtle"
	"encoding/base64"
	"fmt"
	"os"
	"strconv"
	"strings"

	"golang.org/x/crypto/argon2"
)

type PasswordConfig struct {
	Memory      uint32
	Iterations  uint32
	Parallelism uint8
	SaltLen     uint32
	KeyLen      uint32
}

func DefaultPasswordConfig() *PasswordConfig {
	return &PasswordConfig{
		Memory:      parseEnvUint32("ANTCLAW_ARGON2_MEMORY", 64*1024),
		Iterations:  parseEnvUint32("ANTCLAW_ARGON2_ITERATIONS", 3),
		Parallelism: parseEnvUint8("ANTCLAW_ARGON2_PARALLELISM", 2),
		SaltLen:     parseEnvUint32("ANTCLAW_ARGON2_SALT_LEN", 16),
		KeyLen:      parseEnvUint32("ANTCLAW_ARGON2_KEY_LEN", 32),
	}
}

func HashPassword(password string) (string, error) {
	cfg := DefaultPasswordConfig()
	salt := make([]byte, cfg.SaltLen)
	if _, err := rand.Read(salt); err != nil {
		return "", err
	}

	hash := argon2.IDKey([]byte(password), salt, cfg.Iterations, cfg.Memory, cfg.Parallelism, cfg.KeyLen)

	encodedHash := fmt.Sprintf("$argon2id$v=19$m=%d,t=%d,p=%d$%s$%s",
		cfg.Memory, cfg.Iterations, cfg.Parallelism,
		base64.RawStdEncoding.EncodeToString(salt),
		base64.RawStdEncoding.EncodeToString(hash))

	return encodedHash, nil
}

func VerifyPassword(password, encodedHash string) (bool, error) {
	parts := strings.Split(encodedHash, "$")
	if len(parts) != 6 {
		return false, fmt.Errorf("invalid hash format")
	}

	var version int
	_, err := fmt.Sscanf(parts[2], "v=%d", &version)
	if err != nil {
		return false, err
	}

	var memory, iterations uint32
	var parallelism uint8
	_, err = fmt.Sscanf(parts[3], "m=%d,t=%d,p=%d", &memory, &iterations, &parallelism)
	if err != nil {
		return false, err
	}

	salt, err := base64.RawStdEncoding.DecodeString(parts[4])
	if err != nil {
		return false, err
	}

	hash, err := base64.RawStdEncoding.DecodeString(parts[5])
	if err != nil {
		return false, err
	}

	computedHash := argon2.IDKey([]byte(password), salt, iterations, memory, parallelism, uint32(len(hash)))

	return subtle.ConstantTimeCompare(hash, computedHash) == 1, nil
}

func parseEnvUint32(key string, defaultVal uint32) uint32 {
	s := os.Getenv(key)
	if s == "" {
		return defaultVal
	}
	v, err := strconv.ParseUint(s, 10, 32)
	if err != nil {
		return defaultVal
	}
	return uint32(v)
}

func parseEnvUint8(key string, defaultVal uint8) uint8 {
	s := os.Getenv(key)
	if s == "" {
		return defaultVal
	}
	v, err := strconv.ParseUint(s, 10, 8)
	if err != nil {
		return defaultVal
	}
	return uint8(v)
}
