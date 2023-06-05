package middleware

import (
	"bytes"
	"context"
	"kafka-polygon/pkg/log"
	"kafka-polygon/pkg/log/logger"
	"time"
)

// AuditOptions is struct that serves as options for audit logs.
type AuditOptions struct {
	serviceName    string
	isWithResponse bool
}

// NewAuditOptions is AuditOptions constructor.
func NewAuditOptions() *AuditOptions {
	return new(AuditOptions)
}

// WithServiceName sets up specific service name in audit logs.
func (o *AuditOptions) WithServiceName(sn string) *AuditOptions {
	o.serviceName = sn
	return o
}

// WithResponse turns on response logging in audit options.
func (o *AuditOptions) WithResponse() *AuditOptions {
	o.isWithResponse = true
	return o
}

// ServiceName returns current service name in audit options.
func (o *AuditOptions) ServiceName() string {
	return o.serviceName
}

// IsWithResponse returns whether response logging is on.
func (o *AuditOptions) IsWithResponse() bool {
	return o.isWithResponse
}

type AuditHeaders struct {
	XRequestID, XUserID, XClientID interface{}
}

type AuditData struct {
	ServiceName  string
	StatusCode   int
	Method       string
	Path         string
	UserAgent    string
	Headers      *AuditHeaders
	ResponseBody *bytes.Buffer
}

type AuditFuncCall func() (*AuditData, error)

func RenderAuditLogger(c context.Context, f AuditFuncCall) error {
	start := time.Now()
	data, err := f()
	latency := time.Since(start).Milliseconds()

	if latency < 1 {
		latency = 1
	}

	lvl := logger.LevelInfo
	//nolint:gomnd
	if data.StatusCode >= 400 {
		if data.StatusCode == 404 {
			lvl = logger.LevelWarn
		} else {
			lvl = logger.LevelError
		}
	}

	e := logger.NewEvent(c, lvl, "").
		WithValue(logger.FieldStatus, data.StatusCode).
		WithValue(logger.FieldMethod, data.Method).
		WithValue(logger.FieldLatency, latency).
		WithValue(logger.FieldURL, data.Path).
		WithValue(logger.FieldLogType, "audit").
		WithValue(logger.FieldServiceName, data.ServiceName).
		WithValue(logger.FieldUserAgent, data.UserAgent)

	if data.Headers != nil {
		e.WithValue(logger.FieldRequestID, data.Headers.XRequestID).
			WithValue(logger.FieldClientID, data.Headers.XClientID).
			WithValue(logger.FieldUserID, data.Headers.XUserID)
	}

	if data.ResponseBody != nil {
		e.WithValue(logger.FieldResponse, data.ResponseBody.String())
	}

	log.Log(e)

	return err
}
