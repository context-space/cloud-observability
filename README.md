# Observability Package

This package provides a unified observability solution for Go applications, including logging, tracing, and metrics collection.

## Features

- Structured logging with Zap
- Distributed tracing with OpenTelemetry
- Metrics collection with OpenTelemetry
- Unified configuration and initialization
- Easy-to-use provider interface

## Installation

```bash
go get github.com/context-space/observability
```

## Usage

```go
package main

import (
    "context"
    "github.com/context-space/observability"
)

func main() {
    ctx := context.Background()
    
    // Create configurations
    logConfig := &observability.LogConfig{
        Level: "info",
        // ... other log config options
    }
    
    tracingConfig := &observability.TracingConfig{
        Enabled:       true,
        ServiceName:   "my-service",
        ServiceVersion: "1.0.0",
        Endpoint:      "localhost:4317",
        // ... other tracing config options
    }
    
    metricsConfig := &observability.MetricsConfig{
        Enabled:  true,
        Endpoint: "localhost:4317",
        // ... other metrics config options
    }
    
    // Initialize the observability provider
    provider, cleanup, err := observability.InitializeObservabilityProvider(
        ctx,
        logConfig,
        tracingConfig,
        metricsConfig,
    )
    if err != nil {
        panic(err)
    }
    defer cleanup()
    
    // Use the provider
    logger := provider.Logger
    tracer := provider.Tracer
    metrics := provider.Metrics
    
    // Example usage
    logger.Info(ctx, "Application started")
    
    ctx, span := tracer.Start(ctx, "operation-name")
    defer span.End()
    
    metrics.Record(ctx, "metric-name", 1.0)
}
```

## Configuration

The package supports configuration for all three observability components:

### Logging Configuration
- Level: Log level (debug, info, warn, error)
- Format: Log format (json, console)
- Output: Output destination (stdout, file)

### Tracing Configuration
- Enabled: Enable/disable tracing
- ServiceName: Name of the service
- ServiceVersion: Version of the service
- Endpoint: OTLP endpoint
- SamplingRate: Trace sampling rate (0.0 to 1.0)

### Metrics Configuration
- Enabled: Enable/disable metrics
- Endpoint: OTLP endpoint
- ExportInterval: Metrics export interval

## License

MIT License 