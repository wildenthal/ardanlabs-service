package config

import (
	"errors"
	"time"
)

// key names for environment variables
const (
	readTimeoutKey     = "READ_TIMEOUT"
	writeTimeoutKey    = "WRITE_TIMEOUT"
	idleTimeoutKey     = "IDLE_TIMEOUT"
	shutdownTimeoutKey = "SHUTDOWN_TIMEOUT"
	otlpHostKey        = "OTEL_EXPORTER_OTLP_ENDPOINT"
	apiHostKey         = "API_HOST"
	debugHostKey       = "DEBUG_HOST"
)

// default values for environment variables
const (
	defaultReadTimeout     = 5 * time.Second
	defaultWriteTimeout    = 10 * time.Second
	defaultIdleTimeout     = 120 * time.Second
	defaultShutdownTimeout = 5 * time.Second
	defaultOTLPHost        = ""
	defaultApiHost         = "0.0.0.0:3000"
	defaultDebugHost       = "0.0.0.0:3010"
)

// error messages
var (
	errInvalidDuration = errors.New("invalid duration")
	errMissingEnvVar   = errors.New("missing environment variable")
)
