package observability

// ObservabilityProvider provides unified access to all observability components (logging, tracing, metrics)
type ObservabilityProvider struct {
	Logger         *Logger
	Tracer         *Tracer
	Metrics        *Metrics
	serviceName    string
	serviceVersion string
}

// NewObservabilityProvider creates a new observability provider with all components
func NewObservabilityProvider(
	logger *Logger,
	tracer *Tracer,
	metrics *Metrics,
	serviceName, serviceVersion string,
) *ObservabilityProvider {
	return &ObservabilityProvider{
		Logger:         logger,
		Tracer:         tracer,
		Metrics:        metrics,
		serviceName:    serviceName,
		serviceVersion: serviceVersion,
	}
}
