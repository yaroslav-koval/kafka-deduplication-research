package usecase

import (
	"context"
	"encoding/json"
	"kafka-polygon/pkg/cerror"
	"kafka-polygon/pkg/http/consts"
	"kafka-polygon/pkg/workflow/entity"
	"time"
)

type Orchestrator interface {
	WorkflowSchema(ctx context.Context, wsn entity.WorkflowSchemaName) (*entity.WorkflowSchema, error)
	StartFrom(
		ctx context.Context,
		schemaName entity.WorkflowSchemaName,
		stepName entity.WorkflowSchemaStepName,
		parentID *entity.ID,
		payload json.RawMessage) (entity.ID, error)
	RestartFrom(
		ctx context.Context,
		workflowID entity.ID,
		schemaName entity.WorkflowSchemaName,
		stepName entity.WorkflowSchemaStepName,
		payload json.RawMessage) (entity.ID, error)
}

type Store interface {
	NewID() entity.ID
	GetWorkflowByID(ctx context.Context, id entity.ID) (*entity.Workflow, error)
	SearchWorkflows(ctx context.Context, params entity.SearchWorkflowParams) (*entity.SearchWorkflowResult, error)
	CreateWorkflowHistory(ctx context.Context, wh *entity.WorkflowHistory) error
}

type UseCase struct {
	orchestrator Orchestrator
	store        Store
}

func New(o Orchestrator, s Store) *UseCase {
	return &UseCase{orchestrator: o, store: s}
}

// SearchWorkflows - searches workflows by a given parameters
func (uc *UseCase) SearchWorkflows(
	ctx context.Context, params entity.SearchWorkflowParams) (*entity.SearchWorkflowResult, error) {
	return uc.store.SearchWorkflows(ctx, params)
}

// RestartWorkflow restarts failed workflow from the last step.
// If you need to restart the last step with different payload, you can pass payload json.RawMessage.
// If you need to restart the last step with the original payload, pass nil to the payload parameter.
// !!! note that this method restarts only FAILED workflows. Use RestartWorkflowFrom for successful workflows.
func (uc *UseCase) RestartWorkflow(
	ctx context.Context, workflowID entity.ID, payload json.RawMessage) error {
	workflowRecord, err := uc.store.GetWorkflowByID(ctx, workflowID)
	if err != nil {
		return err
	}

	if workflowRecord.Status != entity.WorkflowStatusFailed {
		return cerror.NewF(ctx, cerror.KindBadValidation, "workflow is not in %s status", entity.WorkflowStatusFailed).
			LogError()
	}

	if workflowRecord.Input == nil &&
		(len(workflowRecord.Steps) == 0 || workflowRecord.Steps[len(workflowRecord.Steps)-1].Data == nil) {
		return cerror.NewF(ctx, cerror.KindConflict, "both workflow input and last step data are empty").LogError()
	}

	workflowSchema, err := uc.orchestrator.WorkflowSchema(ctx, workflowRecord.SchemaName)
	if err != nil {
		return err
	}

	stepName := workflowSchema.FirstStep().Name()
	newPayload := workflowRecord.Input
	oldPayload := workflowRecord.Input

	if len(workflowRecord.Steps) > 0 {
		lastStep := workflowRecord.Steps[len(workflowRecord.Steps)-1]
		stepName = lastStep.Name
		oldPayload = lastStep.Data
		newPayload = lastStep.Data
	}

	if payload != nil {
		newPayload = payload
	}

	_, err = uc.orchestrator.RestartFrom(ctx, workflowID, workflowRecord.SchemaName, stepName, newPayload)
	if err != nil {
		return err
	}

	err = uc.store.CreateWorkflowHistory(ctx, &entity.WorkflowHistory{
		ID:                uc.store.NewID(),
		CreatedAt:         time.Now().UTC(),
		Type:              entity.WorkflowHistoryTypeRestart,
		Input:             newPayload,
		InputPrevious:     oldPayload,
		StepName:          stepName,
		WorkflowID:        workflowID,
		WorkflowStatus:    workflowRecord.Status,
		WorkflowError:     workflowRecord.Error,
		WorkflowErrorKind: workflowRecord.ErrorKind,
		RequestID:         requestIDFromCtx(ctx),
	})
	if err != nil {
		_ = cerror.NewF(
			ctx, cerror.KindInternal, "create history for restart workflow with id=%s", workflowID).LogError()
	}

	return nil
}

// RestartWorkflowFrom restarts workflow that was successfully finished earlier from the given.
// Payload is a required parameter. You need to pass it even if you want to restart step with unchanged payload.
// Method doesn't modify the original workflow. Instead it creates a new workflow
// with parentID link to the original workflow.
// !!! note that this method restarts only SUCCESS workflows. Use RestartWorkflow for failed workflows.
func (uc *UseCase) RestartWorkflowFrom(
	ctx context.Context, workflowID entity.ID, from entity.WorkflowSchemaStepName, payload json.RawMessage) error {
	if from.String() == "" || payload == nil {
		return cerror.NewF(ctx, cerror.KindBadValidation, "step name and payload are required").LogError()
	}

	workflowRecord, err := uc.store.GetWorkflowByID(ctx, workflowID)
	if err != nil {
		return err
	}

	if workflowRecord.Status != entity.WorkflowStatusSuccess {
		return cerror.NewF(ctx, cerror.KindBadValidation, "workflow is not in %s status", entity.WorkflowStatusSuccess).
			LogError()
	}

	oldPayload := workflowRecord.Input

	for _, s := range workflowRecord.Steps {
		if s.Name == from {
			oldPayload = s.Data
		}
	}

	_, err = uc.orchestrator.StartFrom(ctx, workflowRecord.SchemaName, from, &workflowID, payload)
	if err != nil {
		return err
	}

	err = uc.store.CreateWorkflowHistory(ctx, &entity.WorkflowHistory{
		ID:                uc.store.NewID(),
		CreatedAt:         time.Now().UTC(),
		Type:              entity.WorkflowHistoryTypeRestartFrom,
		Input:             payload,
		InputPrevious:     oldPayload,
		StepName:          from,
		WorkflowID:        workflowID,
		WorkflowStatus:    workflowRecord.Status,
		WorkflowError:     workflowRecord.Error,
		WorkflowErrorKind: workflowRecord.ErrorKind,
		RequestID:         requestIDFromCtx(ctx),
	})

	if err != nil {
		_ = cerror.NewF(
			ctx, cerror.KindInternal, "create history for restart workflow with id=%s from step=%s", workflowID, from).
			LogError()
	}

	return nil
}

// requestIDFromCtx extracts request id value from a given context.
// If there is no requestID in context, nil is returned.
func requestIDFromCtx(ctx context.Context) *string {
	if s, ok := ctx.Value(consts.HeaderXRequestID).(string); ok && s != "" {
		return &s
	}

	return nil
}
