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

	"github.com/golang-migrate/migrate/v4/database/postgres"
	bindata "github.com/golang-migrate/migrate/v4/source/go_bindata"
	"github.com/jmoiron/sqlx"
)

type DockerCHContainer struct {
	*DockerTestContainer
	Username string
	Password string
	initDB   string
	// Existing database params
	databases map[string]*DockerCHDatabaseInfo
}

func NewDockerCHContainer(cfg *config.TestContainerConfig) *DockerCHContainer {
	c := &DockerCHContainer{
		DockerTestContainer: &DockerTestContainer{
			Name:                GenerateCorrectName("clickhouse"),
			Repository:          "yandex/clickhouse-server",
			Tag:                 "21.3",
			Port:                "9000",
			TestContainerConfig: cfg,
			Host:                "localhost",
			activeConn:          0,
		},
		Username:  "default",
		initDB:    "test-ch",
		databases: make(map[string]*DockerCHDatabaseInfo),
	}

	return c
}

func (c *DockerCHContainer) GetDockerRunOptions() *dockertest.RunOptions {
	return c.DockerTestContainer.GetDockerRunOptions()
}

func (c *DockerCHContainer) GetCHEnvConfig() env.ClickHouse {
	return env.ClickHouse{
		Host:        c.Host,
		Port:        c.Port,
		Username:    c.Username,
		Password:    c.Password,
		DBName:      c.initDB,
		ConnTimeout: "20s",
	}
}

func (c *DockerCHContainer) RunWithOptions(opt *dockertest.RunOptions) *DockerCHContainer {
	c.DockerTestContainer.RunWithOptions(opt)
	return c
}

func (c *DockerCHContainer) CreateCHContainerDatabase(name string) *DockerCHContainer {
	ctx := context.Background()

	if _, ok := c.databases[name]; ok {
		return c
	}

	db := &DockerCHDatabaseInfo{
		name: name,
	}

	var stdout bytes.Buffer

	if err := c.Pool.Retry(func() error {
		code, err := c.resource.Exec(
			[]string{
				"clickhouse",
				"client",
				"-q",
				fmt.Sprintf("CREATE DATABASE IF NOT EXISTS %v", name),
			},
			dockertest.ExecOptions{
				StdOut: &stdout,
			})
		if err != nil {
			return err
		}

		if code != 0 {
			return fmt.Errorf("failed to execute resource command")
		}

		return nil
	}); err != nil {
		_ = cerror.NewF(ctx, cerror.KindInternal, "failed to exec container command to create database with name %v", name).
			LogFatal()
		return nil
	}

	c.databases[name] = db

	return c
}

func (c *DockerCHContainer) ConnectCHDB(name string) *DockerCHContainer {
	ctx := context.Background()
	dbInfo, ok := c.databases[name]

	if !ok {
		_ = cerror.NewF(ctx, cerror.KindInternal, "database must be init").LogFatal()
	} else if dbInfo.DBClient != nil {
		return c
	}

	if err := c.Pool.Retry(func() error {
		dsn := fmt.Sprintf("clickhouse://%s:%s@%s:%s/%s?debug=false",
			c.Username, c.Password, c.Host, c.Port, name)
		fmt.Println(dsn)
		db, err := sqlx.Connect("clickhouse", dsn)
		if err != nil {
			return err
		}
		if internalErr := db.Ping(); internalErr != nil {
			return internalErr
		}
		dbInfo.DBClient = db
		return nil
	}); err != nil {
		_ = cerror.NewF(ctx, cerror.KindInternal, "could not connect to clickhouse database in docker with port %v", c.Port).
			LogFatal()
		return nil
	}
	c.activeConn++

	return c
}

func (c *DockerCHContainer) RunDBMigrations(name string, mv uint, resource *bindata.AssetSource) *DockerCHContainer {
	ctx := context.Background()
	binInstance, err := bindata.WithInstance(resource)

	if err != nil {
		_ = cerror.NewF(ctx, cerror.KindInternal, "failed to init db instance").LogFatal()
		return nil
	}

	db := c.ConnectCHDB(name).databases[name]

	targetInstance, err := postgres.WithInstance(db.DBClient.DB, new(postgres.Config))
	if err != nil {
		_ = cerror.NewF(ctx, cerror.KindInternal, "failed to run migration %v", c.Port).LogFatal()
		return nil
	}

	m, err := migrate.NewWithInstance("go-bindata", binInstance, "postgres", targetInstance)

	if err != nil {
		_ = cerror.NewF(ctx, cerror.KindInternal, "failed to run migration %v", c.Port).LogFatal()
		return nil
	}

	err = m.Migrate(mv)

	if err != nil && err == migrate.ErrNoChange {
		log.Log(logger.NewEventF(ctx, logger.LevelDebug, "No new migrations found"))
		return c
	} else if err != nil {
		_ = cerror.NewF(ctx, cerror.KindInternal, "failed to run migration %v", c.Port).LogFatal()
		return nil
	}

	db.isMigrationUp = true
	db.migrationVersion = mv

	return c
}

func (c *DockerCHContainer) GetDBInfoByName(name string) *DockerCHDatabaseInfo {
	if db, ok := c.databases[name]; ok {
		return db
	}

	return nil
}

//nolint:dupl
func (c *DockerCHContainer) CloseDBConnectionByName(name string) *DockerCHContainer {
	ctx := context.Background()
	db, ok := c.databases[name]

	if !ok {
		_ = cerror.NewF(
			ctx,
			cerror.KindInternal,
			"failed to get database by name. database not initialized or already closed with name %v", name).
			LogError()
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
			"failed to get database by name. database not initialized or already closed with name %v", name).
			LogError()
	}

	delete(c.databases, name)
	c.closeConnection()

	return c
}
