package middleware

import (
	"log/slog"
	"net/http"

	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/metric"
)

type middleware struct {
	logger       *slog.Logger
	meter        metric.Meter
	panicCounter metric.Int64Counter
}

func New(
	logger *slog.Logger,
	meter metric.Meter,
) (*middleware, error) {
	panicCounter, err := meter.Int64Counter("http.panic.responses", metric.WithDescription("Counts the number of panic responses"))
	if err != nil {
		return nil, err
	}
	return &middleware{
		logger:       logger,
		meter:        meter,
		panicCounter: panicCounter,
	}, nil
}

func (m *middleware) HTTPMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		defer func() {
			if err := recover(); err != nil {
				m.logger.ErrorContext(r.Context(), "Recovered from panic", "error", err)
				// Increment the panic response counter
				ctx := r.Context()
				attrs := []attribute.KeyValue{
					attribute.String("method", r.Method),
					attribute.String("path", r.URL.Path),
				}
				m.panicCounter.Add(ctx, 1, metric.WithAttributes(attrs...))
				http.Error(w, "Internal Server Error", http.StatusInternalServerError)
			}
		}()
		next.ServeHTTP(w, r)
	})
}
