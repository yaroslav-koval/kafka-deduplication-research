package log

import (
	"bytes"
	"kafka-polygon/pkg/http/consts"
	"kafka-polygon/pkg/middleware"

	"github.com/gofiber/fiber/v2"
)

// Audit returns middleware function that logs request/response common information.
func Audit(o *middleware.AuditOptions) fiber.Handler {
	return func(fiberCtx *fiber.Ctx) error {
		ctx := fiberCtx.Context()

		return middleware.RenderAuditLogger(ctx, func() (*middleware.AuditData, error) {
			err := fiberCtx.Next()

			al := &middleware.AuditData{
				ServiceName: o.ServiceName(),
				StatusCode:  ctx.Response.StatusCode(),
				Method:      string(ctx.Method()),
				Path:        string(ctx.Path()),
				Headers: &middleware.AuditHeaders{
					XRequestID: ctx.UserValue(consts.HeaderXRequestID),
					XUserID:    ctx.UserValue(consts.HeaderXUserID),
					XClientID:  ctx.UserValue(consts.HeaderXClientID),
				},
				UserAgent: string(ctx.UserAgent()),
			}

			if o.IsWithResponse() {
				al.ResponseBody = bytes.NewBuffer(fiberCtx.Response().Body())
			}

			return al, err
		})
	}
}
