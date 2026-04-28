package auth

import (
	"crypto/ed25519"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/google/uuid"
)

const (
	Issuer     = "antclaw"
	Audience   = "antclaw-api"
	AccessTTL  = 15 * time.Minute
	RefreshTTL = 30 * 24 * time.Hour
)

type TokenType string

const (
	TokenTypeAccess  TokenType = "access"
	TokenTypeRefresh TokenType = "refresh"
)

type Claims struct {
	Issuer          string    `json:"iss"`
	Subject         string    `json:"sub"`
	Audience        string    `json:"aud"`
	IssuedAt        int64     `json:"iat"`
	ExpiresAt       int64     `json:"exp"`
	NotBefore       int64     `json:"nbf"`
	JTI             string    `json:"jti"`
	SessionID       string    `json:"sid"`
	Type            TokenType `json:"typ"`
	Role            string    `json:"role"`
	PasswordVersion int       `json:"pv"`
	Locale          string    `json:"locale"`
}

type JWTKeyPair struct {
	PrivateKey ed25519.PrivateKey
	PublicKey  ed25519.PublicKey
	Kid        string
}

var (
	currentKey *JWTKeyPair
	previousKey *JWTKeyPair
)

func LoadKeys() error {
	privateKeyPEM := os.Getenv("ANTCLAW_JWT_PRIVATE_KEY")
	if privateKeyPEM == "" {
		_, priv, err := ed25519.GenerateKey(nil)
		if err != nil {
			return fmt.Errorf("failed to generate key: %w", err)
		}
		currentKey = &JWTKeyPair{
			PrivateKey: priv,
			PublicKey:  priv.Public().(ed25519.PublicKey),
			Kid:        "current",
		}
		return nil
	}

	privateKey, err := parseEd25519PrivateKey(privateKeyPEM)
	if err != nil {
		return fmt.Errorf("failed to parse private key: %w", err)
	}

	currentKey = &JWTKeyPair{
		PrivateKey: privateKey,
		PublicKey:  privateKey.Public().(ed25519.PublicKey),
		Kid:        "current",
	}

	previousPEM := os.Getenv("ANTCLAW_JWT_PUBLIC_KEY_PREVIOUS")
	if previousPEM != "" {
		publicKey, err := parseEd25519PublicKey(previousPEM)
		if err != nil {
			return fmt.Errorf("failed to parse previous public key: %w", err)
		}
		previousKey = &JWTKeyPair{
			PublicKey: publicKey,
			Kid:       "previous",
		}
	}

	return nil
}

func GenerateAccessToken(userID, sessionID, role, locale string, passwordVersion int) (string, string, error) {
	jti := uuid.New().String()
	now := time.Now()

	claims := Claims{
		Issuer:          Issuer,
		Subject:         userID,
		Audience:        Audience,
		IssuedAt:        now.Unix(),
		ExpiresAt:       now.Add(AccessTTL).Unix(),
		NotBefore:       now.Unix(),
		JTI:             jti,
		SessionID:       sessionID,
		Type:            TokenTypeAccess,
		Role:            role,
		PasswordVersion: passwordVersion,
		Locale:          locale,
	}

	token, err := signToken(claims)
	return token, jti, err
}

func GenerateRefreshToken(userID, sessionID string, passwordVersion int) (string, string, time.Time, error) {
	jti := uuid.New().String()
	now := time.Now()
	expiresAt := now.Add(RefreshTTL)

	claims := Claims{
		Issuer:          Issuer,
		Subject:         userID,
		Audience:        Audience,
		IssuedAt:        now.Unix(),
		ExpiresAt:       expiresAt.Unix(),
		NotBefore:       now.Unix(),
		JTI:             jti,
		SessionID:       sessionID,
		Type:            TokenTypeRefresh,
		Role:            "",
		PasswordVersion: passwordVersion,
		Locale:          "",
	}

	token, err := signToken(claims)
	return token, jti, expiresAt, err
}

func signToken(claims Claims) (string, error) {
	if currentKey == nil {
		return "", fmt.Errorf("JWT keys not loaded")
	}

	header := map[string]interface{}{
		"alg": "EdDSA",
		"typ": "JWT",
		"kid": currentKey.Kid,
	}

	headerJSON, err := json.Marshal(header)
	if err != nil {
		return "", err
	}

	claimsJSON, err := json.Marshal(claims)
	if err != nil {
		return "", err
	}

	headerB64 := base64.RawURLEncoding.EncodeToString(headerJSON)
	claimsB64 := base64.RawURLEncoding.EncodeToString(claimsJSON)

	signingInput := headerB64 + "." + claimsB64
	signature := ed25519.Sign(currentKey.PrivateKey, []byte(signingInput))
	sigB64 := base64.RawURLEncoding.EncodeToString(signature)

	return signingInput + "." + sigB64, nil
}

func ParseToken(token string) (*Claims, error) {
	parts := strings.Split(token, ".")
	if len(parts) != 3 {
		return nil, fmt.Errorf("invalid token format")
	}

	signingInput := parts[0] + "." + parts[1]
	signature, err := base64.RawURLEncoding.DecodeString(parts[2])
	if err != nil {
		return nil, fmt.Errorf("invalid signature encoding")
	}

	headerJSON, err := base64.RawURLEncoding.DecodeString(parts[0])
	if err != nil {
		return nil, fmt.Errorf("invalid header encoding")
	}

	var header struct {
		Kid string `json:"kid"`
	}
	if err := json.Unmarshal(headerJSON, &header); err != nil {
		return nil, fmt.Errorf("invalid header")
	}

	var publicKey ed25519.PublicKey
	if header.Kid == "current" && currentKey != nil {
		publicKey = currentKey.PublicKey
	} else if header.Kid == "previous" && previousKey != nil {
		publicKey = previousKey.PublicKey
	} else {
		return nil, fmt.Errorf("unknown key id")
	}

	if !ed25519.Verify(publicKey, []byte(signingInput), signature) {
		return nil, fmt.Errorf("invalid signature")
	}

	claimsJSON, err := base64.RawURLEncoding.DecodeString(parts[1])
	if err != nil {
		return nil, fmt.Errorf("invalid claims encoding")
	}

	var claims Claims
	if err := json.Unmarshal(claimsJSON, &claims); err != nil {
		return nil, fmt.Errorf("invalid claims")
	}

	return &claims, nil
}

func parseEd25519PrivateKey(pem string) (ed25519.PrivateKey, error) {
	data, err := base64.StdEncoding.DecodeString(pem)
	if err != nil {
		return nil, err
	}
	if len(data) != ed25519.SeedSize {
		return nil, fmt.Errorf("invalid private key size")
	}
	return ed25519.NewKeyFromSeed(data), nil
}

func parseEd25519PublicKey(pem string) (ed25519.PublicKey, error) {
	data, err := base64.StdEncoding.DecodeString(pem)
	if err != nil {
		return nil, err
	}
	if len(data) != ed25519.PublicKeySize {
		return nil, fmt.Errorf("invalid public key size")
	}
	return ed25519.PublicKey(data), nil
}

func ValidateToken(token string, tokenType TokenType) (*Claims, error) {
	claims, err := ParseToken(token)
	if err != nil {
		return nil, err
	}

	now := time.Now().Unix()

	if claims.Issuer != Issuer {
		return nil, fmt.Errorf("invalid issuer")
	}
	if claims.Audience != Audience {
		return nil, fmt.Errorf("invalid audience")
	}
	if claims.ExpiresAt < now {
		return nil, fmt.Errorf("token expired")
	}
	if claims.NotBefore > now {
		return nil, fmt.Errorf("token not yet valid")
	}
	if claims.Type != tokenType {
		return nil, fmt.Errorf("invalid token type")
	}

	return claims, nil
}
