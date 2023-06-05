package logger_test

import (
	"kafka-polygon/pkg/log/logger"
	"testing"

	"github.com/tj/assert"
)

func TestLevel(t *testing.T) {
	for k, v := range map[logger.Level]int8{
		logger.LevelTrace: 1,
		logger.LevelDebug: 2,
		logger.LevelInfo:  3,
		logger.LevelWarn:  4,
		logger.LevelError: 5,
		logger.LevelFatal: 6,
	} {
		assert.Equal(t, v, int8(k))
	}
}

func TestLevelString(t *testing.T) {
	for k, v := range map[logger.Level]string{
		logger.LevelTrace:       "trace",
		logger.LevelDebug:       "debug",
		logger.LevelInfo:        "info",
		logger.LevelWarn:        "warn",
		logger.LevelError:       "error",
		logger.LevelFatal:       "fatal",
		logger.Level(int8(-99)): "unknown",
	} {
		assert.Equal(t, v, k.String())
	}
}
