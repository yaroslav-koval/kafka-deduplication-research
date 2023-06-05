package http

import (
	"context"
	"encoding/json"
	"kafka-polygon/pkg/workflow/entity"
	"net/http"
)

const (
	RouteSearchWorkflows     = "/"
	RouteRestartWorkflow     = "/restart/:id"
	RouteRestartWorkflowFrom = "/restart/from/:id"

	MethodSearchWorkflows     = http.MethodGet
	MethodRestartWorkflow     = http.MethodPost
	MethodRestartWorkflowFrom = http.MethodPost

	QueryParamID = "id"

	SuccessStatusSearchWorkflows     = http.StatusOK
	SuccessStatusRestartWorkflow     = http.StatusOK
	SuccessStatusRestartWorkflowFrom = http.StatusOK
)

type RestartWorkflowRequest struct {
	Payload json.RawMessage `json:"payload,omitempty"`
}

type RestartWorkflowFromRequest struct {
	StepName entity.WorkflowSchemaStepName `json:"step_name"`
	Payload  json.RawMessage               `json:"payload"`
}

type SearchResponse struct {
	Data   []*entity.Workflow `json:"data"`
	Paging entity.Paging      `json:"paging"`
}

type ServerAdapter interface {
	RegisterWorkflowRoutes(ctx context.Context, opts Option) error
}

// RegisterWorkflowRoutes registers workflow routes using a given server adapter.
// Use one from /adapter folder or provide your own implementation.
func RegisterWorkflowRoutes(ctx context.Context, serverAdapter ServerAdapter, opts ...OptionApply) error {
	return serverAdapter.RegisterWorkflowRoutes(ctx, GetOptions(opts...))
}
