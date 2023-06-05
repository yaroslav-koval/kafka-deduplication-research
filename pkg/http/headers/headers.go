package headers

import (
	"context"
	"kafka-polygon/pkg/http/consts"
	"kafka-polygon/pkg/log"
	"kafka-polygon/pkg/log/logger"
	"net/http"

	"github.com/valyala/fasthttp"
)

// AddHeadersFromContext reads predefined list of headers from context
// and adds them to request headers.
// Req parameter should be one of types *http.Request or *fasthttp.Request
func AddHeadersFromContext(ctx context.Context, req interface{}) {
	for _, header := range consts.RequestHeadersToSave() {
		val, ok := ctx.Value(header).(string)
		if !ok || val == "" {
			val = getDefHeaderValue(header)
		}

		addHeader(req, header, val)
	}
}

// ContextHeadersSync copies saved http headers
// from parent context to the given new context
func ContextHeadersSync(parent, nCtx context.Context) context.Context {
	for _, header := range consts.RequestHeadersToSave() {
		val, ok := parent.Value(header).(string)
		if ok {
			//nolint:staticcheck
			nCtx = context.WithValue(nCtx, header, val)
		}
	}

	return nCtx
}

func getDefHeaderValue(header string) string {
	if header == consts.HeaderXClientID {
		return consts.BackendAppName
	}

	return ""
}

func addHeader(req interface{}, key, val string) {
	switch dr := req.(type) {
	case *http.Request:
		dr.Header.Set(key, val)
	case *fasthttp.Request:
		dr.Header.Add(key, val)
	default:
		log.Log(logger.NewEventF(
			context.Background(), logger.LevelError, "request type %T is not supported", req))
	}
}
