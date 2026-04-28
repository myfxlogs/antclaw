// Package config provides configuration management for ANTCLAW_* environment variables.
// Supports dynamic updates from web admin interface.
package config

import (
	"os"
	"sync"
)

// AppConfig holds dynamic configuration with thread-safe access.
type AppConfig struct {
	mu     sync.RWMutex
	values map[string]string
}

// NewAppConfig creates a new config instance, loading from environment variables.
func NewAppConfig() *AppConfig {
	c := &AppConfig{values: make(map[string]string)}
	// Load from environment variables with ANTCLAW_ prefix
	if v := os.Getenv("ANTCLAW_FRED_API_KEY"); v != "" {
		c.values["fred_api_key"] = v
	}
	if v := os.Getenv("ANTCLAW_CALENDAR_API_URL"); v != "" {
		c.values["calendar_api_url"] = v
	}
	return c
}

// NewAppConfigWithDefaults creates config with explicit default values.
func NewAppConfigWithDefaults(defaults map[string]string) *AppConfig {
	c := NewAppConfig()
	for k, v := range defaults {
		c.values[k] = v
	}
	return c
}

// Get retrieves a config value by key.
func (c *AppConfig) Get(key string) string {
	c.mu.RLock()
	defer c.mu.RUnlock()
	return c.values[key]
}

// Set updates a config value (thread-safe, callable from admin UI).
func (c *AppConfig) Set(key, value string) {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.values[key] = value
}

// All returns a copy of all config values.
func (c *AppConfig) All() map[string]string {
	c.mu.RLock()
	defer c.mu.RUnlock()
	m := make(map[string]string, len(c.values))
	for k, v := range c.values {
		m[k] = v
	}
	return m
}
