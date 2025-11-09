package config

import (
	"log"
	"os"
	"strconv"
	"time"
)

// Config holds all application configuration
type Config struct {
	// Server settings
	ListenAddr string
	EnableTLS  bool
	TLSDomain  string

	// ACME/TLS settings (optional, for production)
	CloudflareAPIToken string
	CertCacheDir       string

	// Storage settings
	StorageType string // "memory" or "valkey"
	ValkeyAddr  string

	// Auth settings
	KeyProviderURL      string
	KeyRefreshInterval  time.Duration
	SQLitePath          string

	// API settings
	APIKeyPrefix string

	// Feature flags
	EnableValidation bool
}

// Load reads configuration from environment variables
func Load() *Config {
	return &Config{
		ListenAddr: getEnv("LISTEN_ADDR", ":8080"),
		EnableTLS:  getBoolEnv("ENABLE_TLS", false),
		TLSDomain:  getEnv("TLS_DOMAIN", ""),

		CloudflareAPIToken: getEnv("CLOUDFLARE_API_TOKEN", ""),
		CertCacheDir:       getEnv("CERT_CACHE_DIR", "./.certmagic"),

		StorageType: getEnv("STORAGE_TYPE", "memory"),
		ValkeyAddr:  getEnv("VALKEY_ADDR", "localhost:6379"),

		KeyProviderURL:     getEnv("KEY_PROVIDER_URL", ""),
		KeyRefreshInterval: getDurationEnv("KEY_REFRESH_INTERVAL", 10*time.Minute),
		SQLitePath:         getEnv("SQLITE_PATH", "./kestrel.db"),

		APIKeyPrefix: getEnv("API_KEY_PREFIX", "kestrel_"),

		EnableValidation: getBoolEnv("ENABLE_VALIDATION", true),
	}
}

// Validate checks if the configuration is valid
func (c *Config) Validate() error {
	if c.EnableTLS {
		if c.TLSDomain == "" {
			log.Fatal("TLS_DOMAIN is required when ENABLE_TLS is true")
		}
		if c.CloudflareAPIToken == "" {
			log.Fatal("CLOUDFLARE_API_TOKEN is required when ENABLE_TLS is true")
		}
	}

	if c.StorageType != "memory" && c.StorageType != "valkey" {
		log.Fatal("STORAGE_TYPE must be 'memory' or 'valkey'")
	}

	return nil
}

// Helper functions
func getEnv(key, defaultVal string) string {
	if val := os.Getenv(key); val != "" {
		return val
	}
	return defaultVal
}

func getBoolEnv(key string, defaultVal bool) bool {
	val := os.Getenv(key)
	if val == "" {
		return defaultVal
	}
	b, err := strconv.ParseBool(val)
	if err != nil {
		return defaultVal
	}
	return b
}

func getDurationEnv(key string, defaultVal time.Duration) time.Duration {
	val := os.Getenv(key)
	if val == "" {
		return defaultVal
	}
	d, err := time.ParseDuration(val)
	if err != nil {
		return defaultVal
	}
	return d
}
