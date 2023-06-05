package postgres_test

import (
	"context"
	"fmt"
	"kafka-polygon/pkg/db/postgres"
	"kafka-polygon/pkg/db/postgres/testmigration/schema"
	"kafka-polygon/pkg/env"
	"kafka-polygon/pkg/log"
	"kafka-polygon/pkg/testutil"
	"testing"

	_ "github.com/jackc/pgx/v4/stdlib" // postgres driver
	"github.com/jmoiron/sqlx"
	"github.com/stretchr/testify/suite"
)

type dbTestSuite struct {
	suite.Suite
	db     *sqlx.DB
	pgCont *testutil.DockerPGContainer
	pgCfg  *env.Postgres
}

func TestRepoTestSuite(t *testing.T) {
	log.SetGlobalLogLevel("fatal")
	suite.Run(t, new(dbTestSuite))
}

func (s *dbTestSuite) SetupSuite() {
	dbName := "migration_test"
	pgCont := testutil.NewDockerUtilInstance().
		InitPG().
		CreatePostgresContainerDatabase(dbName).
		ConnectPostgresDB(dbName)
	s.pgCont = pgCont
	s.db = pgCont.GetDBInfoByName(dbName).DBClient

	postgresCfg := pgCont.GetPostgresEnvConfig()
	postgresCfg.DBName = dbName
	s.pgCfg = &postgresCfg
}

func (s *dbTestSuite) TearDownSuite() {
	_ = s.pgCont.CloseDBConnectionByName(s.pgCfg.DBName)
}

func (s *dbTestSuite) TestConst() {
	sConst := map[string]string{
		string(postgres.TxContextKey): "DB_CONTEXT_TRANSACTION_KEY",
		postgres.MigrateNameWorkflow:  "migrate_workflow",
		postgres.MigrateName:          "migrate",
	}

	for actual, expected := range sConst {
		s.Equal(actual, expected)
	}

	s.Equal(20, postgres.DefaultConnRetryCount)
	s.Equal(3, postgres.DefaultConnRetryTimeoutSec)
}

func (s *dbTestSuite) TestDSN() {
	testConfig := &env.Postgres{
		Host:        "localhost",
		Port:        "1234",
		Username:    "admin",
		Password:    "secret-password",
		DBName:      "test_name",
		SSLMode:     "disable",
		ConnTimeout: 222,
	}
	testDSNStr := fmt.Sprintf(
		"postgres://%s:%s@%s:%s/%s?sslmode=%s&connect_timeout=%d",
		testConfig.Username,
		testConfig.Password,
		testConfig.Host,
		testConfig.Port,
		testConfig.DBName,
		testConfig.SSLMode,
		testConfig.ConnTimeout)
	s.Equal(testDSNStr, postgres.DSN(testConfig))
}

func (s *dbTestSuite) TestMigrate() {
	ctx := context.Background()
	testTable := "test_table"
	migVer := &env.Migration{Version: 1}

	// initial migration, creation of test_table
	err := postgres.NewMigration(s.pgCfg, postgres.Setting{
		Name:       postgres.MigrateName,
		Version:    migVer.Version,
		AssetNames: schema.AssetNames(),
		AssetFn:    schema.Asset,
	}).Migrate(ctx)
	s.NoError(err)

	// checks if migration was successful, and we have test_table with 1 column
	columnCount, dbErr := s.selectColumnCount(s.pgCfg.DBName, testTable)
	s.NoError(dbErr)
	s.Equal(1, columnCount)

	migVer.Version = 2

	// trying to apply migration #2
	err = postgres.NewMigration(s.pgCfg, postgres.Setting{
		Name:       postgres.MigrateName,
		Version:    migVer.Version,
		AssetNames: schema.AssetNames(),
		AssetFn:    schema.Asset,
	}).Migrate(ctx)
	s.NoError(err)

	// checks if migration was successful, and we have test_table with 3 column
	columnCount, dbErr = s.selectColumnCount(s.pgCfg.DBName, testTable)
	s.NoError(dbErr)
	s.Equal(3, columnCount)

	migVer.Version = 1

	// trying to downgrade migration to version 1
	err = postgres.NewMigration(s.pgCfg, postgres.Setting{
		Name:       postgres.MigrateName,
		Version:    migVer.Version,
		AssetNames: schema.AssetNames(),
		AssetFn:    schema.Asset,
	}).Migrate(ctx)
	s.NoError(err)

	// checks if migration was successful, and we have test_table with 1 column
	columnCount, dbErr = s.selectColumnCount(s.pgCfg.DBName, testTable)
	s.NoError(dbErr)
	s.Equal(1, columnCount)
}

func (s *dbTestSuite) selectColumnCount(dbName, tableName string) (int, error) {
	var count int

	err := s.db.QueryRow(
		`select count(*) from information_schema.columns
                where table_catalog = $1
                  and table_schema = 'public'
                  and table_name = $2`, dbName, tableName).Scan(&count)

	if err != nil {
		return count, err
	}

	return count, nil
}
