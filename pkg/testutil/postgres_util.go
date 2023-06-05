package testutil

import (
	"bytes"
	"context"
	"fmt"
	"kafka-polygon/pkg/cerror"
	"kafka-polygon/pkg/env"
	"kafka-polygon/pkg/log"
	"kafka-polygon/pkg/log/logger"
	"kafka-polygon/pkg/testutil/config"
	"strings"

	"github.com/golang-migrate/migrate/v4/database/postgres"
	bindata "github.com/golang-migrate/migrate/v4/source/go_bindata"
	"github.com/jmoiron/sqlx"
)

type DockerMigrateConfig struct {
	Name       string
	Version    uint
	SchemaName string
	Resource   *bindata.AssetSource
}

type DockerPGContainer struct {
	*DockerTestContainer

	// Container connection params
	initDB string

	// Postgres connection params
	Username string
	Password string
	SSLMode  string

	// Existing database params
	databases map[string]*DockerPGDatabaseInfo
}

func NewDockerPGContainer(cfg *config.TestContainerConfig) *DockerPGContainer {
	c := &DockerPGContainer{
		DockerTestContainer: &DockerTestContainer{
			Name:                GenerateCorrectName("postgres"),
			Repository:          "postgres",
			Tag:                 "14.0",
			Port:                "5432",
			TestContainerConfig: cfg,
			Host:                "localhost",
			activeConn:          0,
		},
		Username:  "postgres",
		Password:  "secret",
		SSLMode:   "disable",
		initDB:    "postgres",
		databases: make(map[string]*DockerPGDatabaseInfo),
	}
	c.Env = []string{
		fmt.Sprintf("POSTGRES_PASSWORD=%v", c.Password),
		fmt.Sprintf("POSTGRES_DB=%v", c.initDB),
	}

	return c
}

func (c *DockerPGContainer) GetDockerRunOptions() *dockertest.RunOptions {
	return c.DockerTestContainer.GetDockerRunOptions()
}

func (c *DockerPGContainer) GetPostgresEnvConfig() env.Postgres {
	return env.Postgres{
		Host:     c.Host,
		Port:     c.Port,
		Username: c.Username,
		Password: c.Password,
		SSLMode:  c.SSLMode,
	}
}

func (c *DockerPGContainer) RunWithOptions(opt *dockertest.RunOptions) *DockerPGContainer {
	c.DockerTestContainer.RunWithOptions(opt)
	return c
}

func (c *DockerPGContainer) CreatePostgresContainerDatabase(name string) *DockerPGContainer {
	ctx := context.Background()

	if _, ok := c.databases[name]; ok {
		return c
	}

	db := &DockerPGDatabaseInfo{
		name:              name,
		migrationVersions: make(map[string]uint),
	}

	var stdout bytes.Buffer

	if err := c.Pool.Retry(func() error {
		_, err := c.resource.Exec(
			[]string{
				"psql",
				"-qtAX",
				"-U", "postgres",
				"-c", fmt.Sprintf("select 1 from pg_database where datname = '%v'", name),
			},
			dockertest.ExecOptions{
				StdOut: &stdout,
			})
		if err != nil {
			return err
		}
		res := strings.TrimRight(stdout.String(), "\n")
		if res == "" {
			code, err := c.resource.Exec(
				[]string{
					"psql",
					"-qtAX",
					"-U", "postgres",
					"-c", fmt.Sprintf("CREATE DATABASE %v OWNER postgres", name),
				},
				dockertest.ExecOptions{
					StdOut: &stdout,
				},
			)
			if code != 0 {
				return fmt.Errorf("failed to execute resource command")
			}
			return err
		}
		return nil
	}); err != nil {
		_ = cerror.NewF(ctx, cerror.KindInternal, "failed to exec container command to create databes with name %v", name).
			LogFatal()
		return nil
	}

	c.databases[name] = db

	return c
}

func (c *DockerPGContainer) ConnectPostgresDB(name string) *DockerPGContainer {
	ctx := context.Background()
	dbInfo, ok := c.databases[name]

	if !ok {
		_ = cerror.NewF(ctx, cerror.KindInternal, "database must be init").LogFatal()
	} else if dbInfo.DBClient != nil {
		return c
	}

	if err := c.Pool.Retry(func() error {
		dataSourceName := fmt.Sprintf(defaultTemplateConnectionStringPostgres, c.Username, c.Password, c.Host, c.Port, name, c.SSLMode)
		db, internalErr := sqlx.Open("postgres", dataSourceName)
		if internalErr != nil {
			return internalErr
		}
		if internalErr := db.Ping(); internalErr != nil {
			return internalErr
		}
		dbInfo.DBClient = db
		return nil
	}); err != nil {
		_ = cerror.NewF(ctx, cerror.KindInternal, "could not connect to postgres database in docker with port %v", c.Port).
			LogFatal()
		return nil
	}
	c.activeConn++

	return c
}

func (c *DockerPGContainer) RunDBMigrations(ctx context.Context,
	dbName string, migrateCfgs ...DockerMigrateConfig) *DockerPGContainer {
	db, ok := c.ConnectPostgresDB(dbName).databases[dbName]
	if !ok {
		_ = cerror.NewF(ctx, cerror.KindInternal, "not exists current db %s", dbName).LogFatal()
		return nil
	}

	for _, migrateCfg := range migrateCfgs {
		err := c.bindData(ctx, db, migrateCfg)
		if err != nil {
			_ = cerror.New(ctx, cerror.KindInternal, err).LogFatal()
			return nil
		}

		db.migrationVersions[migrateCfg.Name] = migrateCfg.Version
	}

	db.isMigrationUp = true

	return c
}

func (c *DockerPGContainer) GetDBInfoByName(name string) *DockerPGDatabaseInfo {
	if db, ok := c.databases[name]; ok {
		return db
	}

	return nil
}

func (c *DockerPGContainer) bindData(ctx context.Context,
	db *DockerPGDatabaseInfo, migrateCfg DockerMigrateConfig) error {
	if db == nil || db.DBClient == nil {
		return cerror.NewF(ctx, cerror.KindInternal, "failed to init db instance").LogError()
	}

	binInstance, err := bindata.WithInstance(migrateCfg.Resource)

	if err != nil {
		return cerror.NewF(ctx, cerror.KindInternal, "failed to bin instance").LogError()
	}

	pCfg := &postgres.Config{
		MigrationsTable: migrateCfg.SchemaName,
	}

	targetInstance, err := postgres.WithInstance(db.DBClient.DB, pCfg)
	if err != nil {
		return cerror.NewF(ctx, cerror.KindInternal, "failed to run migration: %v", err).LogError()
	}

	m, err := migrate.NewWithInstance("go-bindata", binInstance, "postgres", targetInstance)

	if err != nil {
		return cerror.NewF(ctx, cerror.KindInternal, "[%s] failed to run migration: %v", migrateCfg.Name, err).LogError()
	}

	err = m.Migrate(migrateCfg.Version)

	if err != nil && err == migrate.ErrNoChange {
		log.Log(logger.NewEventF(ctx, logger.LevelDebug, "[%s] No new migrations found", migrateCfg.Name))
	} else if err != nil {
		return cerror.NewF(ctx, cerror.KindInternal,
			"[%s] failed to run migration %v", migrateCfg.Name, err).LogError()
	}

	return nil
}

//nolint:dupl
func (c *DockerPGContainer) CloseDBConnectionByName(name string) *DockerPGContainer {
	ctx := context.Background()
	db, ok := c.databases[name]

	if !ok {
		_ = cerror.NewF(
			ctx,
			cerror.KindInternal,
			"failed to get database by name. database not initialized or already closed with name %v", name).LogError()
	}

	if err := c.Pool.Retry(func() error {
		if err := db.DBClient.Close(); err != nil {
			return err
		}
		return nil
	}); err != nil {
		_ = cerror.NewF(
			ctx,
			cerror.KindInternal,
			"failed to get database by name. database not initialized or already closed with name %v", name).LogError()
	}

	delete(c.databases, name)
	c.closeConnection()

	return c
}
