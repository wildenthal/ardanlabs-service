package config

import (
	"fmt"
	"os"
	"time"
)

const (
	missing   = "environment variable %s is required"
	loadError = "failed to load environment variable %s: %w"
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
	readTimeout, err := loadDuration("READ_TIMEOUT", 5*time.Second)
	if err != nil {
		return nil, err
	}
	writeTimeout, err := loadDuration("WRITE_TIMEOUT", 10*time.Second)
	if err != nil {
		return nil, err
	}
	idleTimeout, err := loadDuration("IDLE_TIMEOUT", 120*time.Second)
	if err != nil {
		return nil, err
	}
	shutdownTimeout, err := loadDuration("SHUTDOWN_TIMEOUT", 5*time.Second)
	if err != nil {
		return nil, err
	}
	otlpHost := getEnv("OTEL_EXPORTER_OTLP_ENDPOINT", "")
	if otlpHost == "" {
		return nil, fmt.Errorf(missing, "OTLP_HOST")
	}

	return &Config{
		Build:           build,
		Desc:            "",
		APIHost:         getEnv("API_HOST", "0.0.0.0:3000"),
		DebugHost:       getEnv("DEBUG_HOST", "0.0.0.0:3010"),
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
	value := getEnv(key, defaultValue.String())
	duration, err := time.ParseDuration(value)
	if err != nil {
		return 0, fmt.Errorf(loadError, key, err)
	}
	return duration, nil
}
