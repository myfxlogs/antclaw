package crypto

import (
	"crypto/rand"
	"crypto/rsa"
	"crypto/sha256"
	"crypto/x509"
	"encoding/pem"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"sync"
)

// RSAManager 管理一对 RSA-2048 密钥对，用于前端 → 后端的混合加密。
//   - 启动时若指定文件存在，则加载；否则生成并持久化。
//   - 公钥以 PEM 形式暴露给前端（Connect：CryptoService.GetCryptoPublicKey）。
//   - 私钥仅在内存与本地文件中存在，权限 0600。
type RSAManager struct {
	priv      *rsa.PrivateKey
	publicPEM []byte
	once      sync.Once
}

// LoadOrCreateRSA 从 path 加载私钥；不存在则生成新的并写入。
func LoadOrCreateRSA(path string) (*RSAManager, error) {
	if path == "" {
		return nil, errors.New("rsa key path is required")
	}
	if err := os.MkdirAll(filepath.Dir(path), 0o700); err != nil {
		return nil, fmt.Errorf("mkdir: %w", err)
	}

	if data, err := os.ReadFile(path); err == nil {
		block, _ := pem.Decode(data)
		if block == nil {
			return nil, errors.New("invalid pem in rsa key file")
		}
		priv, err := x509.ParsePKCS8PrivateKey(block.Bytes)
		if err != nil {
			// 兼容旧 PKCS1
			if k1, e1 := x509.ParsePKCS1PrivateKey(block.Bytes); e1 == nil {
				priv = k1
			} else {
				return nil, fmt.Errorf("parse private key: %w", err)
			}
		}
		rsaPriv, ok := priv.(*rsa.PrivateKey)
		if !ok {
			return nil, errors.New("not an rsa private key")
		}
		return newRSAManager(rsaPriv)
	}

	// 生成新密钥
	priv, err := rsa.GenerateKey(rand.Reader, 2048)
	if err != nil {
		return nil, fmt.Errorf("generate rsa: %w", err)
	}
	der, err := x509.MarshalPKCS8PrivateKey(priv)
	if err != nil {
		return nil, fmt.Errorf("marshal private: %w", err)
	}
	pemBytes := pem.EncodeToMemory(&pem.Block{Type: "PRIVATE KEY", Bytes: der})
	if err := os.WriteFile(path, pemBytes, 0o600); err != nil {
		return nil, fmt.Errorf("write private: %w", err)
	}
	return newRSAManager(priv)
}

func newRSAManager(priv *rsa.PrivateKey) (*RSAManager, error) {
	pubDer, err := x509.MarshalPKIXPublicKey(&priv.PublicKey)
	if err != nil {
		return nil, fmt.Errorf("marshal public: %w", err)
	}
	pubPEM := pem.EncodeToMemory(&pem.Block{Type: "PUBLIC KEY", Bytes: pubDer})
	return &RSAManager{priv: priv, publicPEM: pubPEM}, nil
}

// PublicPEM 返回 PEM 格式公钥（可直接交给浏览器 SubtleCrypto 使用）。
func (m *RSAManager) PublicPEM() []byte { return m.publicPEM }

// DecryptOAEP 使用私钥解密 RSA-OAEP(SHA-256) 密文。
func (m *RSAManager) DecryptOAEP(ciphertext []byte) ([]byte, error) {
	return rsa.DecryptOAEP(sha256.New(), rand.Reader, m.priv, ciphertext, nil)
}
