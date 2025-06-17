package observability

import (
	"context"
	"fmt"
	"time"

	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/exporters/otlp/otlptrace"
	"go.opentelemetry.io/otel/exporters/otlp/otlptrace/otlptracegrpc"
	"go.opentelemetry.io/otel/propagation"
	"go.opentelemetry.io/otel/sdk/resource"
	sdktrace "go.opentelemetry.io/otel/sdk/trace"
	semconv "go.opentelemetry.io/otel/semconv/v1.4.0"
	"go.opentelemetry.io/otel/trace"
	"go.uber.org/zap"
)

// InitializeObservabilityProvider initializes all observability components properly
func InitializeObservabilityProvider(ctx context.Context, logConfig *LogConfig, tracingConfig *TracingConfig, metricsConfig *MetricsConfig) (*ObservabilityProvider, func(), error) {
	// Initialize logger
	logger, err := NewLogger(logConfig)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to initialize logger: %w", err)
	}

	// Initialize tracer
	tracer, tracerShutdown, err := setupTracing(ctx, tracingConfig)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to initialize tracer: %w", err)
	}

	// Initialize metrics
	metrics, err := NewMetrics(ctx, *metricsConfig)
	if err != nil {
		tracerShutdown(ctx)
		return nil, nil, fmt.Errorf("failed to initialize metrics: %w", err)
	}

	// Create cleanup function
	cleanup := func() {
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()

		if err := metrics.Shutdown(ctx); err != nil {
			logger.Error(ctx, "Error shutting down metrics", zap.Error(err))
		}

		if err := tracerShutdown(ctx); err != nil {
			logger.Error(ctx, "Error shutting down tracer", zap.Error(err))
		}

		if err := logger.Sync(); err != nil {
			fmt.Printf("Error syncing logger: %v\n", err)
		}
	}

	// Create and return the provider
	return &ObservabilityProvider{
		Logger:         logger,
		Tracer:         tracer,
		Metrics:        metrics,
		serviceName:    tracingConfig.ServiceName,
		serviceVersion: tracingConfig.ServiceVersion,
	}, cleanup, nil
}

// setupTracing initializes the OpenTelemetry tracer provider
func setupTracing(ctx context.Context, config *TracingConfig) (*Tracer, func(context.Context) error, error) {
	if !config.Enabled {
		// Return a no-op tracer when disabled
		tracer := NewTracer(config.ServiceName)
		return tracer, func(context.Context) error { return nil }, nil
	}

	// Create resource
	res, err := resource.New(ctx,
		resource.WithAttributes(
			semconv.ServiceNameKey.String(config.ServiceName),
			semconv.ServiceVersionKey.String(config.ServiceVersion),
			semconv.DeploymentEnvironmentKey.String(config.Environment),
		),
	)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to create resource: %w", err)
	}

	// Create OTLP exporter
	client := otlptracegrpc.NewClient(
		otlptracegrpc.WithEndpoint(config.Endpoint),
		otlptracegrpc.WithInsecure(),
	)

	exporter, err := otlptrace.New(ctx, client)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to create OTLP exporter: %w", err)
	}

	// Create a sampler
	var sampler sdktrace.Sampler
	if config.SamplingRate >= 1.0 {
		sampler = sdktrace.AlwaysSample()
	} else if config.SamplingRate <= 0.0 {
		sampler = sdktrace.NeverSample()
	} else {
		sampler = sdktrace.TraceIDRatioBased(config.SamplingRate)
	}

	// Create and register the trace provider
	tp := sdktrace.NewTracerProvider(
		sdktrace.WithSampler(sampler),
		sdktrace.WithBatcher(exporter),
		sdktrace.WithResource(res),
	)

	otel.SetTracerProvider(tp)

	// Set global propagator
	otel.SetTextMapPropagator(propagation.NewCompositeTextMapPropagator(
		propagation.TraceContext{},
		propagation.Baggage{},
	))

	// Create our custom tracer
	tracer := NewTracer(config.ServiceName)

	// Return tracer and shutdown function
	return tracer, tp.Shutdown, nil
}

// GetTraceID extracts trace ID from context
func GetTraceID(ctx context.Context) string {
	spanCtx := trace.SpanContextFromContext(ctx)
	if spanCtx.IsValid() {
		return spanCtx.TraceID().String()
	}
	return ""
}

// GetSpanID extracts span ID from context
func GetSpanID(ctx context.Context) string {
	spanCtx := trace.SpanContextFromContext(ctx)
	if spanCtx.IsValid() {
		return spanCtx.SpanID().String()
	}
	return ""
}
