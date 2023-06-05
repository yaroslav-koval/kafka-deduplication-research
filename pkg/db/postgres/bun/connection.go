package bun

import (
	"context"
	"database/sql"
	"fmt"
	"kafka-polygon/pkg/cerror"
	"kafka-polygon/pkg/db/postgres"
	"kafka-polygon/pkg/env"
	"kafka-polygon/pkg/log"
	"kafka-polygon/pkg/log/logger"
	"strings"
	"time"

	"github.com/uptrace/bun"
	"github.com/uptrace/bun/dialect/pgdialect"
	"github.com/uptrace/bun/driver/pgdriver"
	"github.com/uptrace/bun/extra/bunotel"
)

type Connection struct {
	db  *bun.DB
	cfg *env.Postgres
}

func NewConnection(cfg *env.Postgres) *Connection {
	return &Connection{cfg: cfg}
}

// Log levels
const (
	logLvlAll  = "all"
	logLvlErr  = "error"
	logLvlNone = "none"
)

type qLogger struct {
	level string
}

func (l *qLogger) logAll() bool {
	return strings.EqualFold(l.level, logLvlAll)
}

func (l *qLogger) logErr() bool {
	return strings.EqualFold(l.level, logLvlErr)
}

func (l *qLogger) logNone() bool {
	return strings.EqualFold(l.level, logLvlNone)
}

func (l *qLogger) BeforeQuery(ctx context.Context, _ *bun.QueryEvent) context.Context {
	return ctx
}

func (l *qLogger) AfterQuery(ctx context.Context, event *bun.QueryEvent) {
	msg := fmt.Sprintf("error: %+v. Query: %s operation: %s duration: %s",
		event.Err, event.Query, event.Operation(), time.Since(event.StartTime).String())

	if event.Err != nil && (l.logErr() || l.logAll()) {
		_ = cerror.NewF(ctx, cerror.KindDBOther, "Connection query %s", msg).
			LogError()

		return
	}

	if l.logAll() {
		log.Log(logger.NewEventF(ctx, logger.LevelDebug, "%s", msg))
	}
}

func (c *Connection) DB() *bun.DB {
	return c.db
}

// Connect initializes db connection
func (c *Connection) Connect() error {
	dsn := postgres.DSN(c.cfg)

	opts := []pgdriver.Option{
		pgdriver.WithDSN(dsn),
		pgdriver.WithTimeout(c.cfg.MaxConnLifetime),
		pgdriver.WithDialTimeout(c.cfg.MaxConnLifetime),
		pgdriver.WithReadTimeout(c.cfg.MaxConnLifetime),
		pgdriver.WithWriteTimeout(c.cfg.MaxConnLifetime),
	}

	sqldb := sql.OpenDB(pgdriver.NewConnector(opts...))
	sqldb.SetMaxOpenConns(c.cfg.MaxOpenConns)
	sqldb.SetMaxIdleConns(c.cfg.MaxOpenConns)

	db := bun.NewDB(sqldb, pgdialect.New())

	ctx := context.Background()

	err := db.PingContext(ctx)
	if err != nil {
		return cerror.New(ctx, cerror.DBToKind(err), err).LogError()
	}

	l := qLogger{level: c.cfg.DBQueryLogLevel}

	if !l.logNone() {
		db.AddQueryHook(&l)
	}

	// also add request id
	c.db = db

	return nil
}

// Close closes db connection
func (c *Connection) Close() error {
	if success, err := c.ensureConnection(); !success {
		return err
	}

	if err := c.db.Close(); err != nil {
		return cerror.New(context.Background(), cerror.KindDBOther, err).LogError()
	}

	return nil
}

func (c *Connection) EnableTracing() {
	c.db.AddQueryHook(bunotel.NewQueryHook(bunotel.WithDBName(c.cfg.DBName)))
}

func (c *Connection) ensureConnection() (bool, error) {
	isConnected := c.db != nil
	if !isConnected {
		return false, cerror.NewF(
			context.Background(), cerror.KindDBOther, "db connection is not initialized").LogError()
	}

	return true, nil
}
