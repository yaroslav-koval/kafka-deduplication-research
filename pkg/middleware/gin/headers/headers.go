package headers

import (
	"kafka-polygon/pkg/cerror"
	"kafka-polygon/pkg/middleware"

	"github.com/gin-gonic/gin"
)

// RequestHeadersToContext returns middleware that saves given headers to context.
func RequestHeadersToContext(headers []string) gin.HandlerFunc {
	return func(c *gin.Context) {
		for _, h := range headers {
			c.Set(h, middleware.DefaultHeaderXRequestID(h, c.GetHeader(h)))
		}

		c.Next()
	}
}

// ResponseHeadersFromContext returns middleware that reads header values
// from context and sets them to response headers.
func ResponseHeadersFromContext(headers []string) gin.HandlerFunc {
	return func(c *gin.Context) {
		c.Next()

		for _, h := range headers {
			value := c.GetString(h)
			if value != "" {
				c.Header(h, value)
			}
		}
	}
}

// ValidateJSONContentType returns middleware that validates request JSON content type
// for methods that not in the excludeMethods list
func ValidateJSONContentType(excludeMethods ...string) gin.HandlerFunc {
	return func(c *gin.Context) {
		cErr := middleware.ValidateContentTypeJSON(c, c.Request.Method, c.ContentType(), excludeMethods...)
		if cErr != nil {
			c.AbortWithStatusJSON(cErr.Kind().HTTPCode(), cerror.BuildErrorResponse(cErr))

			return
		}

		c.Next()
	}
}
