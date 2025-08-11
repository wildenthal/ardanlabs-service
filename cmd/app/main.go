package main

import (
	"context"
	"expvar"
	"fmt"
	"log/slog"
	"net"
	"net/http"
	"os"
	"os/signal"
	"runtime"
	"time"

	"github.com/caarlos0/env"
	"github.com/wildenthal/ardanlabs-service/internal/api"
	"github.com/wildenthal/ardanlabs-service/internal/middleware"
	"github.com/wildenthal/ardanlabs-service/pkg/debug"
	"github.com/wildenthal/ardanlabs-service/pkg/logging"
	"go.opentelemetry.io/contrib/instrumentation/net/http/otelhttp"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/exporters/otlp/otlpmetric/otlpmetricgrpc"
	"go.opentelemetry.io/otel/exporters/otlp/otlptrace/otlptracegrpc"
	"go.opentelemetry.io/otel/sdk/metric"
	"go.opentelemetry.io/otel/sdk/resource"
	"go.opentelemetry.io/otel/sdk/trace"
	semconv "go.opentelemetry.io/otel/semconv/v1.30.0"
)

var (
	serviceName = "app"
	build       = "develop"
)

func main() {
	logger := slog.New(logging.NewTraceHandler(slog.NewJSONHandler(os.Stdout, nil)))
	ctx := context.Background()

	if err := run(ctx, logger); err != nil {
		logger.ErrorContext(ctx, "Startup failed", "error", err)
		os.Exit(1)
	}
}

func run(ctx context.Context, logger *slog.Logger) error {
	logger.InfoContext(ctx, "Starting up", "GOMAXPROCS", runtime.GOMAXPROCS(0), "build", build)

	// Load configuration
	type config struct {
		APIHost         string        `env:"API_HOST"`
		DebugHost       string        `env:"DEBUG_HOST"`
		ReadTimeout     time.Duration `env:"READ_TIMEOUT"`
		WriteTimeout    time.Duration `env:"WRITE_TIMEOUT"`
		IdleTimeout     time.Duration `env:"IDLE_TIMEOUT"`
		ShutdownTimeout time.Duration `env:"SHUTDOWN_TIMEOUT"`
	}
	var cfg config
	err := env.Parse(&cfg)
	if err != nil {
		return fmt.Errorf("could not load configuration: %w", err)
	}
	expvar.NewString("build").Set(build)

	// Start debug service
	go func() {
		logger.InfoContext(ctx, "Starting debug service", "host", cfg.DebugHost)

		if err := http.ListenAndServe(cfg.DebugHost, debug.Mux()); err != nil {
			logger.ErrorContext(ctx, "Debug service failed", "error", err)
		}
	}()

	// Start API service
	logger.InfoContext(ctx, "Starting API service", "host", cfg.APIHost)

	ctx, stop := signal.NotifyContext(ctx, os.Interrupt)
	defer stop()

	// Initialize OpenTelemetry
	res, err := resource.New(ctx,
		resource.WithAttributes(
			semconv.ServiceName(serviceName),
			semconv.ServiceVersion(build),
		),
	)
	if err != nil {
		return fmt.Errorf("failed to create opentelemetry resource: %w", err)
	}
	traceExporter, err := otlptracegrpc.New(ctx)
	if err != nil {
		return fmt.Errorf("failed to create trace exporter: %w", err)
	}

	tracerProvider := trace.NewTracerProvider(
		trace.WithBatcher(traceExporter),
		trace.WithResource(res),
	)
	defer func() {
		shutdownCtx, cancel := context.WithTimeout(ctx, cfg.ShutdownTimeout)
		defer cancel()
		if err := tracerProvider.Shutdown(shutdownCtx); err != nil {
			logger.ErrorContext(ctx, "failed to shutdown tracer provider", "error", err)
		}
	}()
	otel.SetTracerProvider(tracerProvider)

	meterExporter, err := otlpmetricgrpc.New(ctx)
	if err != nil {
		return fmt.Errorf("failed to create meter exporter: %w", err)
	}

	meterProvider := metric.NewMeterProvider(
		metric.WithReader(metric.NewPeriodicReader(meterExporter)),
		metric.WithResource(res),
	)
	defer func() {
		shutdownCtx, cancel := context.WithTimeout(ctx, cfg.ShutdownTimeout)
		defer cancel()
		if err := meterProvider.Shutdown(shutdownCtx); err != nil {
			logger.ErrorContext(ctx, "failed to shutdown meter provider", "error", err)
		}
	}()
	otel.SetMeterProvider(meterProvider)
	meter := meterProvider.Meter(serviceName)

	// Initialize middleware and set up the HTTP request multiplexer
	mux := http.NewServeMux()
	m, err := middleware.New(logger, meter)
	if err != nil {
		return fmt.Errorf("could not create middleware: %w", err)
	}
	c, err := api.NewHTTPController(logger, meter)
	if err != nil {
		return fmt.Errorf("could not create HTTP controller: %w", err)
	}

	// Helper to tag routes with OTEL spans
	handleFunc := func(pattern string, handlerFunc func(http.ResponseWriter, *http.Request)) {
		handler := otelhttp.WithRouteTag(pattern, http.HandlerFunc(handlerFunc))
		mux.Handle(pattern, handler)
	}

	// Register routes
	handleFunc("GET /", c.StatusOKHandler)
	handleFunc("GET /liveness", c.StatusOKHandler)
	handleFunc("GET /readiness", c.StatusOKHandler)
	handleFunc("GET /panic", c.PanicHandler)

	// Apply middleware and OTEL tracing
	handler := m.HTTPMiddleware(mux)
	handler = otelhttp.NewHandler(handler, "/")

	server := http.Server{
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
		serverErrors <- server.ListenAndServe()
	}()

	// Shutdown
	select {
	case err := <-serverErrors:
		return fmt.Errorf("server error: %w", err)
	case <-ctx.Done():
		logger.InfoContext(ctx, "Received shutdown signal")
		shutdownStart := time.Now()
		defer func() {
			logger.InfoContext(ctx, "Shutdown complete", "took", time.Since(shutdownStart))
		}()

		ctx, cancel := context.WithTimeout(ctx, cfg.ShutdownTimeout)
		defer cancel()

		if err := server.Shutdown(ctx); err != nil {
			if closeErr := server.Close(); closeErr != nil {
				logger.ErrorContext(ctx, "failed to force close server", "error", closeErr)
			}
			return fmt.Errorf("could not shutdown server: %w", err)
		}
	}

	return nil
}
