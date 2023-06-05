package goch_test

import (
	"context"
	"kafka-polygon/pkg/cerror"
	"kafka-polygon/pkg/db/clickhouse/goch"
	"kafka-polygon/pkg/env"
	"kafka-polygon/pkg/log"
	"kafka-polygon/pkg/testutil"
	"testing"

	_ "github.com/ClickHouse/clickhouse-go/v2" // clickhouse driver
	"github.com/jmoiron/sqlx"
	"github.com/stretchr/testify/suite"
)

type dbTestSuite struct {
	suite.Suite
	ctx    context.Context
	db     *sqlx.DB
	chCont *testutil.DockerCHContainer
	chCfg  *env.ClickHouse
	chConn *goch.Connection
}

func TestRepoTestSuite(t *testing.T) {
	log.SetGlobalLogLevel("fatal")
	suite.Run(t, new(dbTestSuite))
}

func (s *dbTestSuite) SetupSuite() {
	s.ctx = context.Background()
	dbName := "connection_test"
	chCont := testutil.NewDockerUtilInstance().
		InitCH().
		CreateCHContainerDatabase(dbName).
		ConnectCHDB(dbName)
	s.chCont = chCont
	s.db = chCont.GetDBInfoByName(dbName).DBClient

	chCfg := chCont.GetCHEnvConfig()
	chCfg.DBName = dbName
	s.chCfg = &chCfg
	s.chConn = goch.NewConnection(s.ctx, s.chCfg)
}

func (s *dbTestSuite) TearDownSuite() {
	_ = s.chCont.CloseDBConnectionByName(s.chCfg.DBName)
}

func (s *dbTestSuite) TestConnectAndClose() {
	err := s.chConn.Close()
	s.Error(err)
	s.Equal("db connection is not initialized", err.Error())
	s.Equal(cerror.KindDBOther, cerror.ErrKind(err))

	s.NotNil(s.chConn)
	s.Nil(s.chConn.DB())

	err = s.chConn.Connect()
	s.NoError(err)
	s.NotNil(s.chConn.DB())

	err = s.chConn.DB().PingContext(context.Background())
	s.NoError(err)

	err = s.chConn.Close()
	s.NoError(err)
}
