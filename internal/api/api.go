package api

import (
	"encoding/json"
	"log/slog"
	"net/http"

	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/metric"
)

type httpController struct {
	logger    *slog.Logger
	meter     metric.Meter
	okCounter metric.Int64Counter
}

func NewHTTPController(
	logger *slog.Logger,
	meter metric.Meter,
) (*httpController, error) {
	// Create a counter for the number of OK responses
	okCounter, err := meter.Int64Counter("http.ok.responses", metric.WithDescription("Counts the number of OK responses"))
	if err != nil {
		return nil, err
	}
	return &httpController{
		logger:    logger,
		meter:     meter,
		okCounter: okCounter,
	}, nil
}

func (c *httpController) StatusOKHandler(w http.ResponseWriter, r *http.Request) {
	data, err := json.Marshal(map[string]string{"Status": "OK"})
	if err != nil {
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	_, err = w.Write(data)
	if err != nil {
		c.logger.ErrorContext(r.Context(), "failed to write response", "error", err)
	}

	// Increment the OK response counter
	ctx := r.Context()
	attrs := []attribute.KeyValue{
		attribute.String("method", r.Method),
		attribute.String("path", r.URL.Path),
	}
	c.okCounter.Add(ctx, 1, metric.WithAttributes(attrs...))
}

// PanicHandler is used to intentionally trigger a panic for testing middleware recovery mechanisms.
func (c *httpController) PanicHandler(w http.ResponseWriter, r *http.Request) {
	panic("This is a panic")
}
