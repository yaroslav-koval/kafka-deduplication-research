package gomigrate_test

import (
	"bytes"
	"context"
	"kafka-polygon/pkg/db/clickhouse"
	"kafka-polygon/pkg/db/clickhouse/gomigrate"
	"kafka-polygon/pkg/env"
	"kafka-polygon/pkg/log"
	"kafka-polygon/pkg/testutil"
	"testing"

	_ "github.com/ClickHouse/clickhouse-go/v2" // clickhouse driver
	"github.com/golang-migrate/migrate/v4/database"
	"github.com/jmoiron/sqlx"
	"github.com/stretchr/testify/suite"
)

type gomigrateTestSuite struct {
	suite.Suite
	ctx    context.Context
	db     *sqlx.DB
	chCont *testutil.DockerCHContainer
	chCfg  *env.ClickHouse
}

func TestMigrationTestSuite(t *testing.T) {
	log.SetGlobalLogLevel("fatal")
	suite.Run(t, new(gomigrateTestSuite))
}

func (gmts *gomigrateTestSuite) SetupSuite() {
	gmts.ctx = context.Background()
	dbName := "gomigrate_test"
	chCont := testutil.NewDockerUtilInstance().
		InitCH().
		CreateCHContainerDatabase(dbName).
		ConnectCHDB(dbName)
	gmts.chCont = chCont
	gmts.db = chCont.GetDBInfoByName(dbName).DBClient

	chCfg := chCont.GetCHEnvConfig()
	chCfg.DBName = dbName
	gmts.chCfg = &chCfg
}

func (gmts *gomigrateTestSuite) TearDownSuite() {
	_ = gmts.chCont.CloseDBConnectionByName(gmts.chCfg.DBName)
}

func (gmts *gomigrateTestSuite) TestCHMigrate() {
	dsn := clickhouse.DSN(gmts.chCfg)
	dbCH := gomigrate.ClickHouse{}
	d, err := dbCH.Open(dsn)
	gmts.NoError(err)

	defer d.Close()

	err = d.Run(bytes.NewReader([]byte("CREATE TABLE bar (id UUID, bar String)ENGINE=MergeTree() PRIMARY KEY id ORDER BY id;")))
	gmts.NoError(err)

	var exists bool

	err = d.(*gomigrate.ClickHouse).Conn().QueryRowContext(gmts.ctx, `select count(*) FROM system.tables
                where database = $1
                  and name = $2`, gmts.chCfg.DBName, "bar").Scan(&exists)
	gmts.NoError(err)
	gmts.Equal(true, exists)
}

func (gmts *gomigrateTestSuite) TestCHMigrateError() {
	dsn := clickhouse.DSN(gmts.chCfg)
	dbCH := gomigrate.ClickHouse{}
	d, err := dbCH.Open(dsn)
	gmts.NoError(err)

	defer d.Close()

	wantErr := "migration failed in line 0: CREATE TABLE foo (id UUID, foo text)ENGINE=MergeTree() PRIMARY KEY id ORDER BY id; " +
		"CREATE TABLEE bar (id UUID, bar text)ENGINE=MergeTree() PRIMARY KEY id ORDER BY id; (details: code: 62, " +
		"message: Syntax error (Multi-statements are not allowed): failed at position 82 (end of query): ; " +
		"CREATE TABLEE bar (id UUID, bar text)ENGINE=MergeTree() PRIMARY KEY id ORDER BY id;.)"

	err = d.Run(bytes.NewReader([]byte("CREATE TABLE foo (id UUID, foo text)ENGINE=MergeTree() PRIMARY KEY id ORDER BY id; " +
		"CREATE TABLEE bar (id UUID, bar text)ENGINE=MergeTree() PRIMARY KEY id ORDER BY id;")))
	gmts.Error(err)
	gmts.Equal(wantErr, err.Error())
}

func (gmts *gomigrateTestSuite) TestCHMigrateLock() {
	dsn := clickhouse.DSN(gmts.chCfg)
	dbCH := gomigrate.ClickHouse{}
	d, err := dbCH.Open(dsn)
	gmts.NoError(err)

	defer d.Close()

	ps := d.(*gomigrate.ClickHouse)

	err = ps.Lock()
	gmts.NoError(err)

	err = ps.Unlock()
	gmts.NoError(err)

	err = ps.Lock()
	gmts.NoError(err)

	err = ps.Unlock()
	gmts.NoError(err)
}

func (gmts *gomigrateTestSuite) TestCHMigrateVersion() {
	dsn := clickhouse.DSN(gmts.chCfg)
	dbCH := gomigrate.ClickHouse{}
	d, err := dbCH.Open(dsn)
	gmts.NoError(err)

	defer d.Close()

	ver, dirty, err := d.Version()
	gmts.Equal(database.NilVersion, ver)
	gmts.Equal(false, dirty)
	gmts.NoError(err)

	err = d.SetVersion(1, true)
	gmts.NoError(err)

	ver, dirty, err = d.Version()
	gmts.Equal(1, ver)
	gmts.Equal(true, dirty)
	gmts.NoError(err)
}

func (gmts *gomigrateTestSuite) TestCHMigrateConn() {
	dsn := clickhouse.DSN(gmts.chCfg)
	dbCH := gomigrate.ClickHouse{}
	d, err := dbCH.Open(dsn)
	gmts.NoError(err)

	defer d.Close()

	var dd string
	err = d.(*gomigrate.ClickHouse).Conn().QueryRowContext(gmts.ctx, `SELECT currentDatabase()`, nil).Scan(&dd)
	gmts.NoError(err)
	gmts.Equal(gmts.chCfg.DBName, dd)
}

func (gmts *gomigrateTestSuite) TestCHMigrateDrop() {
	dsn := clickhouse.DSN(gmts.chCfg)
	dbCH := gomigrate.ClickHouse{}
	d, err := dbCH.Open(dsn)
	gmts.NoError(err)

	defer d.Close()

	err = d.Drop()
	gmts.NoError(err)
}

func (gmts *gomigrateTestSuite) TestCHMigrateOpenError() {
	nCfg := &env.ClickHouse{
		Host:            gmts.chCfg.Host,
		Port:            gmts.chCfg.Port,
		Username:        gmts.chCfg.Username,
		Password:        gmts.chCfg.Password,
		DBName:          "not_exists_database",
		ConnTimeout:     gmts.chCfg.ConnTimeout,
		Compress:        gmts.chCfg.Compress,
		Debug:           gmts.chCfg.Debug,
		MaxOpenConns:    gmts.chCfg.MaxOpenConns,
		MaxIdleConns:    gmts.chCfg.MaxIdleConns,
		MaxConnLifetime: gmts.chCfg.MaxConnLifetime,
	}
	dsn := clickhouse.DSN(nCfg)
	dbCH := gomigrate.ClickHouse{}
	d, err := dbCH.Open(dsn)
	gmts.Error(err)
	gmts.Nil(d)
	gmts.Equal("code: 81, message: Database `not_exists_database` doesn't exist", err.Error())
}
