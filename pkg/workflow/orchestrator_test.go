package workflow_test

import (
	"context"
	"encoding/json"
	"fmt"
	"kafka-polygon/pkg/broker/event"
	pkgStore "kafka-polygon/pkg/broker/store"
	"kafka-polygon/pkg/cmd/metadata"
	"kafka-polygon/pkg/converto"
	"kafka-polygon/pkg/http/consts"
	"kafka-polygon/pkg/workflow"
	"kafka-polygon/pkg/workflow/entity"
	"testing"

	"github.com/tj/assert"
)

const (
	_reqID = "1"
)

var (
	_bgCtx          = context.Background()
	_bgCtxWithReqID = context.WithValue(_bgCtx, consts.HeaderXRequestID, _reqID) //nolint:staticcheck
)

func TestAddWorkflowSchema(t *testing.T) {
	o := workflow.NewOrchestrator(nil, nil)

	schemaName := entity.WorkflowSchemaName("schema1")
	schemaStep := entity.NewWorkflowSchemaSimpleStep("step1", "topic1", &stepWorkerTest{})
	schema, err := entity.NewWorkflowSchema(_bgCtx, schemaName, schemaStep)
	assert.NoError(t, err)

	err = o.AddWorkflowSchema(_bgCtx, schema)
	assert.NoError(t, err)

	// add schema with the same name
	err = o.AddWorkflowSchema(_bgCtx, schema)
	assert.Equal(t, fmt.Sprintf("workflow schema with name [%s] already exists", schemaName), err.Error())

	otherSchema, err := entity.NewWorkflowSchema(_bgCtx, schemaName+"1", schemaStep)
	assert.NoError(t, err)

	err = o.AddWorkflowSchema(_bgCtx, otherSchema)
	assert.NoError(t, err)
}

func TestRemoveWorkflowSchema(t *testing.T) {
	o := workflow.NewOrchestrator(nil, nil)

	schemaName := entity.WorkflowSchemaName("schema1")
	schemaStep := entity.NewWorkflowSchemaSimpleStep("step1", "topic1", &stepWorkerTest{})
	schema, err := entity.NewWorkflowSchema(_bgCtx, schemaName, schemaStep)
	assert.NoError(t, err)

	err = o.AddWorkflowSchema(_bgCtx, schema)
	assert.NoError(t, err)

	assert.False(t, o.RemoveWorkflowSchema(schemaName+"1"))
	assert.True(t, o.RemoveWorkflowSchema(schemaName))
}

func TestWorkflowSchema(t *testing.T) {
	o := workflow.NewOrchestrator(nil, nil)

	schemaName := entity.WorkflowSchemaName("schema1")
	schemaStep := entity.NewWorkflowSchemaSimpleStep("step1", "topic1", &stepWorkerTest{})
	schema, err := entity.NewWorkflowSchema(_bgCtx, schemaName, schemaStep)
	assert.NoError(t, err)

	err = o.AddWorkflowSchema(_bgCtx, schema)
	assert.NoError(t, err)

	_, err = o.WorkflowSchema(_bgCtx, schemaName+"1")
	assert.Equal(t, fmt.Sprintf("workflow schema with name [%s] doesn't exist", schemaName+"1"), err.Error())

	actSchema, err := o.WorkflowSchema(_bgCtx, schemaName)
	assert.NoError(t, err)
	assert.Equal(t, schema, actSchema)
}

func TestStart(t *testing.T) {
	// start with not existing schema
	testWorkflowRunWithNotExistingSchema(t, func(o *workflow.Orchestrator, wsn entity.WorkflowSchemaName) error {
		_, err := o.Start(_bgCtx, wsn, nil)
		return err
	})

	schema, steps, _ := defSchema(t)
	// start with valid schema
	testWorkflowRunWithParameters(t, schema, &testWorkflowRunParameters{
		expStep: steps[0],
		orchestratorCaller: func(ctx context.Context, o *workflow.Orchestrator, payload json.RawMessage) (entity.ID, error) {
			return o.Start(ctx, schema.Name(), payload)
		},
	})
}

func TestStartFrom(t *testing.T) {
	// start with not existing schema
	testWorkflowRunWithNotExistingSchema(t, func(o *workflow.Orchestrator, wsn entity.WorkflowSchemaName) error {
		_, err := o.StartFrom(_bgCtx, wsn, "", nil, nil)
		return err
	})

	schema, steps, _ := defSchema(t)

	// start from not existing step
	testWorkflowRunFromNotExistingStep(t, func(o *workflow.Orchestrator, wssn entity.WorkflowSchemaStepName) error {
		_, err := o.StartFrom(_bgCtx, schema.Name(), wssn, nil, nil)
		return err
	})

	parentWorkflowID := entity.ID("parent-123")
	// start with valid schema
	testWorkflowRunWithParameters(t, schema, &testWorkflowRunParameters{
		expStep:             steps[1],
		expParentWorkflowID: &parentWorkflowID,
		orchestratorCaller: func(ctx context.Context, o *workflow.Orchestrator, payload json.RawMessage) (entity.ID, error) {
			return o.StartFrom(ctx, schema.Name(), steps[1].Name(), &parentWorkflowID, payload)
		},
	})
}

func TestRestart(t *testing.T) {
	// start with not existing schema
	testWorkflowRunWithNotExistingSchema(t, func(o *workflow.Orchestrator, wsn entity.WorkflowSchemaName) error {
		_, err := o.Restart(_bgCtx, "", wsn, nil)
		return err
	})

	schema, steps, _ := defSchema(t)
	workflowID := entity.ID("wf-123")
	// start with valid schema
	testWorkflowRunWithParameters(t, schema, &testWorkflowRunParameters{
		expStep:           steps[0],
		restartWorkflowID: &workflowID,
		orchestratorCaller: func(ctx context.Context, o *workflow.Orchestrator, payload json.RawMessage) (entity.ID, error) {
			return o.Restart(ctx, workflowID, schema.Name(), payload)
		},
	})
}

func TestRestartFrom(t *testing.T) {
	// start with not existing schema
	testWorkflowRunWithNotExistingSchema(t, func(o *workflow.Orchestrator, wsn entity.WorkflowSchemaName) error {
		_, err := o.RestartFrom(_bgCtx, "", wsn, "", nil)
		return err
	})

	schema, steps, _ := defSchema(t)

	// start from not existing step
	testWorkflowRunFromNotExistingStep(t, func(o *workflow.Orchestrator, wssn entity.WorkflowSchemaStepName) error {
		_, err := o.RestartFrom(_bgCtx, "", schema.Name(), wssn, nil)
		return err
	})

	workflowID := entity.ID("wf-123")
	// start with valid schema
	testWorkflowRunWithParameters(t, schema, &testWorkflowRunParameters{
		expStep:           steps[1],
		restartWorkflowID: &workflowID,
		orchestratorCaller: func(ctx context.Context, o *workflow.Orchestrator, payload json.RawMessage) (entity.ID, error) {
			return o.RestartFrom(ctx, workflowID, schema.Name(), steps[1].Name(), payload)
		},
	})
}

func testWorkflowRunWithNotExistingSchema(
	t *testing.T,
	caller func(*workflow.Orchestrator, entity.WorkflowSchemaName) error) {
	t.Helper()

	o := workflow.NewOrchestrator(nil, nil)

	schema, _, _ := defSchema(t)

	err := o.AddWorkflowSchema(_bgCtx, schema)
	assert.NoError(t, err)

	// start with not existing schema
	err = caller(o, schema.Name()+"1")
	assert.Equal(t, fmt.Sprintf("workflow schema with name [%s] doesn't exist", schema.Name()+"1"), err.Error())
}

func testWorkflowRunFromNotExistingStep(
	t *testing.T,
	caller func(*workflow.Orchestrator, entity.WorkflowSchemaStepName) error) {
	t.Helper()

	o := workflow.NewOrchestrator(nil, nil)

	schema, _, _ := defSchema(t)

	err := o.AddWorkflowSchema(_bgCtx, schema)
	assert.NoError(t, err)

	// start with not existing schema
	err = caller(o, schema.FirstStep().Name()+"1")
	assert.Equal(
		t,
		fmt.Sprintf("step [%s] doesn't exist in the workflow schema [%s]",
			schema.FirstStep().Name()+"1", schema.Name()),
		err.Error())
}

func TestQueueEventHandler(t *testing.T) {
	t.Run("handle event with not existing schema", func(t *testing.T) {
		testHandleEventWithNotExistingSchema(t, func(o *workflow.Orchestrator, wsn entity.WorkflowSchemaName) error {
			return o.QueueEventHandler()(
				_bgCtx,
				&event.WorkflowData{Workflow: event.Workflow{Schema: wsn.String()}},
				pkgStore.EventProcessData{Status: pkgStore.EventStatusNew})
		})
	})

	t.Run("start from not existing step", func(t *testing.T) {
		schema, _, _ := defSchema(t)
		testHandleEventWithExistingStep(t, func(o *workflow.Orchestrator, wssn entity.WorkflowSchemaStepName) error {
			return o.QueueEventHandler()(
				_bgCtx,
				&event.WorkflowData{
					Workflow: event.Workflow{
						Schema: schema.Name().String(),
						Step:   wssn.String(),
					},
				},
				pkgStore.EventProcessData{Status: pkgStore.EventStatusNew})
		})
	})

	t.Run("skip duplicate messages", func(t *testing.T) {
		schema, _, _ := defSchema(t)
		store := new(mockStore)
		o := workflow.NewOrchestrator(nil, store)

		assert.NoError(t, o.AddWorkflowSchema(_bgCtx, schema))

		workflowsForSkip := []*entity.Workflow{
			{Status: entity.WorkflowStatusSuccess},
			{Status: entity.WorkflowStatusInProgress},
		}
		expID := entity.ID("wf-123")

		for _, w := range workflowsForSkip {
			store.getWorkflowByIDFunc = func(ctx context.Context, id entity.ID) (*entity.Workflow, error) {
				assert.Equal(t, expID, id)
				return w, nil
			}

			actErr := o.QueueEventHandler()(
				_bgCtx,
				&event.WorkflowData{
					Workflow: event.Workflow{
						ID:     expID.String(),
						Schema: schema.Name().String(),
						Step:   schema.FirstStep().Name().String(),
					},
				}, pkgStore.EventProcessData{Status: pkgStore.EventStatusHandled})

			assert.Nil(t, actErr)
		}
	})

	t.Run("set workflow status in progress if it has another status", func(t *testing.T) {
		schema, _, _ := defSchema(t)
		store := new(mockStore)
		o := workflow.NewOrchestrator(nil, store)

		assert.NoError(t, o.AddWorkflowSchema(_bgCtx, schema))

		statuses := []entity.WorkflowStatus{
			entity.WorkflowStatusSuccess,
			entity.WorkflowStatusFailed,
			entity.WorkflowStatusInProgress,
		}
		expID := entity.ID("wf-123")

		for _, s := range statuses {
			expIsCalled := s != entity.WorkflowStatusInProgress
			actIsCalled := false

			store.getWorkflowByIDFunc = func(ctx context.Context, id entity.ID) (*entity.Workflow, error) {
				return &entity.Workflow{ID: id, Status: s}, nil
			}
			store.setWorkflowStatusFunc = func(ctx context.Context, workflowID entity.ID, status entity.WorkflowStatus) error {
				actIsCalled = true
				assert.Equal(t, expID, workflowID)
				assert.Equal(t, entity.WorkflowStatusInProgress, status)

				return nil
			}
			// minimal implementation to stop after setWorkflowStatus
			store.putWorkflowStepsFunc = func(ctx context.Context, workflowID entity.ID, steps []*entity.WorkflowStep) error {
				return fmt.Errorf("stop")
			}
			store.updateWorkflowForceFunc = func(ctx context.Context, workflowID entity.ID, params entity.UpdateWorkflowForceParams) error {
				return nil
			}

			_ = o.QueueEventHandler()(
				_bgCtx,
				&event.WorkflowData{
					Workflow: event.Workflow{
						ID:     expID.String(),
						Schema: schema.Name().String(),
						Step:   schema.FirstStep().Name().String(),
					},
				}, pkgStore.EventProcessData{Status: pkgStore.EventStatusNew})

			assert.Equal(t, expIsCalled, actIsCalled)
		}
	})

	t.Run("step is appended to previous steps in store", func(t *testing.T) {
		schema, _, _ := defSchema(t)
		store := new(mockStore)
		o := workflow.NewOrchestrator(nil, store)

		assert.NoError(t, o.AddWorkflowSchema(_bgCtx, schema))

		wf := &entity.Workflow{ID: entity.ID("wf-123"), Status: entity.WorkflowStatusInProgress}
		expSteps := []*entity.WorkflowStep{
			{
				Name:     "step1",
				Data:     []byte("{}"),
				Metadata: entity.WorkflowStepMetadata{Version: "v1.0.1"},
			},
		}

		store.getWorkflowByIDFunc = func(ctx context.Context, id entity.ID) (*entity.Workflow, error) {
			return wf, nil
		}
		store.updateWorkflowForceFunc = func(ctx context.Context, workflowID entity.ID, params entity.UpdateWorkflowForceParams) error {
			return nil
		}
		store.putWorkflowStepsFunc = func(ctx context.Context, workflowID entity.ID, steps []*entity.WorkflowStep) error {
			assert.Equal(t, wf.ID, workflowID)
			assert.Equal(t, len(expSteps), len(steps))

			for i := range steps {
				assert.False(t, steps[i].CreatedAt.IsZero())
				expSteps[i].CreatedAt = steps[i].CreatedAt
			}

			assert.Equal(t, expSteps, steps)

			return fmt.Errorf("stop")
		}

		eventData := &event.WorkflowData{
			Workflow: event.Workflow{
				ID:          wf.ID.String(),
				Schema:      schema.Name().String(),
				Step:        schema.FirstStep().Name().String(),
				StepPayload: expSteps[0].Data,
			},
			Metadata: metadata.Meta{
				Version: expSteps[0].Metadata.Version,
			},
		}
		_ = o.QueueEventHandler()(_bgCtx, eventData, pkgStore.EventProcessData{Status: pkgStore.EventStatusNew})

		// step shouldn't be appended if the last existing step is already present with the same name
		wf.Steps = expSteps

		_ = o.QueueEventHandler()(_bgCtx, eventData, pkgStore.EventProcessData{Status: pkgStore.EventStatusNew})
	})

	t.Run("worker is called and result is sent to broker", func(t *testing.T) {
		schema, steps, worker := defSchema(t)
		store := new(mockStore)
		broker := new(mockQueueBroker)
		o := workflow.NewOrchestrator(broker, store)

		assert.NoError(t, o.AddWorkflowSchema(_bgCtx, schema))

		wfID := entity.ID("123")

		store.getWorkflowByIDFunc = func(ctx context.Context, id entity.ID) (*entity.Workflow, error) {
			return &entity.Workflow{ID: wfID, Status: entity.WorkflowStatusInProgress}, nil
		}
		store.putWorkflowStepsFunc = func(ctx context.Context, workflowID entity.ID, steps []*entity.WorkflowStep) error {
			return nil
		}
		broker.sendFunc = func(ctx context.Context, topic string, e event.BaseEvent) error {
			actEvent, ok := e.(*event.WorkflowData)

			assert.True(t, ok)
			assert.NotEmpty(t, e.GetID())
			assert.Equal(t, steps[1].Topic().String(), topic)
			assert.Equal(t, event.Workflow{
				ID:          wfID.String(),
				Schema:      schema.Name().String(),
				Step:        steps[1].Name().String(),
				StepPayload: worker.lastResult,
			}, actEvent.GetWorkflow())

			return nil
		}

		eventData := &event.WorkflowData{
			Workflow: event.Workflow{
				ID:          wfID.String(),
				Schema:      schema.Name().String(),
				Step:        schema.FirstStep().Name().String(),
				StepPayload: []byte("{}"),
			},
		}

		err := o.QueueEventHandler()(_bgCtx, eventData, pkgStore.EventProcessData{Status: pkgStore.EventStatusNew})

		assert.NoError(t, err)
		assert.Equal(t, 1, worker.runCount)

		// workflow data is updated in db when failed to send msg to broker
		sendFuncErr := fmt.Errorf("err")
		broker.sendFunc = func(ctx context.Context, topic string, e event.BaseEvent) error {
			return sendFuncErr
		}

		isUpdateNotNilCalled := false
		isUpdateForceCalled := false
		store.updateWorkflowNotNilFunc = func(ctx context.Context, workflowID entity.ID, params entity.UpdateWorkflowNotNilParams) error {
			isUpdateNotNilCalled = true

			assert.Equal(t, wfID, workflowID)
			assert.Equal(t, 2, len(params.Steps))

			expSteps := []*entity.WorkflowStep{
				{CreatedAt: params.Steps[0].CreatedAt, Name: steps[0].Name(), Data: eventData.Workflow.StepPayload, Metadata: params.Steps[0].Metadata},
				{CreatedAt: params.Steps[1].CreatedAt, Name: steps[1].Name(), Data: worker.lastResult, Metadata: params.Steps[1].Metadata},
			}

			assert.Equal(t, entity.UpdateWorkflowNotNilParams{
				Steps: expSteps,
			}, params)

			return nil
		}
		store.updateWorkflowForceFunc = func(ctx context.Context, workflowID entity.ID, params entity.UpdateWorkflowForceParams) error {
			isUpdateForceCalled = true

			assert.Equal(t, wfID, workflowID)
			assert.Equal(t, entity.WorkflowStatusFailed, params.Status)
			assert.NotEmpty(t, params.Error)
			assert.Equal(t,
				fmt.Errorf("next step was not sent. workflow=%s step=%s workflow_id=%s. error=%s",
					schema.Name(), steps[0].Name(), wfID, sendFuncErr).Error(),
				params.Error.String())
			assert.NotEmpty(t, params.ErrorKind)

			return nil
		}

		err = o.QueueEventHandler()(_bgCtx, &event.WorkflowData{
			Workflow: event.Workflow{
				ID:          wfID.String(),
				Schema:      schema.Name().String(),
				Step:        schema.FirstStep().Name().String(),
				StepPayload: []byte("{}"),
			},
		}, pkgStore.EventProcessData{Status: pkgStore.EventStatusNew})

		assert.NoError(t, err)
		assert.True(t, isUpdateNotNilCalled)
		assert.True(t, isUpdateForceCalled)
	})

	t.Run("workflow is completed after the last step", func(t *testing.T) {
		schema, steps, _ := defSchema(t)
		store := new(mockStore)
		broker := new(mockQueueBroker)
		o := workflow.NewOrchestrator(broker, store)

		assert.NoError(t, o.AddWorkflowSchema(_bgCtx, schema))

		wfID := entity.ID("123")

		store.getWorkflowByIDFunc = func(ctx context.Context, id entity.ID) (*entity.Workflow, error) {
			return &entity.Workflow{ID: wfID, Status: entity.WorkflowStatusInProgress}, nil
		}
		store.putWorkflowStepsFunc = func(ctx context.Context, workflowID entity.ID, steps []*entity.WorkflowStep) error {
			return nil
		}

		isSetStatusCalled := false
		store.setWorkflowStatusFunc = func(ctx context.Context, workflowID entity.ID, status entity.WorkflowStatus) error {
			isSetStatusCalled = true

			assert.Equal(t, wfID, workflowID)
			assert.Equal(t, entity.WorkflowStatusSuccess, status)

			return nil
		}

		err := o.QueueEventHandler()(_bgCtx, &event.WorkflowData{
			Workflow: event.Workflow{
				ID:          wfID.String(),
				Schema:      schema.Name().String(),
				Step:        steps[1].Name().String(),
				StepPayload: []byte("{}"),
			},
		}, pkgStore.EventProcessData{Status: pkgStore.EventStatusNew})

		assert.NoError(t, err)
		assert.True(t, isSetStatusCalled)
	})
}

func defSchema(t *testing.T) (*entity.WorkflowSchema, []entity.WorkflowSchemaStep, *stepWorkerTest) {
	t.Helper()

	worker := new(stepWorkerTest)
	schemaName := entity.WorkflowSchemaName("schema1")
	schemaSteps := []entity.WorkflowSchemaStep{
		entity.NewWorkflowSchemaSimpleStep("step1", "topic1", worker),
		entity.NewWorkflowSchemaSimpleStep("step2", "topic2", worker),
	}
	schema, err := entity.NewWorkflowSchema(_bgCtx, schemaName, schemaSteps...)
	assert.NoError(t, err)

	return schema, schemaSteps, worker
}

func testHandleEventWithNotExistingSchema(
	t *testing.T,
	caller func(*workflow.Orchestrator, entity.WorkflowSchemaName) error) {
	t.Helper()

	schema, _, _ := defSchema(t)

	unexistingSchema := schema.Name() + "1"
	expErr := fmt.Errorf("workflow schema with name [%s] doesn't exist", unexistingSchema)

	var isSaverCalled bool

	o := workflow.NewOrchestrator(nil, &mockStore{
		updateWorkflowForceFunc: func(ctx context.Context, _ entity.ID, params entity.UpdateWorkflowForceParams) error {
			isSaverCalled = true

			assert.Equal(t, entity.WorkflowStatusFailed, params.Status)
			assert.NotEmpty(t, params.Error)
			assert.Equal(t, expErr.Error(), params.Error.String())
			assert.NotEmpty(t, params.ErrorKind)

			return nil
		},
	})

	err := o.AddWorkflowSchema(_bgCtx, schema)
	assert.NoError(t, err)

	// start with not existing schema
	err = caller(o, unexistingSchema)
	assert.NoError(t, err)
	assert.True(t, isSaverCalled)
}

func testHandleEventWithExistingStep(
	t *testing.T,
	caller func(*workflow.Orchestrator, entity.WorkflowSchemaStepName) error) {
	t.Helper()

	schema, _, _ := defSchema(t)
	unexistingStep := schema.FirstStep().Name() + "1"
	expErr := fmt.Errorf("step [%s] doesn't exist in the workflow schema [%s]", unexistingStep, schema.Name())

	var isSaverCalled bool

	o := workflow.NewOrchestrator(nil, &mockStore{
		updateWorkflowForceFunc: func(ctx context.Context, _ entity.ID, params entity.UpdateWorkflowForceParams) error {
			isSaverCalled = true

			assert.Equal(t, entity.WorkflowStatusFailed, params.Status)
			assert.NotEmpty(t, params.Error)
			assert.Equal(t, expErr.Error(), params.Error.String())
			assert.NotEmpty(t, params.ErrorKind)

			return nil
		},
	})

	err := o.AddWorkflowSchema(_bgCtx, schema)
	assert.NoError(t, err)

	// start with not existing schema
	err = caller(o, unexistingStep)
	assert.NoError(t, err)
	assert.True(t, isSaverCalled)
}

type testWorkflowRunParameters struct {
	// step expected to run workflow from (required)
	expStep entity.WorkflowSchemaStep
	// expected parent workflow id (optional)
	expParentWorkflowID *entity.ID
	// workflow id which is intended to be restarted (optional. pass only for restart, restartFrom operations)
	restartWorkflowID *entity.ID
	// function that will call appropriate orchestrator's method (required)
	orchestratorCaller func(ctx context.Context, o *workflow.Orchestrator, payload json.RawMessage) (entity.ID, error)
}

func testWorkflowRunWithParameters(t *testing.T, schema *entity.WorkflowSchema, params *testWorkflowRunParameters) {
	t.Helper()

	isStoreCalled := false
	isBrokerCalled := false

	expPayload := []byte("{}")
	expWorkflowID := entity.ID("123")
	isNewWorkflow := params.restartWorkflowID == nil

	if !isNewWorkflow {
		expWorkflowID = *params.restartWorkflowID
	}

	store := &mockStore{
		newIDFunc: func() entity.ID {
			return expWorkflowID
		},
		createWorkflowFunc: func(ctx context.Context, w *entity.Workflow) error {
			isStoreCalled = true

			assert.False(t, w.CreatedAt.IsZero())
			assert.False(t, w.UpdatedAt.IsZero())

			expW := &entity.Workflow{
				ID:         expWorkflowID,
				CreatedAt:  w.CreatedAt,
				UpdatedAt:  w.UpdatedAt,
				SchemaName: schema.Name(),
				Status:     entity.WorkflowStatusInProgress,
				Input:      expPayload,
				ParentID:   params.expParentWorkflowID,
				RequestID:  converto.StringPointer(_reqID),
			}

			assert.Equal(t, expW, w)

			return nil
		},
	}

	broker := &mockQueueBroker{
		sendFunc: func(ctx context.Context, topic string, e event.BaseEvent) error {
			isBrokerCalled = true

			assert.Equal(t, params.expStep.Topic().String(), topic)
			assert.NotEmpty(t, e.GetID())

			expE := &event.WorkflowData{
				ID: e.GetID(),
				Workflow: event.Workflow{
					ID:          expWorkflowID.String(),
					Schema:      schema.Name().String(),
					Step:        params.expStep.Name().String(),
					StepPayload: expPayload,
				},
			}

			assert.Equal(t, expE, e)

			return nil
		},
	}

	o := workflow.NewOrchestrator(broker, store)
	err := o.AddWorkflowSchema(_bgCtx, schema)
	assert.NoError(t, err)

	actWorkflowID, err := params.orchestratorCaller(_bgCtxWithReqID, o, expPayload)
	assert.NoError(t, err)
	assert.Equal(t, expWorkflowID, actWorkflowID)
	assert.Equal(t, isNewWorkflow, isStoreCalled)
	assert.True(t, isBrokerCalled)
}

type mockStore struct {
	newIDFunc                func() entity.ID
	getWorkflowByIDFunc      func(ctx context.Context, id entity.ID) (*entity.Workflow, error)
	createWorkflowFunc       func(ctx context.Context, w *entity.Workflow) error
	setWorkflowStatusFunc    func(ctx context.Context, workflowID entity.ID, status entity.WorkflowStatus) error
	updateWorkflowForceFunc  func(ctx context.Context, workflowID entity.ID, params entity.UpdateWorkflowForceParams) error
	updateWorkflowNotNilFunc func(ctx context.Context, workflowID entity.ID, params entity.UpdateWorkflowNotNilParams) error
	putWorkflowStepsFunc     func(ctx context.Context, workflowID entity.ID, steps []*entity.WorkflowStep) error
}

func (m *mockStore) NewID() entity.ID {
	return m.newIDFunc()
}

func (m *mockStore) GetWorkflowByID(ctx context.Context, id entity.ID) (*entity.Workflow, error) {
	return m.getWorkflowByIDFunc(ctx, id)
}

func (m *mockStore) CreateWorkflow(ctx context.Context, w *entity.Workflow) error {
	return m.createWorkflowFunc(ctx, w)
}

func (m *mockStore) SetWorkflowStatus(ctx context.Context, workflowID entity.ID, status entity.WorkflowStatus) error {
	return m.setWorkflowStatusFunc(ctx, workflowID, status)
}

func (m *mockStore) UpdateWorkflowForce(ctx context.Context, workflowID entity.ID, params entity.UpdateWorkflowForceParams) error {
	return m.updateWorkflowForceFunc(ctx, workflowID, params)
}

func (m *mockStore) UpdateWorkflowNotNil(ctx context.Context, workflowID entity.ID, params entity.UpdateWorkflowNotNilParams) error {
	return m.updateWorkflowNotNilFunc(ctx, workflowID, params)
}

func (m *mockStore) PutWorkflowSteps(ctx context.Context, workflowID entity.ID, steps []*entity.WorkflowStep) error {
	return m.putWorkflowStepsFunc(ctx, workflowID, steps)
}

type mockQueueBroker struct {
	sendFunc func(ctx context.Context, topic string, e event.BaseEvent) error
}

func (m *mockQueueBroker) Send(ctx context.Context, topic string, e event.BaseEvent) error {
	return m.sendFunc(ctx, topic, e)
}

type stepWorkerTest struct {
	runCount   int
	lastResult json.RawMessage
}

func (s *stepWorkerTest) Run(_ context.Context, _ event.WorkflowEvent) (json.RawMessage, error) {
	s.runCount++
	s.lastResult = []byte(fmt.Sprintf(`{"cnt": %d}`, s.runCount))

	return s.lastResult, nil
}
