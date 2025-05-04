package config

import (
	"os"
	"time"
)

type Config struct {
	Build           string
	Desc            string
	APIHost         string
	DebugHost       string
	ReadTimeout     time.Duration
	WriteTimeout    time.Duration
	IdleTimeout     time.Duration
	ShutdownTimeout time.Duration
}

func LoadConfig(build string) (*Config, error) {
	readTimeout, err := time.ParseDuration(getEnv("READ_TIMEOUT", "5s"))
	if err != nil {
		return nil, err
	}
	writeTimeout, err := time.ParseDuration(getEnv("WRITE_TIMEOUT", "10s"))
	if err != nil {
		return nil, err
	}
	idleTimeout, err := time.ParseDuration(getEnv("IDLE_TIMEOUT", "120s"))
	if err != nil {
		return nil, err
	}
	shutdownTimeout, err := time.ParseDuration(getEnv("SHUTDOWN_TIMEOUT", "5s"))
	if err != nil {
		return nil, err
	}

	return &Config{
		Build:           build,
		Desc:            "Example service",
		APIHost:         getEnv("API_HOST", "0.0.0.0:3000"),
		DebugHost:       getEnv("DEBUG_HOST", "0.0.0.0:3010"),
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
