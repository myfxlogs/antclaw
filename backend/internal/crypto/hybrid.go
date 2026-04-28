package crypto

import (
	"crypto/aes"
	"crypto/cipher"
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
)

// HybridEnvelope 是前端 RSA+AES 混合加密的请求载荷格式。
// 字段全部使用 base64 (StdEncoding) 编码，方便 JSON 传输。
type HybridEnvelope struct {
	KeyEnc     string `json:"key_enc"`    // RSA-OAEP(server pub).encrypt(aesKey)
	IV         string `json:"iv"`         // 12 bytes
	Ciphertext string `json:"ciphertext"` // AES-256-GCM(aesKey).seal(payload)
}

// Decrypt 用 RSA 私钥解出 AES 会话密钥，再用 AES-GCM 解出 payload。
// 若任意步骤失败（含认证失败），返回错误。
func (m *RSAManager) Decrypt(env HybridEnvelope) ([]byte, error) {
	keyEnc, err := base64.StdEncoding.DecodeString(env.KeyEnc)
	if err != nil {
		return nil, fmt.Errorf("decode key_enc: %w", err)
	}
	iv, err := base64.StdEncoding.DecodeString(env.IV)
	if err != nil {
		return nil, fmt.Errorf("decode iv: %w", err)
	}
	ct, err := base64.StdEncoding.DecodeString(env.Ciphertext)
	if err != nil {
		return nil, fmt.Errorf("decode ciphertext: %w", err)
	}
	if len(iv) != 12 {
		return nil, errors.New("iv must be 12 bytes")
	}

	aesKey, err := m.DecryptOAEP(keyEnc)
	if err != nil {
		return nil, fmt.Errorf("rsa decrypt: %w", err)
	}
	if len(aesKey) != 32 {
		return nil, fmt.Errorf("aes key must be 32 bytes, got %d", len(aesKey))
	}

	block, err := aes.NewCipher(aesKey)
	if err != nil {
		return nil, fmt.Errorf("aes: %w", err)
	}
	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return nil, fmt.Errorf("gcm: %w", err)
	}
	plaintext, err := gcm.Open(nil, iv, ct, nil)
	if err != nil {
		return nil, fmt.Errorf("aes-gcm open: %w", err)
	}
	return plaintext, nil
}

// DecryptJSON 是 Decrypt 的便捷方法，再做一次 JSON unmarshal 到 v。
func (m *RSAManager) DecryptJSON(env HybridEnvelope, v any) error {
	plain, err := m.Decrypt(env)
	if err != nil {
		return err
	}
	if err := json.Unmarshal(plain, v); err != nil {
		return fmt.Errorf("unmarshal payload: %w", err)
	}
	return nil
}
