// Package configs defines runtime configuration of the application
package configs

import (
	"log/slog"
	"os"
	"strconv"
)

// AppConfig is a struct holding application config that comes from environment variables
type AppConfig struct {
	EnableWeb bool
	EnableAPI bool
	HTTPPort  int
	DataDir   string
}

func NewConfig() AppConfig {
	return AppConfig{
		EnableWeb: EnvBool("ENABLE_WEB", true),
		EnableAPI: EnvBool("ENABLE_API", true),
		HTTPPort:  EnvInt("HTTP_PORT", 9090),
		DataDir:   EnvString("DATA_DIR", "/promruo-data"),
	}
}

func EnvBool(key string, fallback bool) bool {
	value, found := os.LookupEnv(key)
	if !found {
		return fallback
	}

	v, err := strconv.ParseBool(value)
	if err != nil {
		slog.Debug("could not parse env var", "key", key, "value", value, "fallback", fallback)
		return fallback
	}
	return v
}

func EnvInt(key string, fallback int) int {
	value, found := os.LookupEnv(key)
	if !found {
		return fallback
	}

	v, err := strconv.Atoi(value)
	if err != nil {
		slog.Debug("could not parse env var", "key", key, "value", value, "fallback", fallback)
		return fallback
	}
	return v
}

func EnvString(key string, fallback string) string {
	if value, found := os.LookupEnv(key); found {
		return value
	}
	return fallback
}
