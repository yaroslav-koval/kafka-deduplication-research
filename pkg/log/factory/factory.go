package factory

import (
	"io"
	"kafka-polygon/pkg/log/logger"
	"kafka-polygon/pkg/log/zero"
)

// Logger is a common interface for logging.
type Logger interface {
	// Log should log given event.
	Log(e logger.LogEvent)
	// SetGlobalLogLevel should set global log severity level.
	// All logs with lower severity level shouldn't be logged.
	SetGlobalLogLevel(level string)
	// SetGlobalOutput should set output writer for logs.
	SetGlobalOutput(w io.Writer)
}

var l Logger = zero.New()

// GetLogger returns Logger interface implementation.
func GetLogger() Logger {
	return l
}
