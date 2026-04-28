package redis

import (
	"os"
)

// NewClientFromEnv creates a *Client from environment variables.
// Priority: ANTCLAW_REDIS_URL > REDIS_URL > ANTCLAW_REDIS_HOST:ANTCLAW_REDIS_PORT.
func NewClientFromEnv() *Client {
	addr := os.Getenv("ANTCLAW_REDIS_URL")
	if addr == "" {
		addr = os.Getenv("REDIS_URL")
	}
	if addr == "" {
		host := getenv("ANTCLAW_REDIS_HOST", "localhost")
		port := getenv("ANTCLAW_REDIS_PORT", "6379")
		addr = host + ":" + port
	}
	password := os.Getenv("ANTCLAW_REDIS_PASSWORD")
	return NewClient(addr, password, 0)
}

func getenv(key, fallback string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return fallback
}
