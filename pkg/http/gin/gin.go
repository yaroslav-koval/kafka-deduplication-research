package gin

import (
	"context"
	"fmt"
	"kafka-polygon/pkg/cerror"
	"kafka-polygon/pkg/env"
	"kafka-polygon/pkg/http/consts"
	"kafka-polygon/pkg/http/gin/util"
	"kafka-polygon/pkg/log"
	"kafka-polygon/pkg/log/logger"
	"kafka-polygon/pkg/middleware"
	"kafka-polygon/pkg/middleware/gin/headers"
	middlewareLog "kafka-polygon/pkg/middleware/gin/log"
	middlewareOpenapi "kafka-polygon/pkg/middleware/gin/openapi"
	middlewareTracing "kafka-polygon/pkg/middleware/gin/tracing"
	"kafka-polygon/pkg/tracing"
	"net/http"

	"github.com/getkin/kin-openapi/openapi3"
	"github.com/gin-contrib/pprof"
	"github.com/gin-gonic/gin"
)

// ServerConfig is configuration for http server
type ServerConfig struct {
	Service env.Service
	Server  env.HTTPServer
	// Swagger gives an opportunity to provide API spec if server is based on Open API
	Swagger *openapi3.T
}

type Server struct {
	cfg    *ServerConfig
	srv    *http.Server
	engine *gin.Engine
}

func NewServer(cfg *ServerConfig) *Server {
	return &Server{
		cfg: cfg,
		srv: &http.Server{
			ReadTimeout:  cfg.Server.ReadTimeout(),
			WriteTimeout: cfg.Server.WriteTimeout(),
			IdleTimeout:  cfg.Server.IdleTimeout(),
		},
		engine: gin.New(),
	}
}

func InitMode(cfg *ServerConfig) {
	if !cfg.Service.DebugMode {
		gin.SetMode(gin.ReleaseMode)
	}
}

func (s *Server) Gin() *gin.Engine {
	return s.engine
}

// WithDefaultKit applies predefined set of middleware
func (s *Server) WithDefaultKit() *Server {
	s.Gin().Use(gin.Recovery())

	s.Gin().UnescapePathValues = true

	if s.cfg.Service.PProfMode {
		pprof.Register(s.Gin())
	}

	s.WithHealthCheckRoute()
	s.WithGetSpecRoute()

	auditOptions := middleware.NewAuditOptions().WithServiceName(s.cfg.Service.Name)

	if s.cfg.Server.LogResponse {
		auditOptions = auditOptions.WithResponse()
	}

	s.Gin().Use(middlewareLog.Audit(auditOptions))
	s.Gin().Use(headers.RequestHeadersToContext(consts.RequestHeadersToSave()))
	s.Gin().Use(headers.ResponseHeadersFromContext(consts.ResponseHeadersToSend()))

	if s.cfg.Swagger != nil {
		s.cfg.Swagger.Servers = nil
		s.Gin().Use(middlewareOpenapi.OapiRequestValidatorWithOptions(s.cfg.Swagger, &middlewareOpenapi.Options{
			SkipperGroup: []string{consts.WorkflowRoutesPrefix},
			ErrorHandler: func(c *gin.Context, statusCode int, err error) {
				c.AbortWithStatusJSON(statusCode, cerror.BuildErrorResponse(err))
			},
		}))
	}

	return s
}

// WithHealthCheckRoute adds /health route. Uses default handler if no custom handlers will be passed
func (s *Server) WithHealthCheckRoute(customHandlers ...gin.HandlerFunc) *Server {
	h := []gin.HandlerFunc{DefaultHealthCheckHandler}
	if len(customHandlers) > 0 {
		h = customHandlers
	}

	s.Gin().Handle(http.MethodGet, "/health", h...)

	return s
}

func (s *Server) SetTracing(provider tracing.Provider) {
	s.Gin().Use(middlewareTracing.New(tracing.New(provider)))
}

func HandlerFunc(fn func(c *gin.Context) error) gin.HandlerFunc {
	if fn == nil {
		return nil
	}

	return func(ctx *gin.Context) {
		if err := fn(ctx); err != nil {
			cerror.LogHTTPHandlerErrorCtx(ctx, err)
			util.AbortWithError(ctx, err)
		}
	}
}

// DefaultHealthCheckHandler is default handler for /health route
func DefaultHealthCheckHandler(ctx *gin.Context) {
	ctx.JSON(http.StatusOK, "")
}

// WithGetSpecRoute adds /spec route that returns service specification
func (s *Server) WithGetSpecRoute(customHandlers ...gin.HandlerFunc) *Server {
	h := []gin.HandlerFunc{DefaultGetSpecHandler(s.cfg.Swagger)}
	if len(customHandlers) > 0 {
		h = customHandlers
	}

	s.Gin().Handle(http.MethodGet, "/spec", h...)

	return s
}

// DefaultGetSpecHandler returns default handler for /spec route based on swagger spec
func DefaultGetSpecHandler(s *openapi3.T) gin.HandlerFunc {
	return func(ctx *gin.Context) {
		if s == nil {
			util.AbortErr(ctx, http.StatusNotFound)
			return
		}

		b, _ := s.MarshalJSON()

		ctx.Data(http.StatusOK, gin.MIMEJSON, b)
	}
}

func (s *Server) ListenAndServe(ctx context.Context) chan error {
	errCh := make(chan error)

	go func() {
		listenAddress := fmt.Sprintf("%s:%s", s.cfg.Server.Host, s.cfg.Server.Port)
		log.Log(logger.NewEventF(ctx, logger.LevelInfo, "starting listen and serve on %s", listenAddress))
		s.srv.Addr = listenAddress
		s.srv.Handler = s.Gin()
		errCh <- s.srv.ListenAndServe()
	}()

	return errCh
}

func (s *Server) Shutdown(ctx context.Context) error {
	if s.cfg.Server.CloseOnShutdown {
		return s.srv.Shutdown(ctx)
	}

	return s.srv.Close()
}
