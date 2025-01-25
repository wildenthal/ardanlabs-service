package main

import (
	"context"
	"expvar"
	"fmt"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"runtime"
	"syscall"
	"time"

	"github.com/wildenthal/ardanlabs-service/pkg/debug"
)

var build = "develop"

func main() {
	logger := slog.New(slog.NewJSONHandler(os.Stdout, nil))
	ctx := context.Background()

	if err := run(ctx, logger); err != nil {
		logger.Error("Startup failed", "error", err)
		os.Exit(1)
	}
}

func run(ctx context.Context, logger *slog.Logger) error {
	logger.InfoContext(ctx, "Starting up", "GOMAXPROCS", runtime.GOMAXPROCS(0), "build", build)

	// Load configuration
	cfg, err := loadConfig()
	if err != nil {
		return fmt.Errorf("could not load configuration: %w", err)
	}
	expvar.NewString("build").Set(cfg.Build)

	// Start debug service
	go func() {
		logger.InfoContext(ctx, "Starting debug service", "host", cfg.DebugHost)

		if err := http.ListenAndServe(cfg.DebugHost, debug.Mux()); err != nil {
			logger.ErrorContext(ctx, "Debug service failed", "error", err)
		}
	}()

	// Start API service
	logger.InfoContext(ctx, "Starting API service", "host", cfg.APIHost)

	shutdown := make(chan os.Signal, 1)
	signal.Notify(shutdown, syscall.SIGINT, syscall.SIGTERM)

	api := http.Server{
		Addr:         cfg.APIHost,
		Handler:      nil,
		ReadTimeout:  cfg.ReadTimeout,
		WriteTimeout: cfg.WriteTimeout,
		IdleTimeout:  cfg.IdleTimeout,
		ErrorLog:     slog.NewLogLogger(slog.NewJSONHandler(os.Stderr, nil), slog.LevelError),
	}
	serverErrors := make(chan error, 1)

	go func() {
		serverErrors <- api.ListenAndServe()
	}()

	// Shutdown
	select {
	case err := <-serverErrors:
		return fmt.Errorf("server error: %w", err)
	case sig := <-shutdown:
		logger.InfoContext(ctx, "Received shutdown signal", "signal", sig)
		defer logger.InfoContext(ctx, "Shutdown complete", "took", time.Since(time.Now()))

		ctx, cancel := context.WithTimeout(ctx, cfg.ShutdownTimeout)
		defer cancel()

		if err := api.Shutdown(ctx); err != nil {
			api.Close()
			return fmt.Errorf("could not shutdown server: %w", err)
		}
	}

	return nil
}

type config struct {
	Build           string
	Desc            string
	APIHost         string
	DebugHost       string
	ReadTimeout     time.Duration
	WriteTimeout    time.Duration
	IdleTimeout     time.Duration
	ShutdownTimeout time.Duration
}

func loadConfig() (*config, error) {
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

	return &config{
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
