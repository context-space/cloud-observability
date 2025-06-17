package observability

import (
	"context"
	"fmt"
	"time"

	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/exporters/otlp/otlpmetric/otlpmetricgrpc"
	"go.opentelemetry.io/otel/metric"
	sdkmetric "go.opentelemetry.io/otel/sdk/metric"
	"go.opentelemetry.io/otel/sdk/resource"
	semconv "go.opentelemetry.io/otel/semconv/v1.4.0"
)

// Metrics is a wrapper for OpenTelemetry metrics
type Metrics struct {
	meter      metric.Meter
	counters   map[string]metric.Int64Counter
	gauges     map[string]metric.Float64ObservableGauge
	histograms map[string]metric.Float64Histogram
	shutdown   func() error
}

// NewMetrics creates a new metrics collector
func NewMetrics(ctx context.Context, config MetricsConfig) (*Metrics, error) {
	if !config.Enabled {
		return &Metrics{
			counters:   make(map[string]metric.Int64Counter),
			gauges:     make(map[string]metric.Float64ObservableGauge),
			histograms: make(map[string]metric.Float64Histogram),
			shutdown:   func() error { return nil },
		}, nil
	}

	// Create resource with service information
	res, err := resource.New(ctx,
		resource.WithAttributes(
			semconv.ServiceNameKey.String(config.ServiceName),
			semconv.ServiceVersionKey.String(config.ServiceVersion),
			attribute.String("environment", config.Environment),
		),
	)
	if err != nil {
		return nil, fmt.Errorf("failed to create resource: %w", err)
	}

	// Create OTLP exporter
	exporter, err := otlpmetricgrpc.New(ctx,
		otlpmetricgrpc.WithEndpoint(config.Endpoint),
		otlpmetricgrpc.WithInsecure(),
	)
	if err != nil {
		return nil, fmt.Errorf("failed to create OTLP exporter: %w", err)
	}

	// Create meter provider
	meterProvider := sdkmetric.NewMeterProvider(
		sdkmetric.WithResource(res),
		sdkmetric.WithReader(sdkmetric.NewPeriodicReader(exporter)),
	)
	otel.SetMeterProvider(meterProvider)

	// Create meter
	meter := meterProvider.Meter(config.ServiceName)

	return &Metrics{
		meter:      meter,
		counters:   make(map[string]metric.Int64Counter),
		gauges:     make(map[string]metric.Float64ObservableGauge),
		histograms: make(map[string]metric.Float64Histogram),
		shutdown: func() error {
			return meterProvider.Shutdown(ctx)
		},
	}, nil
}

// Shutdown stops the metrics collection
func (m *Metrics) Shutdown(ctx context.Context) error {
	return m.shutdown()
}

// CreateCounter creates a new counter metric
func (m *Metrics) CreateCounter(name, description string) (metric.Int64Counter, error) {
	if counter, exists := m.counters[name]; exists {
		return counter, nil
	}

	counter, err := m.meter.Int64Counter(
		name,
		metric.WithDescription(description),
	)
	if err != nil {
		return nil, fmt.Errorf("failed to create counter: %w", err)
	}

	m.counters[name] = counter
	return counter, nil
}

// IncrementCounter increments a counter by the given value with optional attributes
func (m *Metrics) IncrementCounter(ctx context.Context, name string, value int64, attrs ...attribute.KeyValue) error {
	counter, exists := m.counters[name]
	if !exists {
		// If counter doesn't exist, create it
		var err error
		counter, err = m.CreateCounter(name, "Counter for "+name)
		if err != nil {
			// Log the error and return
			fmt.Printf("Failed to create counter: %v\n", err)
			return err
		}
	}

	counter.Add(ctx, value, metric.WithAttributes(attrs...))
	return nil
}

// CreateHistogram creates a new histogram metric
func (m *Metrics) CreateHistogram(name, description, unit string) (metric.Float64Histogram, error) {
	if histogram, exists := m.histograms[name]; exists {
		return histogram, nil
	}

	histogram, err := m.meter.Float64Histogram(
		name,
		metric.WithDescription(description),
		metric.WithUnit(unit),
	)
	if err != nil {
		return nil, fmt.Errorf("failed to create histogram: %w", err)
	}

	m.histograms[name] = histogram
	return histogram, nil
}

// RecordHistogram records a value to a histogram with optional attributes
func (m *Metrics) RecordHistogram(ctx context.Context, name string, value float64, attrs ...attribute.KeyValue) error {
	histogram, exists := m.histograms[name]
	if !exists {
		// If histogram doesn't exist, create it
		var err error
		histogram, err = m.CreateHistogram(name, "Duration of "+name, "s")
		if err != nil {
			// Log the error and return
			fmt.Printf("Failed to create histogram: %v\n", err)
			return err
		}
	}

	histogram.Record(ctx, value, metric.WithAttributes(attrs...))
	return nil
}

// CreateGauge creates a new gauge metric
func (m *Metrics) CreateGauge(name, description string, callback func() float64) (metric.Float64ObservableGauge, error) {
	if gauge, exists := m.gauges[name]; exists {
		return gauge, nil
	}

	gauge, err := m.meter.Float64ObservableGauge(
		name,
		metric.WithDescription(description),
	)
	if err != nil {
		return nil, fmt.Errorf("failed to create gauge: %w", err)
	}

	_, err = m.meter.RegisterCallback(
		func(_ context.Context, observer metric.Observer) error {
			observer.ObserveFloat64(gauge, callback())
			return nil
		},
		gauge,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to register callback: %w", err)
	}

	m.gauges[name] = gauge
	return gauge, nil
}

// MeasureDuration measures the duration of a function call and records it to a histogram
func (m *Metrics) MeasureDuration(ctx context.Context, name string, attrs ...attribute.KeyValue) func() {
	start := time.Now()
	return func() {
		duration := time.Since(start).Seconds()
		histogram, exists := m.histograms[name]
		if !exists {
			// If histogram doesn't exist, create it
			var err error
			histogram, err = m.CreateHistogram(name, "Duration of "+name, "s")
			if err != nil {
				// Log the error and return
				fmt.Printf("Failed to create histogram: %v\n", err)
				return
			}
		}
		histogram.Record(ctx, duration, metric.WithAttributes(attrs...))
	}
}
