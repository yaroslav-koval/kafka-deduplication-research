package logger

type Level int8

const (
	// Log levels
	LevelTrace Level = iota + 1
	LevelDebug
	LevelInfo
	LevelWarn
	LevelError
	LevelFatal

	lvlTrace = "trace"
	lvlDebug = "debug"
	lvlInfo  = "info"
	lvlWarn  = "warn"
	lvlError = "error"
	lvlFatal = "fatal"
)

func (l Level) String() string {
	switch l {
	case LevelTrace:
		return lvlTrace
	case LevelDebug:
		return lvlDebug
	case LevelInfo:
		return lvlInfo
	case LevelWarn:
		return lvlWarn
	case LevelError:
		return lvlError
	case LevelFatal:
		return lvlFatal
	}

	return "unknown"
}
