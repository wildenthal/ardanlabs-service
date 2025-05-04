package logging

import (
	"context"
	"log/slog"

	"go.opentelemetry.io/otel/trace"
)

type traceHandler struct {
	slog.Handler
}

func (h *traceHandler) Handle(ctx context.Context, r slog.Record) error {
	span := trace.SpanFromContext(ctx)
	if spanCtx := span.SpanContext(); spanCtx.IsValid() {
		r.Add("trace_id", spanCtx.TraceID().String())
		r.Add("span_id", spanCtx.SpanID().String())
	}
	return h.Handler.Handle(ctx, r)
}

func NewTraceHandler(h slog.Handler) *traceHandler {
	return &traceHandler{Handler: h}
}
