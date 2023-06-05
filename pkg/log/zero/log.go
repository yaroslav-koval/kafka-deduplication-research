package zero

import (
	"context"
	"fmt"
	"io"
	"kafka-polygon/pkg/http/consts"
	"kafka-polygon/pkg/log/logger"
	"os"

	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
)

// Logger is zerolog's implementation of logging.
type Logger struct {
	l zerolog.Logger
}

// New creates instance of zerolog's implementation of logging.
func New() *Logger {
	return &Logger{l: log.Logger.Output(os.Stdout)}
}

// Log logs given log logger.
func (z *Logger) Log(e logger.LogEvent) {
	ze := zeroEventFromLevel(&z.l, e.Level())
	addStr(ze, logger.FieldMessage, e.Msg())
	ze.AnErr(logger.FieldError, e.Err())

	for _, v := range e.Fields() {
		ze.Interface(v.Key(), v.Value())
	}

	cv := newCtxValues(e.Ctx())
	if _, ok := e.Fields().GetValue(logger.FieldRequestID); !ok {
		addStr(ze, logger.FieldRequestID, cv.reqID)
	}

	ze.Send()
}

// SetGlobalLogLevel sets global log severity level.
// All logs with lower severity level will not be logged.
func (z *Logger) SetGlobalLogLevel(level string) {
	l, err := zerolog.ParseLevel(level)
	if err != nil {
		z.l.Error().Err(err).Msg("Default level will be changed to info level")

		l = zerolog.InfoLevel
	}

	zerolog.SetGlobalLevel(l)
}

// SetGlobalOutput sets output writer for logs.
func (z *Logger) SetGlobalOutput(w io.Writer) {
	z.l = z.l.Output(w)
}

type ctxValues struct {
	reqID string
}

func newCtxValues(ctx context.Context) *ctxValues {
	values := new(ctxValues)
	if ctx == nil {
		return values
	}

	if reqID := ctx.Value(consts.HeaderXRequestID); reqID != nil {
		switch v := reqID.(type) {
		case fmt.Stringer:
			values.reqID = v.String()
		default:
			values.reqID = fmt.Sprintf("%+v", v)
		}
	}

	return values
}

func addStr(e *zerolog.Event, key, value string) {
	if value != "" {
		e.Str(key, value)
	}
}

func zeroEventFromLevel(l *zerolog.Logger, level logger.Level) *zerolog.Event {
	switch level {
	case logger.LevelTrace:
		return l.Trace()
	case logger.LevelDebug:
		return l.Debug()
	case logger.LevelInfo:
		return l.Info()
	case logger.LevelWarn:
		return l.Warn()
	case logger.LevelError:
		return l.Error()
	case logger.LevelFatal:
		return l.WithLevel(zerolog.FatalLevel)
	}

	return l.Info()
}
