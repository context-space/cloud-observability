package observability

// LogLevel defines the logging level
type LogLevel int

const (
	DebugLevel LogLevel = iota
	InfoLevel
	WarnLevel
	ErrorLevel
	FatalLevel
)

// LogFormat defines the log output format
type LogFormat int

const (
	JSONFormat LogFormat = iota
	ConsoleFormat
)

// TracingConfig holds configuration for the tracer
type TracingConfig struct {
	ServiceName    string
	ServiceVersion string
	Environment    string
	Endpoint       string
	Enabled        bool
	SamplingRate   float64
}

// LogConfig holds configuration for the logger
type LogConfig struct {
	Level       LogLevel
	Format      LogFormat
	OutputPaths []string
	Development bool
}

// MetricsConfig holds configuration for metrics
type MetricsConfig struct {
	ServiceName    string
	ServiceVersion string
	Environment    string
	Enabled        bool
	Endpoint       string
}

// ObservabilityConfig holds all observability configuration
type ObservabilityConfig struct {
	Logging LogConfig
	Tracing TracingConfig
	Metrics MetricsConfig
	Service ServiceConfig
}

// ServiceConfig holds service information
type ServiceConfig struct {
	Name        string
	Version     string
	Environment string
}

// ParseLogLevel converts a string log level to a LogLevel enum
func ParseLogLevel(level string) LogLevel {
	switch level {
	case "debug":
		return DebugLevel
	case "info":
		return InfoLevel
	case "warn", "warning":
		return WarnLevel
	case "error":
		return ErrorLevel
	case "fatal":
		return FatalLevel
	default:
		return InfoLevel // Default to InfoLevel for unknown values
	}
}

// ParseLogFormat converts a string log format to a LogFormat enum
func ParseLogFormat(format string) LogFormat {
	switch format {
	case "json":
		return JSONFormat
	case "console":
		return ConsoleFormat
	default:
		return JSONFormat // Default to JSONFormat for unknown values
	}
}
