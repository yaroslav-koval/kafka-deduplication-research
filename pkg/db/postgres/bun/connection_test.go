package bun_test

import (
	"kafka-polygon/pkg/cerror"
	"kafka-polygon/pkg/db/postgres/bun"
	"kafka-polygon/pkg/env"
	"kafka-polygon/pkg/log"
	"kafka-polygon/pkg/testutil"
	"testing"

	_ "github.com/jackc/pgx/v4/stdlib" // postgres driver
	"github.com/jmoiron/sqlx"
	"github.com/stretchr/testify/suite"
)

type dbConnectionTestSuite struct {
	suite.Suite
	db     *sqlx.DB
	pgCont *testutil.DockerPGContainer
	pgConn *bun.Connection
	pgCfg  *env.Postgres
}

func TestRepoTestSuite(t *testing.T) {
	log.SetGlobalLogLevel("fatal")
	suite.Run(t, new(dbConnectionTestSuite))
}

func (s *dbConnectionTestSuite) SetupSuite() {
	dbName := "connection_test"
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
}

func (s *dbConnectionTestSuite) TearDownSuite() {
	_ = s.pgCont.CloseDBConnectionByName(s.pgCfg.DBName)
}

func (s *dbConnectionTestSuite) TestConnectAndClose() {
	// trying to close db connection when it is not opened
	err := s.pgConn.Close()
	s.Error(err)
	s.Equal("db connection is not initialized", err.Error())
	s.Equal(cerror.KindDBOther, cerror.ErrKind(err))

	s.NotNil(s.pgConn)

	// trying to get db connection when it hasn't established yet
	s.Nil(s.pgConn.DB())

	// trying to connect to db
	err = s.pgConn.Connect()
	s.NoError(err)
	s.NotNil(s.pgConn.DB())

	// trying to close db connection
	err = s.pgConn.Close()
	s.NoError(err)

	// trying to close db connection when it is already closed
	err = s.pgConn.Close()
	s.NoError(err)
}
