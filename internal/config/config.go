package config

import (
	"fmt"
	"os"
	"time"
)

type Config struct {
	Build           string
	Desc            string
	APIHost         string
	DebugHost       string
	OTLPHost        string
	ReadTimeout     time.Duration
	WriteTimeout    time.Duration
	IdleTimeout     time.Duration
	ShutdownTimeout time.Duration
}

func LoadConfig(build string) (*Config, error) {
	readTimeout, err := loadDuration(readTimeoutKey, defaultReadTimeout)
	if err != nil {
		return nil, err
	}
	writeTimeout, err := loadDuration(writeTimeoutKey, defaultWriteTimeout)
	if err != nil {
		return nil, err
	}
	idleTimeout, err := loadDuration(idleTimeoutKey, defaultIdleTimeout)
	if err != nil {
		return nil, err
	}
	shutdownTimeout, err := loadDuration(shutdownTimeoutKey, defaultShutdownTimeout)
	if err != nil {
		return nil, err
	}
	otlpHost, ok := os.LookupEnv(otlpHostKey)
	if !ok {
		return nil, fmt.Errorf("environment variable %s is required: %w", otlpHostKey, errMissingEnvVar)
	}

	return &Config{
		Build:           build,
		Desc:            "",
		APIHost:         getEnv(apiHostKey, defaultApiHost),
		DebugHost:       getEnv(debugHostKey, defaultDebugHost),
		OTLPHost:        otlpHost,
		ReadTimeout:     readTimeout,
		WriteTimeout:    writeTimeout,
		IdleTimeout:     idleTimeout,
		ShutdownTimeout: shutdownTimeout,
	}, nil
}

func getEnv(key, fallback string) string {
	if value, ok := os.LookupEnv(key); ok {
		return value
	}
	return fallback
}

func loadDuration(key string, defaultValue time.Duration) (time.Duration, error) {
	value, ok := os.LookupEnv(key)
	if !ok {
		return defaultValue, nil
	}
	duration, err := time.ParseDuration(value)
	if err != nil {
		return 0, fmt.Errorf("invalid duration for %s: %w", key, errInvalidDuration)
	}
	return duration, nil
}
