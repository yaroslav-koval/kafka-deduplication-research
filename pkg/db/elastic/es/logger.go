package es

import (
	"context"
	"io"
	"kafka-polygon/pkg/log"
	"kafka-polygon/pkg/log/logger"
	"net/http"
	"time"
)

// CJSONLogger implements estansport Logger interface
type CJSONLogger struct {
	ctx context.Context
}

func (l *CJSONLogger) WithContext(ctx context.Context) {
	l.ctx = ctx
}

// LogRoundTrip prints some information about the request and response
func (l *CJSONLogger) LogRoundTrip(
	req *http.Request,
	res *http.Response,
	err error,
	start time.Time,
	dur time.Duration,
) error {
	var (
		nReq int64
		nRes int64
	)

	if req != nil && req.Body != nil && req.Body != http.NoBody {
		nReq, _ = io.Copy(io.Discard, req.Body)
	}

	if res != nil && res.Body != nil && res.Body != http.NoBody {
		nRes, _ = io.Copy(io.Discard, res.Body)
	}

	event := logger.NewEvent(l.ctx, l.getLevel(res, err), req.URL.String())

	values := make(map[string]interface{})
	values["method"] = req.Method
	values["status_code"] = res.StatusCode
	values["duration"] = dur
	values["req_bytes"] = nReq
	values["res_bytes"] = nRes

	for key, val := range values {
		event.WithValue(key, val)
	}

	log.Log(event)

	return nil
}

// RequestBodyEnabled output request body
func (l *CJSONLogger) RequestBodyEnabled() bool { return true }

// ResponseBodyEnabled output response body
func (l *CJSONLogger) ResponseBodyEnabled() bool { return true }

func (l *CJSONLogger) getLevel(res *http.Response, err error) logger.Level {
	var level logger.Level

	switch {
	case err != nil:
		level = logger.LevelError
	case res != nil && res.StatusCode > 0 && res.StatusCode < 300:
		level = logger.LevelInfo
	case res != nil && res.StatusCode > 299 && res.StatusCode < 500:
		level = logger.LevelWarn
	case res != nil && res.StatusCode > 499:
		level = logger.LevelError
	default:
		level = logger.LevelError
	}

	return level
}
