package api

import (
	"os"
	"time"
)

// Config holds API server configuration from environment.
type Config struct {
	Addr        string
	RSAKeyPath  string
	MT4GWURL    string
	MT5GWURL    string
	ServerReady time.Duration
}

// DefaultConfig returns a Config populated from environment variables.
func DefaultConfig() Config {
	return Config{
		Addr:        ":8080",
		RSAKeyPath:  getEnv("ANTCLAW_RSA_KEY_PATH", "/data/rsa_private.pem"),
		MT4GWURL:    getEnv("MT4_GATEWAY_URL", "http://localhost:8080"),
		MT5GWURL:    getEnv("MT5_GATEWAY_URL", "http://localhost:8080"),
		ServerReady: 5 * time.Second,
	}
}

func getEnv(key, def string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return def
}
