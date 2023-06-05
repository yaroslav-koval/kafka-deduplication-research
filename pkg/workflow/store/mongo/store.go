package mongo

import (
	"context"
	"fmt"
	"kafka-polygon/pkg/cerror"
	"kafka-polygon/pkg/converto"
	"kafka-polygon/pkg/workflow"
	"kafka-polygon/pkg/workflow/entity"
	"kafka-polygon/pkg/workflow/entrypoint/usecase"
	"kafka-polygon/pkg/workflow/store"
	"strings"
	"time"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

const (
	_defCollWorkflows       = "workflow"
	_defCollWorkflowHistory = "workflow_history"
)

type Store struct {
	dbName                      string
	collName, wfHistoryCollName string
	cl                          *mongo.Client
}

var _ workflow.Store = (*Store)(nil)
var _ usecase.Store = (*Store)(nil)

func NewStore(client *mongo.Client, dbName string) *Store {
	return &Store{
		cl:                client,
		dbName:            dbName,
		collName:          _defCollWorkflows,
		wfHistoryCollName: _defCollWorkflowHistory,
	}
}

func (s *Store) Type() string {
	return store.StoreTypeMongo
}

func (s *Store) NewID() entity.ID {
	return entity.ID(primitive.NewObjectID().Hex())
}

func (s *Store) SetCollectionName(collName string) {
	s.collName = collName
}

func (s *Store) GetWorkflowByID(ctx context.Context, id entity.ID) (*entity.Workflow, error) {
	w := &entity.Workflow{}

	if err := s.getCollection().FindOne(ctx, bson.M{"_id": id}).Decode(w); err != nil {
		if err == mongo.ErrNoDocuments {
			return nil, cerror.NewF(ctx,
				cerror.KindDBNoRows,
				"workflow with id %s does not exist", id).LogError()
		}

		return nil, cerror.NewF(ctx,
			cerror.DBToKind(err),
			"get workflow with id %s. err: %+v", id, err).LogError()
	}

	fmt.Printf("w: %+v", w)

	return w, nil
}

func (s *Store) SearchWorkflows(
	ctx context.Context, params entity.SearchWorkflowParams) (*entity.SearchWorkflowResult, error) {
	filters := make(map[string]interface{}, 0)

	if params.ID != nil {
		filters["_id"] = params.ID.String()
	}

	if params.Status != nil {
		filters["status"] = strings.ToUpper(params.Status.String())
	}

	p := entity.Paging{
		Limit:  store.DefLimit,
		Offset: store.DefOffset,
	}
	if params.Paging != nil {
		p.Offset = params.Paging.Offset
		p.Limit = params.Paging.Limit
	}

	ops := &options.FindOptions{
		Skip:  converto.Int64Pointer(int64(p.Offset)),
		Limit: converto.Int64Pointer(int64(p.Limit)),
		Sort:  map[string]int{"created_at": -1},
	}

	workflows := make([]*entity.Workflow, 0)

	cursor, err := s.getCollection().Find(ctx, filters, ops)
	if err != nil {
		return nil, cerror.New(ctx, cerror.DBToKind(err), err).LogError()
	}

	defer func() {
		_ = cursor.Close(ctx)
	}()

	if err := cursor.All(ctx, &workflows); err != nil {
		return nil, cerror.New(ctx, cerror.DBToKind(err), err).LogError()
	}

	return &entity.SearchWorkflowResult{
		Workflows: workflows,
		Paging:    p,
	}, nil
}

func (s *Store) CreateWorkflow(ctx context.Context, w *entity.Workflow) error {
	if _, err := s.getCollection().InsertOne(ctx, w); err != nil {
		return cerror.New(ctx, cerror.DBToKind(err), err).LogError()
	}

	return nil
}

func (s *Store) SetWorkflowStatus(ctx context.Context, workflowID entity.ID, status entity.WorkflowStatus) error {
	updateData := bson.M{
		"updated_at": time.Now().UTC(),
		"status":     status,
	}

	fields := bson.M{"$set": updateData}
	updateResults, err := s.getCollection().UpdateOne(ctx, bson.M{"_id": workflowID}, fields)

	if err != nil {
		return cerror.NewF(ctx,
			cerror.DBToKind(err),
			"update workflow with id: %s. err: %+v", workflowID, err).LogError()
	}

	if updateResults.MatchedCount <= 0 {
		return cerror.NewF(ctx,
			cerror.KindDBNoRows,
			"update workflow with id: %s. not found", workflowID).LogError()
	}

	return nil
}

// UpdateWorkflowForce updates certain task fields.
// All fields (even empty) from entity.ForceUpdateParams will be saved in db
func (s *Store) UpdateWorkflowForce(
	ctx context.Context, workflowID entity.ID, params entity.UpdateWorkflowForceParams) error {
	updateData := bson.M{
		"updated_at": time.Now().UTC(),
		"status":     params.Status,
		"error":      params.Error,
		"error_kind": params.ErrorKind,
	}

	fields := bson.M{"$set": updateData}
	updateResults, err := s.getCollection().UpdateOne(ctx, bson.M{"_id": workflowID}, fields)

	if err != nil {
		return cerror.NewF(ctx,
			cerror.DBToKind(err),
			"update workflow with id: %s. err: %+v", workflowID, err).LogError()
	}

	if updateResults.MatchedCount <= 0 {
		return cerror.NewF(ctx,
			cerror.KindDBNoRows,
			"update workflow with id: %s. not found", workflowID).LogError()
	}

	return nil
}

// UpdateWorkflowNotNil updates certain workflow fields.
// Only not nil fields will be saved in db.
func (s *Store) UpdateWorkflowNotNil(ctx context.Context, workflowID entity.ID, params entity.UpdateWorkflowNotNilParams) error {
	updateData := bson.M{}

	if params.Status != nil {
		updateData["status"] = *params.Status
	}

	if params.Error != nil {
		updateData["error"] = params.Error
	}

	if params.ErrorKind != nil {
		updateData["error_kind"] = params.ErrorKind
	}

	if params.Steps != nil {
		updateData["steps"] = params.Steps
	}

	if len(updateData) == 0 {
		return nil
	}

	updateData["updated_at"] = time.Now().UTC()

	fields := bson.M{"$set": updateData}
	updateResults, err := s.getCollection().UpdateOne(ctx, bson.M{"_id": workflowID}, fields)

	if err != nil {
		return cerror.NewF(ctx,
			cerror.DBToKind(err),
			"update workflow with id: %s. err: %+v", workflowID, err).LogError()
	}

	if updateResults.MatchedCount <= 0 {
		return cerror.NewF(ctx,
			cerror.KindDBNoRows,
			"update workflow with id: %s. not found", workflowID).LogError()
	}

	return nil
}

func (s *Store) PutWorkflowSteps(ctx context.Context, workflowID entity.ID, steps []*entity.WorkflowStep) error {
	fields := bson.M{"$set": bson.M{
		"steps":      steps,
		"updated_at": time.Now().UTC(),
	}}

	res, err := s.getCollection().UpdateOne(ctx, bson.M{"_id": workflowID}, fields)
	if err != nil {
		return cerror.NewF(ctx,
			cerror.DBToKind(err),
			"put steps for workflow with id: %s. err: %+v", workflowID, err).LogError()
	}

	if res.MatchedCount <= 0 {
		return cerror.NewF(ctx,
			cerror.KindDBNoRows,
			"put steps for workflow with id: %s. not found", workflowID).LogError()
	}

	return nil
}

func (s *Store) CreateWorkflowHistory(ctx context.Context, wh *entity.WorkflowHistory) error {
	if _, err := s.getCollectionHistory().InsertOne(ctx, wh); err != nil {
		return cerror.New(ctx, cerror.DBToKind(err), err).LogError()
	}

	return nil
}

func (s *Store) getCollection() *mongo.Collection {
	return s.cl.Database(s.dbName).Collection(s.collName)
}

func (s *Store) getCollectionHistory() *mongo.Collection {
	return s.cl.Database(s.dbName).Collection(s.wfHistoryCollName)
}
