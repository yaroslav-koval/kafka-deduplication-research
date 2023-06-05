package postgres

import (
	"context"
	"database/sql"
	"errors"
	"kafka-polygon/pkg/cerror"
	"kafka-polygon/pkg/workflow"
	"kafka-polygon/pkg/workflow/entity"
	"kafka-polygon/pkg/workflow/entrypoint/usecase"
	"kafka-polygon/pkg/workflow/store"
	"strings"
	"time"

	uuid "github.com/satori/go.uuid"
	"github.com/uptrace/bun"
)

type Store struct {
	db *bun.DB
}

var _ workflow.Store = (*Store)(nil)
var _ usecase.Store = (*Store)(nil)

func NewStore(db *bun.DB) *Store {
	return &Store{db: db}
}

func (s *Store) Type() string {
	return store.StoreTypePostgres
}

func (s *Store) NewID() entity.ID {
	return entity.ID(uuid.NewV4().String())
}

func (s *Store) GetWorkflowByID(ctx context.Context, id entity.ID) (*entity.Workflow, error) {
	dst := &entity.Workflow{ID: id}
	if err := s.db.NewSelect().Model(dst).WherePK().Scan(ctx); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, cerror.NewF(ctx,
				cerror.KindDBNoRows,
				"workflow with id %s does not exist", id).LogError()
		}

		return nil, cerror.NewF(ctx,
			cerror.DBToKind(err),
			"get workflow with id %s. err: %+v", id, err).LogError()
	}

	return dst, nil
}

func (s *Store) SearchWorkflows(ctx context.Context, params entity.SearchWorkflowParams) (*entity.SearchWorkflowResult, error) {
	dst := make([]*entity.Workflow, 0)
	q := s.db.NewSelect().Model(&dst)

	if params.ID != nil {
		q.Where("id=?", params.ID.String())
	}

	if params.Status != nil {
		q.Where("status=?", strings.ToUpper(params.Status.String()))
	}

	p := entity.Paging{
		Limit:  store.DefLimit,
		Offset: store.DefOffset,
	}
	if params.Paging != nil {
		p.Offset = params.Paging.Offset
		p.Limit = params.Paging.Limit
	}

	err := q.Offset(p.Offset).Limit(p.Limit).Order("created_at").Scan(ctx)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, cerror.New(ctx, cerror.KindDBNoRows, err).LogError()
		}

		return nil, cerror.New(ctx, cerror.DBToKind(err), err).LogError()
	}

	return &entity.SearchWorkflowResult{
		Workflows: dst,
		Paging:    p,
	}, nil
}

func (s *Store) CreateWorkflow(ctx context.Context, w *entity.Workflow) error {
	if _, err := s.db.NewInsert().Model(w).Exec(ctx); err != nil {
		return cerror.New(ctx, cerror.DBToKind(err), err).LogError()
	}

	return nil
}

func (s *Store) SetWorkflowStatus(ctx context.Context, workflowID entity.ID, status entity.WorkflowStatus) error {
	if _, err := s.db.
		NewUpdate().
		Model(&entity.Workflow{
			ID:        workflowID,
			Status:    status,
			UpdatedAt: time.Now().UTC(),
		}).
		Column("status", "updated_at").
		WherePK().
		Exec(ctx); err != nil {
		return cerror.NewF(ctx,
			cerror.DBToKind(err),
			"set status for workflow with id: %s. err: %+v", workflowID, err).LogError()
	}

	return nil
}

// UpdateWorkflowForce updates certain workflow fields.
// All fields (even empty) from entity.UpdateWorkflowForceParams will be saved in db.
func (s *Store) UpdateWorkflowForce(ctx context.Context, workflowID entity.ID, params entity.UpdateWorkflowForceParams) error {
	if _, err := s.db.
		NewUpdate().
		Model(&entity.Workflow{
			ID:        workflowID,
			Status:    params.Status,
			Error:     params.Error,
			ErrorKind: params.ErrorKind,
			UpdatedAt: time.Now().UTC(),
		}).
		Column("status", "error", "error_kind", "updated_at").
		WherePK().
		Exec(ctx); err != nil {
		return cerror.NewF(ctx,
			cerror.DBToKind(err),
			"update workflow with id: %s. err: %+v", workflowID, err).LogError()
	}

	return nil
}

// UpdateWorkflowNotNil updates certain workflow fields.
// Only not nil fields will be saved in db.
func (s *Store) UpdateWorkflowNotNil(ctx context.Context, workflowID entity.ID, params entity.UpdateWorkflowNotNilParams) error {
	updateModel := &entity.Workflow{
		ID:        workflowID,
		Steps:     params.Steps,
		Error:     params.Error,
		ErrorKind: params.ErrorKind,
		UpdatedAt: time.Now().UTC(),
	}

	if params.Status != nil {
		updateModel.Status = *params.Status
	}

	if _, err := s.db.
		NewUpdate().
		Model(updateModel).
		Column("status", "steps", "error", "error_kind", "updated_at").
		OmitZero().
		WherePK().
		Exec(ctx); err != nil {
		return cerror.NewF(ctx,
			cerror.DBToKind(err),
			"update workflow with id: %s. err: %+v", workflowID, err).LogError()
	}

	return nil
}

func (s *Store) PutWorkflowSteps(ctx context.Context, workflowID entity.ID, steps []*entity.WorkflowStep) error {
	if _, err := s.db.
		NewUpdate().
		Model(&entity.Workflow{
			ID:        workflowID,
			Steps:     steps,
			UpdatedAt: time.Now().UTC(),
		}).
		Column("steps", "updated_at").
		WherePK().
		Exec(ctx); err != nil {
		return cerror.NewF(ctx,
			cerror.DBToKind(err),
			"put steps for workflow with id: %s. err: %+v", workflowID, err).LogError()
	}

	return nil
}

func (s *Store) CreateWorkflowHistory(ctx context.Context, wh *entity.WorkflowHistory) error {
	if _, err := s.db.NewInsert().Model(wh).Exec(ctx); err != nil {
		return cerror.New(ctx, cerror.DBToKind(err), err).LogError()
	}

	return nil
}
