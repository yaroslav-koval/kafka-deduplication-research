package gin

import (
	"context"
	"kafka-polygon/pkg/cerror"
	pkgHTTPGin "kafka-polygon/pkg/http/gin"
	"kafka-polygon/pkg/http/gin/util"
	"kafka-polygon/pkg/workflow/entity"
	"kafka-polygon/pkg/workflow/entrypoint/controller/http"
	"kafka-polygon/pkg/workflow/entrypoint/controller/http/adapter"

	"github.com/gin-gonic/gin"
)

type Adapter struct {
	srv *pkgHTTPGin.Server
	uc  adapter.UseCase
}

var _ http.ServerAdapter = (*Adapter)(nil)

func NewAdapter(srv *pkgHTTPGin.Server, uc adapter.UseCase) *Adapter {
	return &Adapter{srv: srv, uc: uc}
}

func (a *Adapter) RegisterWorkflowRoutes(ctx context.Context, opts http.Option) error {
	prefixRouter := a.srv.Gin().RouterGroup.Group(opts.GetPrefix())

	prefixRouter.Handle(http.MethodSearchWorkflows, http.RouteSearchWorkflows, func(c *gin.Context) {
		if err := a.SearchWorkflows(c); err != nil {
			cerror.LogHTTPHandlerErrorCtx(c, err)
			util.AbortWithError(c, err)
		}
	})

	prefixRouter.Handle(http.MethodRestartWorkflow, http.RouteRestartWorkflow, func(c *gin.Context) {
		if err := a.RestartWorkflow(c); err != nil {
			cerror.LogHTTPHandlerErrorCtx(c, err)
			util.AbortWithError(c, err)
		}
	})

	prefixRouter.Handle(http.MethodRestartWorkflowFrom, http.RouteRestartWorkflowFrom, func(c *gin.Context) {
		if err := a.RestartWorkflowFrom(c); err != nil {
			cerror.LogHTTPHandlerErrorCtx(c, err)
			util.AbortWithError(c, err)
		}
	})

	return nil
}

func (a *Adapter) SearchWorkflows(ctx *gin.Context) error {
	var params entity.SearchWorkflowParams
	if err := ctx.ShouldBindQuery(&params); err != nil {
		return cerror.New(ctx, cerror.KindBadParams, err).LogError()
	}

	searchResult, err := a.uc.SearchWorkflows(ctx, params)
	if err != nil {
		return err
	}

	ctx.JSON(http.SuccessStatusSearchWorkflows, &http.SearchResponse{
		Data:   searchResult.Workflows,
		Paging: searchResult.Paging,
	})

	return ctx.Err()
}

func (a *Adapter) RestartWorkflow(ctx *gin.Context) error {
	var req http.RestartWorkflowRequest

	if err := ctx.ShouldBindJSON(&req); ctx.Request.ContentLength > 0 && err != nil {
		return cerror.New(ctx, cerror.KindBadParams, err).LogError()
	}

	err := a.uc.RestartWorkflow(ctx, entity.ID(ctx.Param(http.QueryParamID)), req.Payload)
	if err != nil {
		return err
	}

	ctx.JSON(http.SuccessStatusRestartWorkflow, "")

	return ctx.Err()
}

func (a *Adapter) RestartWorkflowFrom(ctx *gin.Context) error {
	var req http.RestartWorkflowFromRequest

	if err := ctx.ShouldBindJSON(&req); err != nil {
		return cerror.New(ctx, cerror.KindBadParams, err).LogError()
	}

	err := a.uc.RestartWorkflowFrom(ctx, entity.ID(ctx.Param(http.QueryParamID)), req.StepName, req.Payload)
	if err != nil {
		return err
	}

	ctx.JSON(http.SuccessStatusRestartWorkflow, "")

	return ctx.Err()
}
