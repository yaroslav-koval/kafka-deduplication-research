package adapter

import (
	"context"
	"encoding/json"
	"kafka-polygon/pkg/workflow/entity"
)

// UseCase is a common business logic abstraction that is required for all adapters
type UseCase interface {
	SearchWorkflows(
		ctx context.Context, params entity.SearchWorkflowParams) (*entity.SearchWorkflowResult, error)
	RestartWorkflow(
		ctx context.Context, workflowID entity.ID, payload json.RawMessage) error
	RestartWorkflowFrom(
		ctx context.Context, workflowID entity.ID, from entity.WorkflowSchemaStepName, payload json.RawMessage) error
}
