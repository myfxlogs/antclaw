// Package crypto 提供数据源敏感配置的存储级加密能力。
//
// 方案：Argon2id（KDF）+ AES-256-GCM（认证加密）。
//   - master_key：进程级，由环境变量 ANTCLAW_SECRET_MASTER_KEY 注入（base64 编码，至少 32 字节熵）。
//   - 每条记录独立的 16B salt 与 12B nonce；
//   - 派生密钥 dk = Argon2id(master_key, salt, time=3, memory=64MiB, threads=4, keyLen=32)；
//   - 密文 = AES-256-GCM(dk, nonce).Seal(plaintext)；密文自带认证标签。
//
// 设计要点：稳健（per-record salt 防 rainbow table）、健壮（GCM 认证防篡改）、
// 实用（Worker 可正常解密获得明文用于调用外部 API）。
package crypto

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"encoding/base64"
	"errors"
	"fmt"
	"os"
	"path/filepath"

	"golang.org/x/crypto/argon2"
)

// Argon2 参数：参考 OWASP 推荐（生产环境）。
// 性能：单次 ~70ms@4 核；管理端低频写入可接受。
const (
	argonTime    = 3
	argonMemory  = 64 * 1024 // KiB → 64 MiB
	argonThreads = 4
	argonKeyLen  = 32
	saltLen      = 16
	nonceLen     = 12 // GCM 标准
)

// SecretBox 是单次加解密的可复用上下文，持有 master key。
type SecretBox struct {
	masterKey []byte
}

// NewSecretBox 从 base64 编码字符串解析 master key。要求至少 32 字节。
func NewSecretBox(masterB64 string) (*SecretBox, error) {
	if masterB64 == "" {
		return nil, errors.New("ANTCLAW_SECRET_MASTER_KEY is required")
	}
	key, err := base64.StdEncoding.DecodeString(masterB64)
	if err != nil {
		return nil, fmt.Errorf("decode master key: %w", err)
	}
	if len(key) < 32 {
		return nil, fmt.Errorf("master key too short: %d bytes (need ≥32)", len(key))
	}
	return &SecretBox{masterKey: key}, nil
}

// Seal 加密 plaintext，返回 (ciphertext, salt, nonce)。
// 三个返回值都需要持久化以便后续 Open。
func (s *SecretBox) Seal(plaintext []byte) (ciphertext, salt, nonce []byte, err error) {
	salt = make([]byte, saltLen)
	if _, err = rand.Read(salt); err != nil {
		return nil, nil, nil, fmt.Errorf("salt: %w", err)
	}
	nonce = make([]byte, nonceLen)
	if _, err = rand.Read(nonce); err != nil {
		return nil, nil, nil, fmt.Errorf("nonce: %w", err)
	}

	dk := argon2.IDKey(s.masterKey, salt, argonTime, argonMemory, argonThreads, argonKeyLen)
	block, err := aes.NewCipher(dk)
	if err != nil {
		return nil, nil, nil, fmt.Errorf("aes: %w", err)
	}
	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return nil, nil, nil, fmt.Errorf("gcm: %w", err)
	}
	ciphertext = gcm.Seal(nil, nonce, plaintext, nil)
	return ciphertext, salt, nonce, nil
}

// Open 用 (ciphertext, salt, nonce) 解密；任何字段被篡改都会触发认证失败。
const defaultKeyPath = "/data/antclaw_master_key"

// LoadOrCreateMasterKey 获取主密钥：环境变量优先（用于迁移场景），
// 其次读取持久化文件，最后自动生成并保存。确保重启不丢失。
func LoadOrCreateMasterKey() (string, error) {
	if k := os.Getenv("ANTCLAW_SECRET_MASTER_KEY"); k != "" {
		return k, nil
	}
	path := os.Getenv("ANTCLAW_MASTER_KEY_PATH")
	if path == "" {
		path = defaultKeyPath
	}
	data, err := os.ReadFile(path)
	if err == nil {
		return string(data), nil
	}
	if !os.IsNotExist(err) {
		return "", fmt.Errorf("read master key: %w", err)
	}
	// 首次运行：生成并保存
	key := make([]byte, 32)
	if _, err := rand.Read(key); err != nil {
		return "", fmt.Errorf("generate master key: %w", err)
	}
	b64 := base64.StdEncoding.EncodeToString(key)
	if err := os.MkdirAll(filepath.Dir(path), 0700); err != nil {
		return "", fmt.Errorf("mkdir for key: %w", err)
	}
	if err := os.WriteFile(path, []byte(b64), 0600); err != nil {
		return "", fmt.Errorf("write master key: %w", err)
	}
	return b64, nil
}

func (s *SecretBox) Open(ciphertext, salt, nonce []byte) ([]byte, error) {
	if len(salt) != saltLen || len(nonce) != nonceLen {
		return nil, fmt.Errorf("invalid salt/nonce length")
	}
	dk := argon2.IDKey(s.masterKey, salt, argonTime, argonMemory, argonThreads, argonKeyLen)
	block, err := aes.NewCipher(dk)
	if err != nil {
		return nil, fmt.Errorf("aes: %w", err)
	}
	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return nil, fmt.Errorf("gcm: %w", err)
	}
	plaintext, err := gcm.Open(nil, nonce, ciphertext, nil)
	if err != nil {
		return nil, fmt.Errorf("decrypt failed: %w", err)
	}
	return plaintext, nil
}
