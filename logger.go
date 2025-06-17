package observability

import (
	"context"
	"io"
	"os"

	"go.opentelemetry.io/otel/trace"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

// Logger is a wrapper around zap.Logger with context-aware methods
type Logger struct {
	logger *zap.Logger
}

// NewLogger creates a new logger from configuration
func NewLogger(config *LogConfig) (*Logger, error) {
	var logLevel zapcore.Level
	switch config.Level {
	case DebugLevel:
		logLevel = zapcore.DebugLevel
	case InfoLevel:
		logLevel = zapcore.InfoLevel
	case WarnLevel:
		logLevel = zapcore.WarnLevel
	case ErrorLevel:
		logLevel = zapcore.ErrorLevel
	case FatalLevel:
		logLevel = zapcore.FatalLevel
	default:
		logLevel = zapcore.InfoLevel
	}

	var outputs []io.Writer
	for _, path := range config.OutputPaths {
		if path == "stdout" {
			outputs = append(outputs, os.Stdout)
		} else if path == "stderr" {
			outputs = append(outputs, os.Stderr)
		} else {
			// Open file for writing
			file, err := os.OpenFile(path, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
			if err != nil {
				return nil, err
			}
			outputs = append(outputs, file)
		}
	}

	// Use default output if none specified
	if len(outputs) == 0 {
		outputs = append(outputs, os.Stdout)
	}

	encoderConfig := zapcore.EncoderConfig{
		TimeKey:        "timestamp",
		LevelKey:       "level",
		NameKey:        "logger",
		CallerKey:      "caller",
		FunctionKey:    zapcore.OmitKey,
		MessageKey:     "message",
		StacktraceKey:  "stacktrace",
		LineEnding:     zapcore.DefaultLineEnding,
		EncodeLevel:    zapcore.LowercaseLevelEncoder,
		EncodeTime:     zapcore.ISO8601TimeEncoder,
		EncodeDuration: zapcore.MillisDurationEncoder,
		EncodeCaller:   zapcore.ShortCallerEncoder,
	}

	var encoder zapcore.Encoder
	if config.Format == JSONFormat {
		encoder = zapcore.NewJSONEncoder(encoderConfig)
	} else {
		encoder = zapcore.NewConsoleEncoder(encoderConfig)
	}

	var syncer zapcore.WriteSyncer
	if len(outputs) == 1 {
		syncer = zapcore.AddSync(outputs[0])
	} else {
		syncers := make([]zapcore.WriteSyncer, len(outputs))
		for i, output := range outputs {
			syncers[i] = zapcore.AddSync(output)
		}
		syncer = zapcore.NewMultiWriteSyncer(syncers...)
	}

	core := zapcore.NewCore(encoder, syncer, logLevel)

	// Create logger with caller and stacktrace
	var logger *zap.Logger
	if config.Development {
		logger = zap.New(core, zap.AddCaller(), zap.AddStacktrace(zapcore.ErrorLevel), zap.Development())
	} else {
		logger = zap.New(core, zap.AddCaller(), zap.AddStacktrace(zapcore.ErrorLevel))
	}

	return &Logger{logger: logger}, nil
}

// With adds structured context to the Logger
func (l *Logger) With(fields ...zap.Field) *Logger {
	// Need to preserve the same caller skip behavior in the new logger instance
	return &Logger{logger: l.logger.With(fields...)}
}

// WithFields adds fields to the logger
func (l *Logger) WithFields(fields map[string]interface{}) *Logger {
	zapFields := make([]zap.Field, 0, len(fields))
	for k, v := range fields {
		zapFields = append(zapFields, zap.Any(k, v))
	}
	return &Logger{logger: l.logger.With(zapFields...)}
}

// getSkippedLogger returns a logger with the caller skip set to skip this file's methods
func (l *Logger) getSkippedLogger() *zap.Logger {
	// This ensures both caller information and stacktraces skip the wrapper logger methods
	return l.logger.WithOptions(zap.AddCallerSkip(1))
}

// Debug logs a debug message with trace context
func (l *Logger) Debug(ctx context.Context, msg string, fields ...zap.Field) {
	fields = append(fields, extractTraceFields(ctx)...)
	l.getSkippedLogger().Debug(msg, fields...)
}

// Info logs an info message with trace context
func (l *Logger) Info(ctx context.Context, msg string, fields ...zap.Field) {
	fields = append(fields, extractTraceFields(ctx)...)
	l.getSkippedLogger().Info(msg, fields...)
}

// Warn logs a warning message with trace context
func (l *Logger) Warn(ctx context.Context, msg string, fields ...zap.Field) {
	fields = append(fields, extractTraceFields(ctx)...)
	l.getSkippedLogger().Warn(msg, fields...)
}

// Error logs an error message with trace context
func (l *Logger) Error(ctx context.Context, msg string, fields ...zap.Field) {
	fields = append(fields, extractTraceFields(ctx)...)
	l.getSkippedLogger().Error(msg, fields...)
}

// Fatal logs a fatal message with trace context and exits
func (l *Logger) Fatal(ctx context.Context, msg string, fields ...zap.Field) {
	fields = append(fields, extractTraceFields(ctx)...)
	l.getSkippedLogger().Fatal(msg, fields...)
}

// extractTraceFields extracts trace information from context
func extractTraceFields(ctx context.Context) []zap.Field {
	fields := []zap.Field{}

	spanCtx := trace.SpanContextFromContext(ctx)
	if spanCtx.IsValid() {
		fields = append(fields, zap.String("trace_id", spanCtx.TraceID().String()))
		fields = append(fields, zap.String("span_id", spanCtx.SpanID().String()))
	}

	return fields
}

// Sync flushes any buffered log entries
func (l *Logger) Sync() error {
	return l.logger.Sync()
}
