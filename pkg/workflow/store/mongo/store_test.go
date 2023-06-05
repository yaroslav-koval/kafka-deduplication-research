package mongo_test

import (
	"context"
	"encoding/json"
	"fmt"
	"kafka-polygon/pkg/cerror"
	"kafka-polygon/pkg/converto"
	"kafka-polygon/pkg/workflow/entity"
	storeMongo "kafka-polygon/pkg/workflow/store/mongo"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
	"github.com/tj/assert"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo/integration/mtest"
)

var (
	mockClientTypeOpts = mtest.NewOptions().ClientType(mtest.Mock)
	bgCtx              = context.Background()
	dbName             = "test-db"
	collName           = "test-coll-name"
	now                = time.Now().UTC()

	modifiedResponse = mtest.CreateSuccessResponse(
		bson.E{Key: "n", Value: 2},
		bson.E{Key: "nModified", Value: 1},
		bson.E{Key: "upserted", Value: bson.A{
			bson.D{
				{Key: "index", Value: 0},
			},
		}},
	)

	noModifiedResponse = mtest.CreateSuccessResponse(
		bson.E{Key: "n", Value: 0},
		bson.E{Key: "nModified", Value: 0},
		bson.E{Key: "upserted", Value: bson.A{
			bson.D{
				{Key: "index", Value: 0},
			},
		}},
	)
)

func TestMongoStore(t *testing.T) {
	steps := make([]*entity.WorkflowStep, 0)
	stepData := map[string]interface{}{"case-key": "case-value"}
	stepDataBytes, _ := json.Marshal(stepData)

	steps = append(steps, &entity.WorkflowStep{
		Name:      "step-1",
		Data:      stepDataBytes,
		CreatedAt: now,
	})
	inputData := map[string]interface{}{"key": "val"}
	inputDataBytes, _ := json.Marshal(inputData)

	m := &entity.Workflow{
		ID:         entity.ID(primitive.NewObjectID().Hex()),
		Status:     entity.WorkflowStatusInProgress,
		SchemaName: "test-flow",
		Input:      inputDataBytes,
		Steps:      steps,
		Error:      entity.PointerWorkflowErrorMsg("test-err"),
		RequestID:  converto.StringPointer("test-request-id"),
		CreatedAt:  now,
		UpdatedAt:  now,
	}

	mt := mtest.New(t, mtest.NewOptions().ClientType(mtest.Mock))
	defer mt.Close()

	mt.RunOpts("workflow mongo store", mockClientTypeOpts, func(mt *mtest.T) {
		mt.Run("put task", func(mt *mtest.T) {
			mt.AddMockResponses(mtest.CreateSuccessResponse())
			s := storeMongo.NewStore(mt.Client, dbName)
			s.SetCollectionName(collName)
			err := s.CreateWorkflow(bgCtx, m)
			require.NoError(mt, err)
		})

		mt.Run("put task error", func(mt *mtest.T) {
			mt.AddMockResponses(mtest.CreateWriteConcernErrorResponse(mtest.WriteConcernError{
				Name:    "CreateWorkflowErr",
				Code:    1000,
				Message: "not create workflow",
			}))
			s := storeMongo.NewStore(mt.Client, dbName)
			s.SetCollectionName(collName)
			err := s.CreateWorkflow(bgCtx, m)
			require.Error(mt, err)
			assert.Equal(t, cerror.KindDBOther, cerror.ErrKind(err))
		})

		mt.Run("put case", func(mt *mtest.T) {
			mt.AddMockResponses(mtest.CreateSuccessResponse(), modifiedResponse)
			s := storeMongo.NewStore(mt.Client, dbName)
			s.SetCollectionName(collName)
			err := s.CreateWorkflow(bgCtx, m)
			require.NoError(mt, err)

			err = s.PutWorkflowSteps(bgCtx, m.ID, steps)
			require.NoError(mt, err)
		})

		mt.Run("put case error", func(mt *mtest.T) {
			mt.AddMockResponses(noModifiedResponse)
			s := storeMongo.NewStore(mt.Client, dbName)
			s.SetCollectionName(collName)
			err := s.PutWorkflowSteps(bgCtx, m.ID, steps)
			require.Error(mt, err)
			assert.Equal(t, fmt.Sprintf("put steps for workflow with id: %s. not found", m.ID), err.Error())
			assert.Equal(t, cerror.KindDBNoRows, cerror.ErrKind(err))
		})

		mt.Run("put case db error", func(mt *mtest.T) {
			mt.AddMockResponses(mtest.CreateWriteConcernErrorResponse(mtest.WriteConcernError{
				Name:    "PutWorkflowStepsErr",
				Code:    1000,
				Message: "not put case",
			}))
			s := storeMongo.NewStore(mt.Client, dbName)
			s.SetCollectionName(collName)
			err := s.PutWorkflowSteps(bgCtx, m.ID, steps)
			require.Error(mt, err)
			assert.Equal(t,
				fmt.Sprintf(
					"put steps for workflow with id: %s. "+
						"err: write exception: write concern error: (PutWorkflowStepsErr) not put case",
					m.ID),
				err.Error())
			assert.Equal(t, cerror.KindDBOther, cerror.ErrKind(err))
		})

		mt.Run("task status", func(mt *mtest.T) {
			mt.AddMockResponses(modifiedResponse)
			s := storeMongo.NewStore(mt.Client, dbName)
			s.SetCollectionName(collName)
			err := s.SetWorkflowStatus(bgCtx, m.ID, entity.WorkflowStatusSuccess)
			require.NoError(mt, err)
		})

		mt.Run("task status error", func(mt *mtest.T) {
			mt.AddMockResponses(noModifiedResponse)
			s := storeMongo.NewStore(mt.Client, dbName)
			s.SetCollectionName(collName)
			err := s.SetWorkflowStatus(bgCtx, m.ID, entity.WorkflowStatusSuccess)
			require.Error(mt, err)
			assert.Equal(t, fmt.Sprintf("update workflow with id: %s. not found", m.ID), err.Error())
			assert.Equal(t, cerror.KindDBNoRows, cerror.ErrKind(err))
		})

		mt.Run("task status db error", func(mt *mtest.T) {
			mt.AddMockResponses(mtest.CreateWriteConcernErrorResponse(mtest.WriteConcernError{
				Name:    "SetWorkflowStatusErr",
				Code:    1000,
				Message: "not task status",
			}))
			s := storeMongo.NewStore(mt.Client, dbName)
			s.SetCollectionName(collName)
			err := s.SetWorkflowStatus(bgCtx, m.ID, entity.WorkflowStatusSuccess)
			require.Error(mt, err)
			assert.Equal(t, fmt.Sprintf("update workflow with id: %s. "+
				"err: write exception: write concern error: (SetWorkflowStatusErr) not task status",
				m.ID), err.Error())
			assert.Equal(t, cerror.KindDBOther, cerror.ErrKind(err))
		})

		mt.Run("update task", func(mt *mtest.T) {
			mt.AddMockResponses(modifiedResponse)
			s := storeMongo.NewStore(mt.Client, dbName)
			s.SetCollectionName(collName)
			err := s.UpdateWorkflowForce(bgCtx, m.ID, entity.UpdateWorkflowForceParams{
				Status: entity.WorkflowStatusSuccess,
			})
			require.NoError(mt, err)
		})

		mt.Run("update task error", func(mt *mtest.T) {
			mt.AddMockResponses(noModifiedResponse)
			s := storeMongo.NewStore(mt.Client, dbName)
			s.SetCollectionName(collName)
			err := s.UpdateWorkflowForce(bgCtx, m.ID, entity.UpdateWorkflowForceParams{
				Status: entity.WorkflowStatusSuccess,
			})
			require.Error(mt, err)
			assert.Equal(t, fmt.Sprintf("update workflow with id: %s. not found", m.ID), err.Error())
			assert.Equal(t, cerror.KindDBNoRows, cerror.ErrKind(err))
		})

		mt.Run("update task db error", func(mt *mtest.T) {
			mt.AddMockResponses(mtest.CreateWriteConcernErrorResponse(mtest.WriteConcernError{
				Name:    "UpdateWorkflowForceErr",
				Code:    1000,
				Message: "not update task",
			}))
			s := storeMongo.NewStore(mt.Client, dbName)
			s.SetCollectionName(collName)
			err := s.UpdateWorkflowForce(bgCtx, m.ID, entity.UpdateWorkflowForceParams{
				Status: entity.WorkflowStatusSuccess,
			})
			require.Error(mt, err)
			assert.Equal(t, fmt.Sprintf("update workflow with id: %s. "+
				"err: write exception: write concern error: (UpdateWorkflowForceErr) not update task",
				m.ID), err.Error())
			assert.Equal(t, cerror.KindDBOther, cerror.ErrKind(err))
		})

		mt.Run("update task not nil fields", func(mt *mtest.T) {
			mt.AddMockResponses(modifiedResponse)
			s := storeMongo.NewStore(mt.Client, dbName)
			s.SetCollectionName(collName)
			err := s.UpdateWorkflowNotNil(bgCtx, m.ID, entity.UpdateWorkflowNotNilParams{
				Status: entity.PointerWorkflowStatus(entity.WorkflowStatusSuccess.String()),
			})
			require.NoError(mt, err)
		})

		mt.Run("update task not nil fields error", func(mt *mtest.T) {
			mt.AddMockResponses(noModifiedResponse)
			s := storeMongo.NewStore(mt.Client, dbName)
			s.SetCollectionName(collName)
			err := s.UpdateWorkflowNotNil(bgCtx, m.ID, entity.UpdateWorkflowNotNilParams{
				Status: entity.PointerWorkflowStatus(entity.WorkflowStatusSuccess.String()),
			})
			require.Error(mt, err)
			assert.Equal(t, fmt.Sprintf("update workflow with id: %s. not found", m.ID), err.Error())
			assert.Equal(t, cerror.KindDBNoRows, cerror.ErrKind(err))
		})

		mt.Run("task by id", func(mt *mtest.T) {
			rows := []bson.D{{{Key: "_id", Value: m.ID}}}
			findRes := mtest.CreateCursorResponse(1, "test-db.test-coll-name", mtest.FirstBatch, rows...)
			mt.AddMockResponses(findRes)
			s := storeMongo.NewStore(mt.Client, dbName)
			s.SetCollectionName(collName)
			nTask, err := s.GetWorkflowByID(bgCtx, m.ID)
			require.NoError(mt, err)
			assert.NotNil(t, nTask)
		})

		mt.Run("task by id error", func(mt *mtest.T) {
			mt.AddMockResponses(mtest.CreateWriteConcernErrorResponse(mtest.WriteConcernError{
				Name:    "CreateWorkflowErr",
				Code:    1000,
				Message: "task by id",
			}))

			s := storeMongo.NewStore(mt.Client, dbName)
			s.SetCollectionName(collName)
			nTask, err := s.GetWorkflowByID(bgCtx, m.ID)
			require.Error(mt, err)
			assert.Nil(t, nTask)
			assert.Equal(t, cerror.KindDBOther, cerror.ErrKind(err))
		})

		mt.Run("search tasks", func(mt *mtest.T) {
			ns := "test-db.test-coll-name"
			mFail := &entity.Workflow{
				ID:         entity.ID(primitive.NewObjectID().Hex()),
				Status:     entity.WorkflowStatusFailed,
				SchemaName: "test-flow",
				Input:      inputDataBytes,
				Error:      entity.PointerWorkflowErrorMsg("test-err"),
				RequestID:  converto.StringPointer("test-request-id"),
				Steps:      steps,
				CreatedAt:  now,
				UpdatedAt:  now,
			}

			bsonOk := bson.D{
				{Key: "ok", Value: 1},
				{Key: "_id", Value: m.ID},
			}

			bsonFail := bson.D{
				{Key: "ok", Value: 1},
				{Key: "_id", Value: mFail.ID},
			}

			find := mtest.CreateCursorResponse(1, ns, mtest.FirstBatch, bsonOk)
			getMore := mtest.CreateCursorResponse(1, ns, mtest.NextBatch, bsonFail)
			killCursors := mtest.CreateCursorResponse(0, ns, mtest.NextBatch)
			mt.AddMockResponses(find, getMore, killCursors)

			s := storeMongo.NewStore(mt.Client, dbName)
			s.SetCollectionName(collName)

			res, err := s.SearchWorkflows(bgCtx, entity.SearchWorkflowParams{})
			assert.Nil(mt, err, "Find all error: %v", err)
			assert.True(mt, len(res.Workflows) > 0)
		})

		mt.Run("save history", func(mt *mtest.T) {
			mt.AddMockResponses(mtest.CreateSuccessResponse(), modifiedResponse)
			s := storeMongo.NewStore(mt.Client, dbName)
			s.SetCollectionName(collName)
			err := s.CreateWorkflow(bgCtx, m)
			require.NoError(mt, err)

			mH := &entity.WorkflowHistory{
				Type:              entity.WorkflowHistoryTypeRestart,
				Input:             inputDataBytes,
				InputPrevious:     inputDataBytes,
				StepName:          "test-step-name",
				WorkflowID:        m.ID,
				RequestID:         converto.StringPointer("test-request-id"),
				WorkflowStatus:    "test-wf-status",
				WorkflowError:     entity.PointerWorkflowErrorMsg("test-wf-err"),
				WorkflowErrorKind: entity.PointerWorkflowErrorKind("test-wf-err-kind"),
				CreatedAt:         now,
			}

			err = s.CreateWorkflowHistory(bgCtx, mH)
			require.NoError(mt, err)
		})

		mt.Run("save history error", func(mt *mtest.T) {
			mt.AddMockResponses(mtest.CreateWriteConcernErrorResponse(mtest.WriteConcernError{
				Name:    "CreateWorkflowHistoryErr",
				Code:    1000,
				Message: "not save history",
			}))

			s := storeMongo.NewStore(mt.Client, dbName)
			s.SetCollectionName(collName)
			mH := &entity.WorkflowHistory{
				Type:              entity.WorkflowHistoryTypeRestart,
				Input:             inputDataBytes,
				InputPrevious:     inputDataBytes,
				StepName:          "test-step-name",
				WorkflowID:        m.ID,
				RequestID:         converto.StringPointer("test-request-id"),
				WorkflowStatus:    "test-wf-status",
				WorkflowError:     entity.PointerWorkflowErrorMsg("test-wf-err"),
				WorkflowErrorKind: entity.PointerWorkflowErrorKind("test-wf-err-kind"),
				CreatedAt:         now,
			}

			err := s.CreateWorkflowHistory(bgCtx, mH)
			require.Error(mt, err)
			assert.Equal(t, cerror.KindDBOther, cerror.ErrKind(err))
		})
	})
}
