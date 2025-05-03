package main

import (
	"context"
	"encoding/json"
	"expvar"
	"fmt"
	"log/slog"
	"net"
	"net/http"
	"os"
	"os/signal"
	"runtime"
	"time"

	"github.com/wildenthal/ardanlabs-service/pkg/debug"
	"go.opentelemetry.io/contrib/instrumentation/net/http/otelhttp"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/exporters/otlp/otlptrace/otlptracegrpc"
	"go.opentelemetry.io/otel/sdk/resource"
	"go.opentelemetry.io/otel/sdk/trace"
	semconv "go.opentelemetry.io/otel/semconv/v1.30.0"
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

	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt)
	defer stop()

	exp, err := otlptracegrpc.New(
		ctx,
		otlptracegrpc.WithEndpoint("otel-collector:4317"),
		otlptracegrpc.WithInsecure(),
	)
	if err != nil {
		return fmt.Errorf("failed to create trace exporter: %w", err)
	}

	tracerProvider := trace.NewTracerProvider(
		trace.WithBatcher(exp),
		trace.WithResource(resource.NewWithAttributes(
			semconv.SchemaURL,
			semconv.ServiceName("app"),
			semconv.ServiceVersion(build),
		)),
	)
	defer func() {
		if err := tracerProvider.Shutdown(ctx); err != nil {
			logger.ErrorContext(ctx, "failed to shutdown tracer provider", "error", err)
		}
	}()

	otel.SetTracerProvider(tracerProvider)
	mux := http.NewServeMux()
	setUpRouter(mux)
	handler := otelhttp.NewHandler(mux, "/")

	api := http.Server{
		Addr:         cfg.APIHost,
		BaseContext:  func(_ net.Listener) context.Context { return ctx },
		ErrorLog:     slog.NewLogLogger(slog.NewJSONHandler(os.Stderr, nil), slog.LevelError),
		Handler:      handler,
		IdleTimeout:  cfg.IdleTimeout,
		ReadTimeout:  cfg.ReadTimeout,
		WriteTimeout: cfg.WriteTimeout,
	}
	serverErrors := make(chan error, 1)

	go func() {
		serverErrors <- api.ListenAndServe()
	}()

	// Shutdown
	select {
	case err := <-serverErrors:
		return fmt.Errorf("server error: %w", err)
	case <-ctx.Done():
		logger.InfoContext(ctx, "Received shutdown signal")
		defer logger.InfoContext(ctx, "Shutdown complete", "took", time.Since(time.Now()))

		ctx, cancel := context.WithTimeout(ctx, cfg.ShutdownTimeout)
		defer cancel()

		if err := api.Shutdown(ctx); err != nil {
			api.Close()
			return fmt.Errorf("could not shutdown server: %w", err)
		}
		stop()
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

func statusOKHandler(w http.ResponseWriter, r *http.Request) {
	data, err := json.Marshal(map[string]string{"Status": "OK"})
	if err != nil {
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.Write(data)
}

func setUpRouter(mux *http.ServeMux) {
	// handleFunc is a replacement for mux.HandleFunc
	// which enriches the handler's HTTP instrumentation with the pattern as the http.route.
	handleFunc := func(pattern string, handlerFunc func(http.ResponseWriter, *http.Request)) {
		// Configure the "http.route" for the HTTP instrumentation.
		handler := otelhttp.WithRouteTag(pattern, http.HandlerFunc(handlerFunc))
		mux.Handle(pattern, handler)
	}
	handleFunc("GET /liveness", statusOKHandler)
	handleFunc("GET /readiness", statusOKHandler)
	return
}
