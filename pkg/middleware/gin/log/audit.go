package log

import (
	"bytes"
	"kafka-polygon/pkg/http/consts"
	"kafka-polygon/pkg/middleware"

	"github.com/gin-gonic/gin"
)

type bodyLogWriter struct {
	gin.ResponseWriter
	body *bytes.Buffer
}

func (w bodyLogWriter) Write(b []byte) (int, error) {
	w.body.Write(b)
	return w.ResponseWriter.Write(b)
}

// Audit returns middleware function that logs request/response common information.
func Audit(o *middleware.AuditOptions) gin.HandlerFunc {
	return func(c *gin.Context) {
		_ = middleware.RenderAuditLogger(c, func() (*middleware.AuditData, error) {
			blw := &bodyLogWriter{body: &bytes.Buffer{}, ResponseWriter: c.Writer}
			c.Writer = blw
			c.Next()

			al := &middleware.AuditData{
				ServiceName: o.ServiceName(),
				StatusCode:  c.Writer.Status(),
				Method:      c.Request.Method,
				Path:        c.FullPath(),
				Headers: &middleware.AuditHeaders{
					XRequestID: c.GetString(consts.HeaderXRequestID),
					XUserID:    c.GetString(consts.HeaderXUserID),
					XClientID:  c.GetString(consts.HeaderXClientID),
				},
			}

			if c.Request != nil {
				al.UserAgent = c.Request.UserAgent()
			}

			if o.IsWithResponse() {
				al.ResponseBody = blw.body
			}

			return al, nil
		})
	}
}
