package log

import (
	"context"
	"io"
	"kafka-polygon/pkg/log/factory"
	"kafka-polygon/pkg/log/logger"
)

var l factory.Logger = factory.GetLogger()

func Log(e logger.LogEvent) {
	l.Log(e)
}

func Info(ctx context.Context, msg string) {
	l.Log(logger.NewEvent(ctx, logger.LevelInfo, msg))
}

func InfoF(ctx context.Context, msg string, args ...interface{}) {
	l.Log(logger.NewEventF(ctx, logger.LevelInfo, msg, args...))
}

func Debug(ctx context.Context, msg string) {
	l.Log(logger.NewEvent(ctx, logger.LevelDebug, msg))
}

func DebugF(ctx context.Context, msg string, args ...interface{}) {
	l.Log(logger.NewEventF(ctx, logger.LevelDebug, msg, args...))
}

// SetGlobalLogLevel sets global log level from parameter. If can not parse level set default info level.
func SetGlobalLogLevel(level string) {
	l.SetGlobalLogLevel(level)
}

// SetGlobalOutput sets global output for logs.
func SetGlobalOutput(w io.Writer) {
	l.SetGlobalOutput(w)
}
