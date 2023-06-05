package workflow

import (
	"context"
	"encoding/json"
	"kafka-polygon/pkg/broker/event"
	"kafka-polygon/pkg/broker/provider"
	"kafka-polygon/pkg/broker/store"
	"kafka-polygon/pkg/cerror"
	"kafka-polygon/pkg/http/consts"
	"kafka-polygon/pkg/workflow/entity"
	"kafka-polygon/pkg/workflow/entrypoint/usecase"
	"sync"
	"time"

	uuid "github.com/satori/go.uuid"
)

// QueueBroker represents a message queue broker abstraction
type QueueBroker interface {
	Send(ctx context.Context, topic string, e event.BaseEvent) error
}

// Store represents a storage with workflow data
type Store interface {
	// NewID generates a new storage compatible ID
	NewID() entity.ID
	// GetWorkflowByID gets workflow by given id
	GetWorkflowByID(ctx context.Context, id entity.ID) (*entity.Workflow, error)
	// CreateWorkflow creates a new workflow record
	CreateWorkflow(ctx context.Context, w *entity.Workflow) error
	// SetWorkflowStatus sets status to existing workflow
	SetWorkflowStatus(ctx context.Context, workflowID entity.ID, status entity.WorkflowStatus) error
	// UpdateWorkflowForce sets values from params to workflow record. All values from params must be set, even nils
	UpdateWorkflowForce(ctx context.Context, workflowID entity.ID, params entity.UpdateWorkflowForceParams) error
	// UpdateWorkflowNotNil sets values from params to workflow record. Only not nil values should be set
	UpdateWorkflowNotNil(ctx context.Context, workflowID entity.ID, params entity.UpdateWorkflowNotNilParams) error
	// PutWorkflowSteps fully replace existing workflow steps
	PutWorkflowSteps(ctx context.Context, workflowID entity.ID, steps []*entity.WorkflowStep) error
}

// ProcessingError is an error that occurred during workflow processing.
type ProcessingError interface {
	error
	// whether retry should be performed for this error
	Retry() bool
	// underlying error
	OriginalError() error
}

// Orchestrator is responsible for executing workflow related tasks
type Orchestrator struct {
	mx              sync.Mutex
	queueBroker     QueueBroker
	store           Store
	workflowSchemas []*entity.WorkflowSchema
	noRetryOnError  bool
}

var _ usecase.Orchestrator = (*Orchestrator)(nil)

// NewOrchestrator creates an Orchestrator instance without workflows.
// Use Orchestrator.AddWorkflow method to add steps.
func NewOrchestrator(qb QueueBroker, s Store) *Orchestrator {
	return &Orchestrator{
		queueBroker: qb,
		store:       s,
	}
}

// SetNoRetryOnError sets noRetries setting to given value.
// If value is set to true, retry policy will not be applied to any errors.
func (o *Orchestrator) SetNoRetryOnError(v bool) {
	o.noRetryOnError = v
}

// AddWorkflowSchema adds a given workflow schema to the list.
// Method calls Validate on the given workflow schema.
// In such a way all workflow schemas in o.workflowSchemas can be assumed to be valid in the future.
// It is the only correct way to add a workflow schema, direct appending to o.workflowSchemas is prohibited.
func (o *Orchestrator) AddWorkflowSchema(ctx context.Context, ws *entity.WorkflowSchema) error {
	o.mx.Lock()
	defer o.mx.Unlock()

	for _, existingSchema := range o.workflowSchemas {
		if existingSchema.Name() == ws.Name() {
			return cerror.NewF(ctx, cerror.KindExist, "workflow schema with name [%s] already exists", ws.Name()).LogError()
		}
	}

	o.workflowSchemas = append(o.workflowSchemas, ws)

	return nil
}

// RemoveWorkflowSchema removes a given workflow schema from the list.
// The returned parameter indicates whether workflow schema was found.
// It is the only correct way to remove workflow schemas, direct manipulating with o.workflowSchemas is prohibited.
func (o *Orchestrator) RemoveWorkflowSchema(wsn entity.WorkflowSchemaName) bool {
	o.mx.Lock()
	defer o.mx.Unlock()

	for i, ws := range o.workflowSchemas {
		if ws.Name() == wsn {
			o.workflowSchemas = append(o.workflowSchemas[:i], o.workflowSchemas[i+1:]...)
			return true
		}
	}

	return false
}

// WorkflowSchema returns a workflow schema with a given workflow schema name.
// Returns error if workflow schema was not found.
func (o *Orchestrator) WorkflowSchema(
	ctx context.Context, wsn entity.WorkflowSchemaName) (*entity.WorkflowSchema, error) {
	o.mx.Lock()
	defer o.mx.Unlock()

	for _, ws := range o.workflowSchemas {
		if ws.Name() == wsn {
			return ws, nil
		}
	}

	return nil, cerror.NewF(ctx, cerror.KindNotExist, "workflow schema with name [%s] doesn't exist", wsn).
		LogError()
}

// Start starts a new workflow with a given input from the first step
func (o *Orchestrator) Start(
	ctx context.Context, schemaName entity.WorkflowSchemaName, payload json.RawMessage) (entity.ID, error) {
	schema, err := o.WorkflowSchema(ctx, schemaName)
	if err != nil {
		return "", err
	}

	return o.runWorkflow(ctx, &runWorkflowParam{
		workflowSchema:     schema,
		workflowSchemaStep: schema.FirstStep(),
		payload:            payload,
	})
}

// StartFrom starts a new workflow with a given input from the given step.
// You may pass parentID if the workflow is based on existing one.
func (o *Orchestrator) StartFrom(
	ctx context.Context,
	schemaName entity.WorkflowSchemaName,
	stepName entity.WorkflowSchemaStepName,
	parentID *entity.ID,
	payload json.RawMessage) (entity.ID, error) {
	schema, err := o.WorkflowSchema(ctx, schemaName)
	if err != nil {
		return "", err
	}

	step, ok := schema.Step(stepName)
	if !ok {
		return "", cerror.NewF(
			ctx, cerror.KindInternal, "step [%s] doesn't exist in the workflow schema [%s]", stepName, schemaName).
			LogError()
	}

	return o.runWorkflow(ctx, &runWorkflowParam{
		workflowSchema:     schema,
		workflowSchemaStep: step,
		payload:            payload,
		parentWorkflowID:   parentID,
	})
}

// Restart restarts an existing workflow from the first step
func (o *Orchestrator) Restart(
	ctx context.Context,
	workflowID entity.ID,
	schemaName entity.WorkflowSchemaName,
	payload json.RawMessage) (entity.ID, error) {
	schema, err := o.WorkflowSchema(ctx, schemaName)
	if err != nil {
		return "", err
	}

	from := schema.FirstStep()

	return o.RestartFrom(ctx, workflowID, schemaName, from.Name(), payload)
}

// RestartFrom restarts an existing workflow from a given step
func (o *Orchestrator) RestartFrom(
	ctx context.Context,
	workflowID entity.ID,
	schemaName entity.WorkflowSchemaName,
	stepName entity.WorkflowSchemaStepName,
	payload json.RawMessage) (entity.ID, error) {
	schema, err := o.WorkflowSchema(ctx, schemaName)
	if err != nil {
		return "", err
	}

	step, ok := schema.Step(stepName)
	if !ok {
		return "", cerror.NewF(
			ctx, cerror.KindInternal, "step [%s] doesn't exist in the workflow schema [%s]", stepName, schemaName).
			LogError()
	}

	return o.runWorkflow(ctx, &runWorkflowParam{
		workflowSchema:     schema,
		workflowSchemaStep: step,
		payload:            payload,
		workflowID:         &workflowID,
	})
}

// handleWorkflowEvent handles queue topic messages.
// Each message intends to execute some workflow step with some payload.
//
//nolint:gocyclo
func (o *Orchestrator) handleWorkflowEvent(ctx context.Context, e event.WorkflowEvent, eventData store.EventProcessData) error {
	if eventData.Status == store.EventStatusHandled {
		_ = cerror.NewF(ctx, cerror.KindExist,
			"skipped duplicate workflow event. event_id=%s. event_status=%s", e.GetID(), eventData.Status).LogWarn()
		return nil
	}

	eventWorkflow := e.GetWorkflow()
	schemaName := entity.WorkflowSchemaName(eventWorkflow.Schema)
	stepName := entity.WorkflowSchemaStepName(eventWorkflow.Step)
	workflowID := entity.ID(eventWorkflow.ID)

	schema, err := o.WorkflowSchema(ctx, schemaName)
	if err != nil {
		if saveErr := o.saveWorkflowError(ctx, workflowID, err); saveErr != nil {
			_ = cerror.NewF(ctx, cerror.KindInternal,
				"couldn't save workflow unexisting schema error. workflow=%s step=%s workflow_id=%s. error=%s",
				schemaName, stepName, workflowID, saveErr.Error()).
				LogError()
		}

		return err
	}

	schemaStep, ok := schema.Step(stepName)
	if !ok {
		err := cerror.NewF(
			ctx, cerror.KindNotExist, "step [%s] doesn't exist in the workflow schema [%s]", stepName, schemaName).
			LogError()
		if saveErr := o.saveWorkflowError(ctx, workflowID, err); saveErr != nil {
			_ = cerror.NewF(ctx, cerror.KindInternal,
				"couldn't save workflow unexisting step error. workflow=%s step=%s workflow_id=%s. error=%s",
				schemaName, stepName, workflowID, saveErr.Error()).
				LogError()
		}

		return err
	}

	workflow, err := o.store.GetWorkflowByID(ctx, workflowID)
	if err != nil {
		if cerror.IsNotExist(err) {
			return err
		}

		if saveErr := o.saveWorkflowError(ctx, workflowID, err); saveErr != nil {
			_ = cerror.NewF(ctx, cerror.KindInternal,
				"couldn't save get workflow by id error. workflow=%s step=%s workflow_id=%s. error=%s",
				schemaName, stepName, workflowID, saveErr.Error()).
				LogError()
		}

		return entity.NewProcessingError(err).SetRetry(true)
	}

	if workflow.Status != entity.WorkflowStatusInProgress {
		if err := o.store.SetWorkflowStatus(ctx, workflowID, entity.WorkflowStatusInProgress); err != nil {
			if saveErr := o.saveWorkflowError(ctx, workflowID, err); saveErr != nil {
				_ = cerror.NewF(ctx, cerror.KindInternal,
					"couldn't save set workflow status error. workflow=%s step=%s workflow_id=%s. error=%s",
					schemaName, stepName, workflowID, saveErr.Error()).
					LogError()
			}

			return entity.NewProcessingError(err).SetRetry(true)
		}
	}

	if err := o.appendWorkflowStep(ctx, e, workflow); err != nil {
		if saveErr := o.saveWorkflowError(ctx, workflowID, err); saveErr != nil {
			_ = cerror.NewF(ctx, cerror.KindInternal,
				"couldn't save append workflow step error. workflow=%s step=%s workflow_id=%s. error=%s",
				schemaName, stepName, workflowID, saveErr.Error()).
				LogError()
		}

		return entity.NewProcessingError(err).SetRetry(true)
	}

	if err := o.processWorkflowEvent(ctx, workflow, schema, schemaStep, e); err != nil {
		if saveErr := o.saveWorkflowError(ctx, workflowID, err); saveErr != nil {
			_ = cerror.NewF(ctx, cerror.KindInternal,
				"couldn't save workflow event processing failed info. workflow=%s step=%s workflow_id=%s. error=%s",
				schemaName, stepName, workflowID, saveErr.Error()).
				LogError()
		}

		return err
	}

	return nil
}

func (o *Orchestrator) saveWorkflowError(ctx context.Context, wfID entity.ID, err error) error {
	return o.store.UpdateWorkflowForce(ctx, wfID, entity.UpdateWorkflowForceParams{
		Status:    entity.WorkflowStatusFailed,
		Error:     entity.PointerWorkflowErrorMsg(err.Error()),
		ErrorKind: entity.PointerWorkflowErrorKind(cerror.ErrKind(err).String()),
	})
}

// decorateQHandlerWithErrRetry returns a func that calls a given function and transforms returned error
// according to the error retry logic.
// New function returns:
// - the original error. It assumes subsequent returning message to the queue and retry handling after some period
// - nil. It assumes committing message to the queue
// By default all the messages are committed, even handled with error.
// To perform retry you must explicitly return ProcessingError with Retry() = true
// from underlying HandlerWorkflow.
func (o *Orchestrator) decorateQHandlerWithErrRetry(f provider.HandlerWorkflow) provider.HandlerWorkflow {
	return func(ctx context.Context, e event.WorkflowEvent, ed store.EventProcessData) error {
		err := f(ctx, e, ed)
		if err == nil {
			return nil
		}

		if perr, ok := err.(ProcessingError); ok && perr.Retry() && !o.noRetryOnError {
			return perr.OriginalError()
		}

		return nil
	}
}

// QueueEventHandler returns functions that handles workflow queue events.
func (o *Orchestrator) QueueEventHandler() provider.HandlerWorkflow {
	return o.decorateQHandlerWithErrRetry(o.handleWorkflowEvent)
}

// runWorkflowParam represents parameters to run a workflow
type runWorkflowParam struct {
	// schema to run through
	workflowSchema *entity.WorkflowSchema
	// schema step from which to run the workflow
	workflowSchemaStep entity.WorkflowSchemaStep
	// payload to run the step with
	payload json.RawMessage
	// workflowID must be set only if the workflow already exists in db
	workflowID *entity.ID
	// parentWorkflowID must be set only if the workflow is based on some existing workflow
	parentWorkflowID *entity.ID
}

// runWorkflow starts a new workflow or runs existing one from the given step.
// If workflowID is not set in runFlowParam the flow is considered to be new
// and will be saved in db with an auto generated id.
// Returns a workflowID that is equal to passed param or newly generated for new workflows.
func (o *Orchestrator) runWorkflow(ctx context.Context, param *runWorkflowParam) (entity.ID, error) {
	isNewWorkflow := param.workflowID == nil
	now := time.Now().UTC()

	workflowID := o.store.NewID()
	if param.workflowID != nil {
		workflowID = *param.workflowID
	}

	if isNewWorkflow {
		w := &entity.Workflow{
			ID:         workflowID,
			CreatedAt:  now,
			UpdatedAt:  now,
			ParentID:   param.parentWorkflowID,
			SchemaName: param.workflowSchema.Name(),
			Status:     entity.WorkflowStatusInProgress,
			Input:      param.payload,
			RequestID:  requestIDFromCtx(ctx),
		}
		if err := o.store.CreateWorkflow(ctx, w); err != nil {
			return "", err
		}
	}

	e := &event.WorkflowData{
		ID: uuid.NewV4().String(),
		Workflow: event.Workflow{
			ID:          workflowID.String(),
			Schema:      param.workflowSchema.Name().String(),
			Step:        param.workflowSchemaStep.Name().String(),
			StepPayload: param.payload,
		},
	}

	if err := o.queueBroker.Send(ctx, param.workflowSchemaStep.Topic().String(), e); err != nil {
		_ = cerror.NewF(ctx, cerror.KindInternal,
			"workflow start event was not sent. workflow=%s step=%s workflow_id=%s. error=%s",
			param.workflowSchema.Name(), param.workflowSchemaStep.Name(), e.Workflow.ID, err.Error()).
			LogError()

		if saveErr := o.store.UpdateWorkflowForce(ctx, workflowID, entity.UpdateWorkflowForceParams{
			Status:    entity.WorkflowStatusFailed,
			Error:     entity.PointerWorkflowErrorMsg(err.Error()),
			ErrorKind: entity.PointerWorkflowErrorKind(cerror.ErrKind(err).String()),
		}); saveErr != nil {
			_ = cerror.NewF(ctx, cerror.KindInternal,
				"couldn't save workflow start event failed info. workflow=%s step=%s workflow_id=%s. error=%s",
				param.workflowSchema.Name(), param.workflowSchemaStep.Name(), e.Workflow.ID, saveErr.Error()).
				LogError()
		}

		return "", err
	}

	return workflowID, nil
}

// appendWorkflowStep appends workflow step to existing workflow based on received event
// and saves to store
func (o *Orchestrator) appendWorkflowStep(
	ctx context.Context, e event.WorkflowEvent, workflow *entity.Workflow) error {
	eventWorkflow := e.GetWorkflow()

	if len(workflow.Steps) > 0 {
		// if step is already present in db replace it with new one (for retries with new inputs)
		eventStep := entity.WorkflowSchemaStepName(eventWorkflow.Step)
		if currStep := workflow.Steps[len(workflow.Steps)-1]; currStep.Name == eventStep {
			workflow.Steps = workflow.Steps[:len(workflow.Steps)-1]
		}
	}

	workflow.Steps = append(workflow.Steps, &entity.WorkflowStep{
		CreatedAt: time.Now().UTC(),
		Name:      entity.WorkflowSchemaStepName(eventWorkflow.Step),
		Data:      eventWorkflow.StepPayload,
		Metadata: entity.WorkflowStepMetadata{
			Version: e.GetMeta().Version,
		},
	})

	return o.store.PutWorkflowSteps(ctx, workflow.ID, workflow.Steps)
}

// processWorkflowEvent extracts and runs step's underlying worker(business logic executor).
// If worker returns no error, next step (if it exists) is pushed to the queue.
func (o *Orchestrator) processWorkflowEvent(
	ctx context.Context,
	workflow *entity.Workflow,
	schema *entity.WorkflowSchema,
	step entity.WorkflowSchemaStep,
	e event.WorkflowEvent) error {
	eventWorkflow := e.GetWorkflow()
	workflowID := entity.ID(eventWorkflow.ID)

	nextPayload, err := step.Worker().Run(ctx, e)
	if err != nil {
		return err
	}

	// subsequent errors shouldn't be retried to avoid business logic call duplication.
	// so don't wrap in NewProcessingError(err).SetRetry(true)

	nextStep, nextExists := schema.NextStep(step.Name())
	if !nextExists {
		err = o.store.SetWorkflowStatus(ctx, workflowID, entity.WorkflowStatusSuccess)
		if err != nil {
			return cerror.NewF(ctx, cerror.KindInternal,
				"workflow_id=%s was completed but failed to update it's status in DB: %s", workflowID, err.Error()).
				LogError()
		}

		return nil
	}

	nextStepEvent := &event.WorkflowData{
		ID: uuid.NewV4().String(),
		Workflow: event.Workflow{
			ID:          workflowID.String(),
			Schema:      schema.Name().String(),
			Step:        nextStep.Name().String(),
			StepPayload: nextPayload,
		},
	}

	if brokerErr := o.queueBroker.Send(ctx, nextStep.Topic().String(), nextStepEvent); brokerErr != nil {
		err = cerror.NewF(ctx, cerror.KindInternal,
			"next step was not sent. workflow=%s step=%s workflow_id=%s. error=%s",
			eventWorkflow.Schema, eventWorkflow.Step, eventWorkflow.ID, brokerErr.Error()).
			LogError()

		// save next step data to db to have an opportunity
		// to restart the workflow from the next step later
		workflow.Steps = append(workflow.Steps, &entity.WorkflowStep{
			CreatedAt: time.Now().UTC(),
			Name:      entity.WorkflowSchemaStepName(nextStepEvent.Workflow.Step),
			Data:      nextStepEvent.Workflow.StepPayload,
			Metadata: entity.WorkflowStepMetadata{
				Version: nextStepEvent.Metadata.Version,
			},
		})

		if saveErr := o.store.UpdateWorkflowNotNil(ctx, workflowID, entity.UpdateWorkflowNotNilParams{
			Steps: workflow.Steps,
		}); saveErr != nil {
			_ = cerror.NewF(ctx, cerror.KindInternal,
				"unable to save next step data for the unsent next step. workflow_id=%s. error=%s",
				workflowID, saveErr.Error()).
				LogError()
		}

		return err
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
