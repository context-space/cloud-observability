package observability

import (
	"context"

	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/trace"
)

// Tracer provides a simplified interface for tracing
type Tracer struct {
	tracer trace.Tracer
	name   string
}

// NewTracer creates a new Tracer instance
func NewTracer(name string) *Tracer {
	return &Tracer{
		tracer: otel.Tracer(name),
		name:   name,
	}
}

// Start starts a new span
func (t *Tracer) Start(ctx context.Context, name string, opts ...trace.SpanStartOption) (context.Context, trace.Span) {
	return t.tracer.Start(ctx, name, opts...)
}

// GetTracer returns the underlying OpenTelemetry tracer
func (t *Tracer) GetTracer() trace.Tracer {
	return t.tracer
}

// GetName returns the name of the tracer
func (t *Tracer) GetName() string {
	return t.name
}

// GetTraceID extracts trace ID from context
func (t *Tracer) GetTraceID(ctx context.Context) string {
	spanCtx := trace.SpanContextFromContext(ctx)
	if spanCtx.IsValid() {
		return spanCtx.TraceID().String()
	}
	return ""
}

// GetSpanID extracts span ID from context
func (t *Tracer) GetSpanID(ctx context.Context) string {
	spanCtx := trace.SpanContextFromContext(ctx)
	if spanCtx.IsValid() {
		return spanCtx.SpanID().String()
	}
	return ""
}
