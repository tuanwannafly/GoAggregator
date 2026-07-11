package config

import (
	"log/slog"
	"os"
	"strconv"
	"strings"
)

const (
	defaultPort              = "8081"
	defaultLogLevel          = "info"
	defaultProviderTimeoutMs = 3000
	defaultProviderHosts     = "http://provider-fast:8080,http://provider-slow:8080,http://provider-flaky:8080,http://provider-timeout:8080,http://provider-down:8080"
)

type Config struct {
	Port              string
	LogLevel          string
	ProviderTimeoutMs int
	ProviderHosts     []string
}

func Load() *Config {
	cfg := &Config{
		Port:              getEnv("PORT", defaultPort),
		LogLevel:          strings.ToLower(getEnv("LOG_LEVEL", defaultLogLevel)),
		ProviderTimeoutMs: getEnvInt("PROVIDER_TIMEOUT_MS", defaultProviderTimeoutMs),
		ProviderHosts:     getEnvSlice("PROVIDER_HOSTS", defaultProviderHosts),
	}
	return cfg
}

func (c *Config) SlogLevel() slog.Level {
	switch c.LogLevel {
	case "debug":
		return slog.LevelDebug
	case "warn":
		return slog.LevelWarn
	case "error":
		return slog.LevelError
	default:
		return slog.LevelInfo
	}
}

func getEnv(key, defaultVal string) string {
	if val, ok := os.LookupEnv(key); ok && val != "" {
		return val
	}
	return defaultVal
}

func getEnvInt(key string, defaultVal int) int {
	if val, ok := os.LookupEnv(key); ok && val != "" {
		if n, err := strconv.Atoi(val); err == nil {
			return n
		}
	}
	return defaultVal
}

func getEnvSlice(key, defaultVal string) []string {
	raw := getEnv(key, defaultVal)
	parts := strings.Split(raw, ",")
	values := make([]string, 0, len(parts))
	for _, part := range parts {
		value := strings.TrimSpace(part)
		if value != "" {
			values = append(values, value)
		}
	}
	return values
}
