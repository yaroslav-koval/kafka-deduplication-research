package clickhouse_test

import (
	"context"
	"kafka-polygon/pkg/db/clickhouse"
	"kafka-polygon/pkg/db/clickhouse/goch"
	"kafka-polygon/pkg/db/clickhouse/testmigration/schema"
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

func TestMigrationTestSuite(t *testing.T) {
	log.SetGlobalLogLevel("fatal")
	suite.Run(t, new(dbTestSuite))
}

func (s *dbTestSuite) SetupSuite() {
	s.ctx = context.Background()
	dbName := "migrations_test"
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

func (s *dbTestSuite) TestMigrate() {
	testTable := "instance_test"

	err := s.chConn.Close()
	s.Error(err)
	s.Equal("db connection is not initialized", err.Error())

	s.NotNil(s.chConn)
	s.Nil(s.chConn.DB())

	err = s.chConn.Connect()
	s.NoError(err)
	s.NotNil(s.chConn.DB())

	err = s.chConn.DB().PingContext(context.Background())
	s.NoError(err)

	migVer := &env.Migration{Version: 1}
	testCnf := clickhouse.NewMigrationCfg(s.chCfg, migVer, schema.AssetNames(), schema.Asset)

	err = clickhouse.Migrate(testCnf)
	s.NoError(err)

	// checks if migration was successful, and we have test_table with 1 column
	columnCount, dbErr := s.selectColumnCount(s.chCfg.DBName, testTable)
	s.NoError(dbErr)
	s.Equal(1, columnCount)

	err = s.chConn.Close()
	s.NoError(err)
}

func (s *dbTestSuite) selectColumnCount(dbName, tableName string) (int, error) {
	var count int

	err := s.db.QueryRow(
		`select count(*) FROM system.tables
                where database = $1
                  and name = $2`, dbName, tableName).Scan(&count)

	if err != nil {
		return count, err
	}

	return count, nil
}
