package fiber

import (
	"context"
	"fmt"
	"kafka-polygon/pkg/env"
	"kafka-polygon/pkg/http/consts"
	"kafka-polygon/pkg/log"
	"kafka-polygon/pkg/log/logger"
	"kafka-polygon/pkg/middleware"
	"kafka-polygon/pkg/middleware/fiber/headers"
	middlewareLog "kafka-polygon/pkg/middleware/fiber/log"
	middlewareOpenapi "kafka-polygon/pkg/middleware/fiber/openapi"
	middlewareTracing "kafka-polygon/pkg/middleware/fiber/tracing"
	"kafka-polygon/pkg/tracing"

	"github.com/getkin/kin-openapi/openapi3"
	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/recover"
	"github.com/gofiber/fiber/v2/utils"
)

// ServerConfig is configuration for http server
type ServerConfig struct {
	Service env.Service
	Server  env.HTTPServer
	// Swagger gives an opportunity to provide API spec if server is based on Open API
	Swagger *openapi3.T
	// JSONEncoder gives an opportunity to provide custom json encoder
	JSONEncoder utils.JSONMarshal
	// JSONDecoder gives an opportunity to provide custom json decoder
	JSONDecoder utils.JSONUnmarshal
}

type Server struct {
	cfg   *ServerConfig
	fiber *fiber.App
}

func NewServer(cfg *ServerConfig) *Server {
	fConf := fiber.Config{
		EnablePrintRoutes: false,
		AppName:           cfg.Service.Name,
		ReadTimeout:       cfg.Server.ReadTimeout(),
		WriteTimeout:      cfg.Server.WriteTimeout(),
		IdleTimeout:       cfg.Server.IdleTimeout(),
		JSONEncoder:       cfg.JSONEncoder,
		JSONDecoder:       cfg.JSONDecoder,
	}

	return &Server{
		cfg:   cfg,
		fiber: fiber.New(fConf),
	}
}

func (s *Server) Fiber() *fiber.App {
	return s.fiber
}

// WithDefaultKit applies predefined set of middleware
func (s *Server) WithDefaultKit() *Server {
	s.WithHealthCheckRoute()
	s.WithGetSpecRoute()

	s.SetCloseOnShutdown(s.cfg.Server.CloseOnShutdown)
	s.Fiber().Use(recover.New(recover.Config{EnableStackTrace: true}))

	auditOptions := middleware.NewAuditOptions().WithServiceName(s.cfg.Service.Name)

	if s.cfg.Server.LogResponse {
		auditOptions = auditOptions.WithResponse()
	}

	s.Fiber().Use(middlewareLog.Audit(auditOptions))
	s.Fiber().Use(headers.RequestHeadersToContext(consts.RequestHeadersToSave()))
	s.Fiber().Use(headers.ResponseHeadersFromContext(consts.ResponseHeadersToSend()))

	if s.cfg.Swagger != nil {
		s.cfg.Swagger.Servers = nil
		s.Fiber().Use(middlewareOpenapi.OapiRequestValidator(s.cfg.Swagger))
	}

	return s
}

// WithHealthCheckRoute adds /health route. Uses default handler if no custom handlers will be passed
func (s *Server) WithHealthCheckRoute(customHandlers ...fiber.Handler) *Server {
	h := []fiber.Handler{DefaultHealthCheckHandler}
	if len(customHandlers) > 0 {
		h = customHandlers
	}

	s.Fiber().Add(fiber.MethodGet, "/health", h...)

	return s
}

// WithGetSpecRoute adds /spec route that returns service specification
func (s *Server) WithGetSpecRoute(customHandlers ...fiber.Handler) *Server {
	h := []fiber.Handler{DefaultGetSpecHandler(s.cfg.Swagger)}
	if len(customHandlers) > 0 {
		h = customHandlers
	}

	s.Fiber().Add(fiber.MethodGet, "/spec", h...)

	return s
}

func (s *Server) Shutdown(ctx context.Context) error {
	if s.Fiber().Config().ReadTimeout == 0 {
		return nil
	}

	return s.Fiber().Shutdown()
}

// SetCloseOnShutdown when true adds a `Connection: close` header when the server is shutting down.
func (s *Server) SetCloseOnShutdown(enabled bool) {
	s.Fiber().Server().CloseOnShutdown = enabled
}

func (s *Server) ListenAndServe(ctx context.Context) chan error {
	errCh := make(chan error)

	go func() {
		listenAddress := fmt.Sprintf("%s:%s", s.cfg.Server.Host, s.cfg.Server.Port)
		log.Log(logger.NewEventF(ctx, logger.LevelInfo, "starting listen and serve on %s", listenAddress))
		errCh <- s.Fiber().Listen(listenAddress)
	}()

	return errCh
}

func (s *Server) SetTracing(provider tracing.Provider) {
	s.Fiber().Use(middlewareTracing.New(tracing.New(provider)))
}

// DefaultHealthCheckHandler is default handler for /health route
func DefaultHealthCheckHandler(ctx *fiber.Ctx) error {
	return ctx.Status(fiber.StatusOK).JSON("")
}

// DefaultGetSpecHandler returns default handler for /spec route based on swagger spec
func DefaultGetSpecHandler(s *openapi3.T) fiber.Handler {
	return func(ctx *fiber.Ctx) error {
		if s == nil {
			return ctx.SendStatus(fiber.StatusNotFound)
		}

		ctx.Set(fiber.HeaderContentType, fiber.MIMEApplicationJSON)

		b, _ := s.MarshalJSON()

		return ctx.Status(fiber.StatusOK).Send(b)
	}
}
