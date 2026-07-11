package config

import (
	"log/slog"
	"os"
	"strconv"
	"strings"
)

const (
	defaultPort                   = "8081"
	defaultLogLevel               = "info"
	defaultProviderTimeoutMs      = 3000
	defaultProviderHosts          = "http://provider-fast:8080,http://provider-slow:8080,http://provider-flaky:8080,http://provider-timeout:8080,http://provider-down:8080"
	defaultBreakerCooldownMs      = 30000
	defaultBreakerHalfOpenCalls   = 1
	defaultProviderRateLimitRPS   = 10
	defaultProviderRateLimitBurst = 10
	defaultRedisAddr              = "redis:6379"
	defaultRedisDB                = 0
	defaultCacheTTLSeconds        = 45
)

type Config struct {
	Port                   string
	LogLevel               string
	ProviderTimeoutMs      int
	ProviderHosts          []string
	BreakerCooldownMs      int
	BreakerHalfOpenCalls   int
	ProviderRateLimitRPS   int
	ProviderRateLimitBurst int
	RedisAddr              string
	RedisDB                int
	CacheTTLSeconds        int
}

func Load() *Config {
	cfg := &Config{
		Port:                   getEnv("PORT", defaultPort),
		LogLevel:               strings.ToLower(getEnv("LOG_LEVEL", defaultLogLevel)),
		ProviderTimeoutMs:      getEnvInt("PROVIDER_TIMEOUT_MS", defaultProviderTimeoutMs),
		ProviderHosts:          getEnvSlice("PROVIDER_HOSTS", defaultProviderHosts),
		BreakerCooldownMs:      getEnvInt("BREAKER_COOLDOWN_MS", defaultBreakerCooldownMs),
		BreakerHalfOpenCalls:   getEnvInt("BREAKER_HALF_OPEN_CALLS", defaultBreakerHalfOpenCalls),
		ProviderRateLimitRPS:   getEnvInt("PROVIDER_RATE_LIMIT_RPS", defaultProviderRateLimitRPS),
		ProviderRateLimitBurst: getEnvInt("PROVIDER_RATE_LIMIT_BURST", defaultProviderRateLimitBurst),
		RedisAddr:              getEnv("REDIS_ADDR", defaultRedisAddr),
		RedisDB:                getEnvInt("REDIS_DB", defaultRedisDB),
		CacheTTLSeconds:        getEnvInt("CACHE_TTL_SECONDS", defaultCacheTTLSeconds),
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
