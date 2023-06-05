package headers

import (
	"fmt"
	"kafka-polygon/pkg/cerror"
	"kafka-polygon/pkg/http/consts"
	"strings"

	"github.com/gofiber/fiber/v2"
	uuid "github.com/satori/go.uuid"
)

// RequestHeadersToContext returns middleware that saves given headers to context.
func RequestHeadersToContext(headers []string) fiber.Handler {
	return func(c *fiber.Ctx) error {
		for _, h := range headers {
			value := c.Get(h)
			if h == consts.HeaderXRequestID && value == "" {
				value = uuid.NewV4().String()
			}

			c.Context().SetUserValue(h, value)
		}

		return c.Next()
	}
}

// ResponseHeadersFromContext returns middleware that reads header values
// from context and sets them to response headers.
func ResponseHeadersFromContext(headers []string) fiber.Handler {
	return func(c *fiber.Ctx) error {
		err := c.Next()

		for _, h := range headers {
			value := fmt.Sprintf("%v", c.Context().UserValue(h))
			if value != "" {
				c.Set(h, value)
			}
		}

		return err
	}
}

// ValidateJSONContentType returns middleware that validates request JSON content type
// for methods that not in the excludeMethods list
func ValidateJSONContentType(excludeMethods ...string) fiber.Handler {
	return func(c *fiber.Ctx) error {
		m := c.Method()

		if !contains(excludeMethods, m) {
			currentContentType := strings.ToLower(c.Get(fiber.HeaderContentType))
			if !strings.HasPrefix(currentContentType, fiber.MIMEApplicationJSON) {
				err := cerror.NewF(
					c.Context(),
					cerror.KindBadContentType,
					"content-type header should be %v. Current value: %v",
					fiber.MIMEApplicationJSON, c.Get(fiber.HeaderContentType)).LogError()

				return c.Status(err.Kind().HTTPCode()).JSON(cerror.BuildErrorResponse(err))
			}
		}

		return c.Next()
	}
}

func contains(s []string, e string) bool {
	for _, a := range s {
		if a == e {
			return true
		}
	}

	return false
}
