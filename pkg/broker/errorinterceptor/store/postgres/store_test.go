package postgres_test

import (
	"context"
	"fmt"
	"kafka-polygon/pkg/broker/errorinterceptor/entity"
	"kafka-polygon/pkg/broker/errorinterceptor/migration/schema"
	pgStore "kafka-polygon/pkg/broker/errorinterceptor/store/postgres"
	"kafka-polygon/pkg/cerror"
	"kafka-polygon/pkg/db/postgres"
	"kafka-polygon/pkg/db/postgres/bun"
	"kafka-polygon/pkg/env"
	"kafka-polygon/pkg/log"
	"kafka-polygon/pkg/testutil"
	"testing"

	bindata "github.com/golang-migrate/migrate/v4/source/go_bindata"
	"github.com/jmoiron/sqlx"
	"github.com/stretchr/testify/suite"
)

const (
	migrationVersion uint = 1
)

var _bgCtx = context.Background()

type repoTestSuite struct {
	suite.Suite
	pgCont *testutil.DockerPGContainer
	db     *sqlx.DB
	repo   *pgStore.Store
	pgConn *bun.Connection
	pgCfg  *env.Postgres
}

func TestRepoTestSuite(t *testing.T) {
	log.SetGlobalLogLevel("fatal")
	suite.Run(t, new(repoTestSuite))
}

func (ts *repoTestSuite) SetupSuite() {
	dbName := "hcppr_test"
	pgCont := testutil.NewDockerUtilInstance().
		InitPG().
		CreatePostgresContainerDatabase(dbName).
		ConnectPostgresDB(dbName).
		RunDBMigrations(_bgCtx, dbName, testutil.DockerMigrateConfig{
			Name:     postgres.MigrateName,
			Version:  migrationVersion,
			Resource: bindata.Resource(schema.AssetNames(), schema.Asset),
		})
	ts.pgCont = pgCont
	ts.db = pgCont.GetDBInfoByName(dbName).DBClient

	postgresCfg := pgCont.GetPostgresEnvConfig()
	postgresCfg.DBName = dbName
	ts.pgCfg = &postgresCfg

	ts.pgConn = bun.NewConnection(&postgresCfg)
	err := ts.pgConn.Connect()

	if err != nil {
		_ = cerror.New(_bgCtx, cerror.KindInternal, err).LogError()
		return
	}

	ts.repo = pgStore.NewStore(ts.pgConn.DB())
}

func (ts *repoTestSuite) TearDownTest() {
	ts.deleteAllRecordsByTableName("failed_broker_event")
}

func (ts *repoTestSuite) TearDownSuite() {
	_ = ts.pgConn.Close()
	_ = ts.db.Close()
	_ = ts.pgCont.CloseDBConnectionByName(ts.pgCfg.DBName)
}

func (ts *repoTestSuite) TestSaveEvent() {
	e := &entity.FailedBrokerEvent{
		ID:    "id",
		Topic: "topic",
		Data:  []byte("{}"),
		Error: "error",
	}
	actualEv, err := ts.repo.SaveEvent(_bgCtx, e)
	ts.NoError(err)
	ts.Equal(e.ID, actualEv.ID)
	ts.Equal(e.Topic, actualEv.Topic)
	ts.Equal(e.Data, actualEv.Data)
	ts.Equal(e.Error, actualEv.Error)

	e.Error = "other error"
	actualEv, err = ts.repo.SaveEvent(_bgCtx, e)
	ts.NoError(err)
	ts.Equal(e.ID, actualEv.ID)
	ts.Equal(e.Topic, actualEv.Topic)
	ts.Equal(e.Data, actualEv.Data)
	ts.Equal(e.Error, actualEv.Error)
}

func (ts *repoTestSuite) TestSearchEvent() {
	e := &entity.FailedBrokerEvent{
		ID:    "id",
		Topic: "topic",
		Data:  []byte("{}"),
		Error: "error",
	}
	_, err := ts.repo.SaveEvent(_bgCtx, e)
	ts.NoError(err)

	events, err := ts.repo.SearchEvents(_bgCtx)
	ts.NoError(err)
	ts.Equal(1, len(events))
}

func (ts *repoTestSuite) deleteAllRecordsByTableName(tableName string) {
	if _, err := ts.db.Exec(fmt.Sprintf("DELETE FROM %v;", tableName)); err != nil {
		_ = cerror.NewF(_bgCtx, cerror.KindInternal, "error delete from tables: %v", err).LogError()
	}
}
