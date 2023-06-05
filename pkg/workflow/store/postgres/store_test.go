package postgres_test

import (
	"context"
	"kafka-polygon/pkg/cerror"
	"kafka-polygon/pkg/converto"
	"kafka-polygon/pkg/db/postgres/bun"
	"kafka-polygon/pkg/env"
	"kafka-polygon/pkg/log"
	"kafka-polygon/pkg/testutil"
	"kafka-polygon/pkg/workflow/entity"
	"kafka-polygon/pkg/workflow/store/postgres"
	"testing"
	"time"

	"github.com/jmoiron/sqlx"
	uuid "github.com/satori/go.uuid"
	"github.com/stretchr/testify/suite"
)

var (
	testTimeFormat = "2006-06-06 01:02:03"
	bgCtx          = context.Background()
)

type storeTestSuite struct {
	suite.Suite
	pgCont *testutil.DockerPGContainer
	db     *sqlx.DB
	pgConn *bun.Connection
	pgCfg  *env.Postgres
	st     *postgres.Store
}

func TestStoreTestSuite(t *testing.T) {
	log.SetGlobalLogLevel("fatal")
	suite.Run(t, new(storeTestSuite))
}

func (s *storeTestSuite) SetupSuite() {
	ctx := context.Background()
	dbName := "workflow_test"
	pgCont := testutil.NewDockerUtilInstance().
		InitPG().
		CreatePostgresContainerDatabase(dbName).
		ConnectPostgresDB(dbName)
	s.pgCont = pgCont
	s.db = pgCont.GetDBInfoByName(dbName).DBClient

	postgresCfg := pgCont.GetPostgresEnvConfig()
	postgresCfg.DBName = dbName
	s.pgCfg = &postgresCfg

	s.pgConn = bun.NewConnection(&postgresCfg)
	err := s.pgConn.Connect()

	if err != nil {
		_ = cerror.New(ctx, cerror.KindInternal, err).LogError()
		return
	}

	query := `CREATE TABLE IF NOT EXISTS workflow (
    id UUID PRIMARY KEY,
    parent_id UUID NULL references workflow(id),
    status TEXT NOT NULL,
    schema_name TEXT NOT NULL DEFAULT '',
    error TEXT NULL,
    error_kind varchar(50) NULL,
    request_id varchar(50) NULL,
    input JSONB NOT NULL DEFAULT '{}',
    steps JSONB,
	created_at timestamp NOT NULL DEFAULT (NOW() AT TIME ZONE 'UTC'),
	updated_at timestamp NOT NULL DEFAULT (NOW() AT TIME ZONE 'UTC')
);`

	_, err = s.pgConn.DB().ExecContext(ctx, query)
	if err != nil {
		s.T().Error(err)
		return
	}

	queryHistory := `CREATE TABLE IF NOT EXISTS workflow_history (
    id UUID PRIMARY KEY,
    type varchar(50) NOT NULL,
    input JSONB NOT NULL DEFAULT '{}',
    input_previous JSONB NOT NULL DEFAULT '{}',
    step_name varchar(50) NOT NULL,
    workflow_id UUID NOT NULL references workflow(id),
    request_id varchar(50) NULL,
    workflow_status varchar(50) NOT NULL,
    workflow_error TEXT NULL,
    workflow_error_kind varchar(50) NULL,
    created_at timestamp NOT NULL DEFAULT (NOW() AT TIME ZONE 'UTC')
);`

	_, err = s.pgConn.DB().ExecContext(ctx, queryHistory)
	if err != nil {
		s.T().Error(err)
		return
	}

	s.st = postgres.NewStore(s.pgConn.DB())
}

func (s *storeTestSuite) TearDownTest() {
}

func (s *storeTestSuite) TearDownSuite() {
	_ = s.pgConn.Close()
	_ = s.db.Close()
	_ = s.pgCont.CloseDBConnectionByName(s.pgCfg.DBName)
}

func (s *storeTestSuite) TestCreateWorkflow() {
	uID := entity.ID(uuid.NewV4().String())
	now := time.Now().UTC()

	m := &entity.Workflow{
		ID:     uID,
		Status: "test-status",
		Steps: []*entity.WorkflowStep{
			{
				Name:      "test-case-name",
				Data:      []byte("{}"),
				CreatedAt: now,
			},
		},
		Error:     entity.PointerWorkflowErrorMsg("test-error"),
		RequestID: converto.StringPointer("test-request-id"),
		CreatedAt: now,
		UpdatedAt: now,
	}

	err := s.st.CreateWorkflow(bgCtx, m)
	s.NoError(err)

	err = s.st.CreateWorkflow(bgCtx, m)
	s.Error(err)
	s.Equal(cerror.KindDBOther, cerror.ErrKind(err))
}

func (s *storeTestSuite) TestGetWorkflowByID() {
	uID := entity.ID(uuid.NewV4().String())
	now := time.Now().UTC()
	m := &entity.Workflow{
		ID:         uID,
		Status:     "test-status",
		SchemaName: "test-flow-type",
		Steps: []*entity.WorkflowStep{
			{
				Name:      "test-case-name",
				Data:      []byte("{}"),
				CreatedAt: now,
			},
		},
		Error:     entity.PointerWorkflowErrorMsg("test-error"),
		RequestID: converto.StringPointer("test-request-id"),
		CreatedAt: now,
		UpdatedAt: now,
	}

	err := s.st.CreateWorkflow(bgCtx, m)
	s.NoError(err)

	data, err := s.st.GetWorkflowByID(bgCtx, uID)
	s.NoError(err)
	s.Equal(m.ID, data.ID)
	s.Equal(m.CreatedAt.Format(testTimeFormat), data.CreatedAt.Format(testTimeFormat))
	s.Equal(m.UpdatedAt.Format(testTimeFormat), data.UpdatedAt.Format(testTimeFormat))
	s.Equal(m.Status, data.Status)
	s.Equal(m.Error, data.Error)
	s.Equal(m.RequestID, data.RequestID)
	s.Equal(m.Steps, data.Steps)

	newID := entity.ID(uuid.NewV4().String())
	data, err = s.st.GetWorkflowByID(bgCtx, newID)
	s.Error(err)
	s.Nil(data)
	s.Equal(cerror.KindDBNoRows, cerror.ErrKind(err))
}

func (s *storeTestSuite) TestWorkflowStatus() {
	uID := entity.ID(uuid.NewV4().String())
	now := time.Now().UTC()
	m := &entity.Workflow{
		ID:     uID,
		Status: entity.WorkflowStatusInProgress,
		Steps: []*entity.WorkflowStep{
			{
				Name:      "test-case-name",
				Data:      []byte("{}"),
				CreatedAt: now,
			},
		},
		CreatedAt: now,
		UpdatedAt: now,
	}

	err := s.st.CreateWorkflow(bgCtx, m)
	s.NoError(err)

	err = s.st.SetWorkflowStatus(bgCtx, uID, entity.WorkflowStatusSuccess)
	s.NoError(err)

	data, err := s.st.GetWorkflowByID(bgCtx, uID)
	s.NoError(err)
	s.Equal(m.ID, data.ID)
	s.Equal(entity.WorkflowStatusSuccess, data.Status)
}

func (s *storeTestSuite) TestUpdateWorkflowForce() {
	uID := entity.ID(uuid.NewV4().String())
	now := time.Now().UTC()
	expErr := converto.StringPointer("test error")
	m := &entity.Workflow{
		ID:     uID,
		Status: entity.WorkflowStatusInProgress,
		Steps: []*entity.WorkflowStep{
			{
				Name:      "test-case-name",
				Data:      []byte("{}"),
				CreatedAt: now,
			},
		},
		CreatedAt: now,
		UpdatedAt: now,
	}

	err := s.st.CreateWorkflow(bgCtx, m)
	s.NoError(err)

	err = s.st.UpdateWorkflowForce(bgCtx, uID, entity.UpdateWorkflowForceParams{
		Status: entity.WorkflowStatusSuccess,
		Error:  entity.PointerWorkflowErrorMsg(*expErr)})
	s.NoError(err)

	data, err := s.st.GetWorkflowByID(bgCtx, uID)
	s.NoError(err)
	s.Equal(m.ID, data.ID)
	s.Equal(entity.WorkflowStatusSuccess, data.Status)
	s.Equal(*expErr, data.Error.String())
}

func (s *storeTestSuite) TestUpdateWorkflowNotNil() {
	uID := entity.ID(uuid.NewV4().String())
	now := time.Now().UTC()
	expErr := converto.StringPointer("test error")
	expErrKind := converto.StringPointer("test error kind")
	expSteps := []*entity.WorkflowStep{
		{
			Name:      "test-case-name-updated",
			Data:      []byte("{}"),
			CreatedAt: now,
		},
	}

	m := &entity.Workflow{
		ID:     uID,
		Status: entity.WorkflowStatusInProgress,
		Steps: []*entity.WorkflowStep{
			{
				Name:      "test-case-name",
				Data:      []byte("{}"),
				CreatedAt: now,
			},
		},
		CreatedAt: now,
		UpdatedAt: now,
	}

	err := s.st.CreateWorkflow(bgCtx, m)
	s.NoError(err)

	err = s.st.UpdateWorkflowNotNil(bgCtx, uID, entity.UpdateWorkflowNotNilParams{
		Status:    entity.PointerWorkflowStatus(entity.WorkflowStatusSuccess.String()),
		Error:     entity.PointerWorkflowErrorMsg(*expErr),
		ErrorKind: entity.PointerWorkflowErrorKind(*expErrKind),
		Steps:     expSteps,
	})
	s.NoError(err)

	data, err := s.st.GetWorkflowByID(bgCtx, uID)
	s.NoError(err)
	s.Equal(m.ID, data.ID)
	s.Equal(entity.WorkflowStatusSuccess, data.Status)
	s.Equal(*expErr, data.Error.String())
	s.Equal(*expErrKind, data.ErrorKind.String())
	s.Equal(expSteps, data.Steps)

	err = s.st.UpdateWorkflowNotNil(bgCtx, uID, entity.UpdateWorkflowNotNilParams{
		Status: entity.PointerWorkflowStatus(entity.WorkflowStatusFailed.String()),
	})
	s.NoError(err)

	data, err = s.st.GetWorkflowByID(bgCtx, uID)
	s.NoError(err)
	s.Equal(entity.WorkflowStatusFailed, data.Status)
	s.NotNil(data.Error)
	s.NotNil(data.ErrorKind)
	s.NotNil(data.Steps)
}

func (s *storeTestSuite) TestPutStep() {
	uID := entity.ID(uuid.NewV4().String())
	now := time.Now().UTC()
	m := &entity.Workflow{
		ID:         uID,
		Status:     entity.WorkflowStatusInProgress,
		SchemaName: "test-flow-type",
		Steps: []*entity.WorkflowStep{
			{
				Name:      "test-case-name",
				Data:      []byte("{}"),
				CreatedAt: now,
			},
		},
		CreatedAt: now,
		UpdatedAt: now,
	}

	err := s.st.CreateWorkflow(bgCtx, m)
	s.NoError(err)

	nStep := []*entity.WorkflowStep{
		{
			Name:      "test-case-name-new",
			CreatedAt: now,
		},
	}
	err = s.st.PutWorkflowSteps(bgCtx, uID, nStep)
	s.NoError(err)

	data, err := s.st.GetWorkflowByID(bgCtx, uID)
	s.NoError(err)
	s.Equal(m.ID, data.ID)
	s.Equal(entity.WorkflowStatusInProgress, data.Status)
	s.Equal(len(nStep), len(data.Steps))
}

func (s *storeTestSuite) TestSearchWorkflows() {
	uID := entity.ID(uuid.NewV4().String())
	now := time.Now().UTC()
	m := &entity.Workflow{
		ID:         uID,
		Status:     entity.WorkflowStatusFailed,
		SchemaName: "test-flow-type",
		Steps: []*entity.WorkflowStep{
			{
				Name:      "test-case-name",
				Data:      []byte("{}"),
				CreatedAt: now,
			},
		},
		Error:     entity.PointerWorkflowErrorMsg("test-error"),
		RequestID: converto.StringPointer("test-request-id"),
		CreatedAt: now,
		UpdatedAt: now,
	}

	err := s.st.CreateWorkflow(bgCtx, m)
	s.NoError(err)

	data, err := s.st.SearchWorkflows(bgCtx, entity.SearchWorkflowParams{
		ID:     entity.PointerID(uID.String()),
		Status: entity.PointerWorkflowStatus(entity.WorkflowStatusFailed.String()),
	})
	s.NoError(err)
	s.True(len(data.Workflows) > 0)

	for _, d := range data.Workflows {
		s.Equal(m.ID, d.ID)
		s.Equal(m.CreatedAt.Format(testTimeFormat), d.CreatedAt.Format(testTimeFormat))
		s.Equal(m.UpdatedAt.Format(testTimeFormat), d.UpdatedAt.Format(testTimeFormat))
		s.Equal(m.Status, d.Status)
		s.Equal(m.Error, d.Error)
		s.Equal(m.RequestID, d.RequestID)
		s.Equal(m.Steps, d.Steps)
	}
}

func (s *storeTestSuite) TestSaveHistory() {
	uID := entity.ID(uuid.NewV4().String())
	now := time.Now().UTC()

	mT := &entity.Workflow{
		ID:         uID,
		Status:     entity.WorkflowStatusFailed,
		SchemaName: "test-flow-type",
		Steps: []*entity.WorkflowStep{
			{
				Name:      "test-case-name",
				Data:      []byte("{}"),
				CreatedAt: now,
			},
		},
		Error:     entity.PointerWorkflowErrorMsg("test-error"),
		RequestID: converto.StringPointer("test-request-id"),
		CreatedAt: now,
		UpdatedAt: now,
	}

	err := s.st.CreateWorkflow(bgCtx, mT)
	s.NoError(err)

	mH := &entity.WorkflowHistory{
		ID:                uID,
		Type:              entity.WorkflowHistoryTypeRestart,
		StepName:          "test-step-name",
		WorkflowID:        uID,
		RequestID:         converto.StringPointer("test-request-id"),
		WorkflowStatus:    "test-wf-status",
		WorkflowError:     entity.PointerWorkflowErrorMsg("test-wf-err"),
		WorkflowErrorKind: entity.PointerWorkflowErrorKind("test-wf-err-kind"),
		CreatedAt:         now,
	}

	err = s.st.CreateWorkflowHistory(bgCtx, mH)
	s.NoError(err)

	err = s.st.CreateWorkflowHistory(bgCtx, mH)
	s.Error(err)
	s.Equal(cerror.KindDBOther, cerror.ErrKind(err))
}
