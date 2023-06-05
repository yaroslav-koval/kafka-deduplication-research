package goch

import (
	"context"
	"kafka-polygon/pkg/cerror"
	"kafka-polygon/pkg/db/clickhouse"
	"kafka-polygon/pkg/env"
	"time"

	"github.com/jmoiron/sqlx"
)

type Connection struct {
	ctx context.Context
	db  *sqlx.DB
	cfg *env.ClickHouse
}

func NewConnection(ctx context.Context, p *env.ClickHouse) *Connection {
	return &Connection{ctx: ctx, cfg: p}
}

func (r *Connection) DB() *sqlx.DB {
	return r.db
}

// Connect initializes db connetion
func (r *Connection) Connect() error {
	dsn := clickhouse.DSN(r.cfg)

	db, err := sqlx.ConnectContext(r.ctx, "clickhouse", dsn)
	if err != nil {
		return cerror.New(r.ctx, cerror.KindDBOther, err).LogError()
	}

	db.SetMaxOpenConns(r.cfg.MaxOpenConns)
	db.SetMaxIdleConns(r.cfg.MaxIdleConns)
	db.SetConnMaxLifetime(time.Duration(r.cfg.MaxConnLifetime) * time.Second)
	db.SetConnMaxIdleTime(time.Duration(r.cfg.MaxConnLifetime) * time.Second)

	r.db = db

	return nil
}

// Close closes db connetion
func (r *Connection) Close() error {
	if success, err := r.ensureConnection(); !success {
		return err
	}

	if err := r.db.Close(); err != nil {
		return cerror.New(r.ctx, cerror.KindDBOther, err).LogError()
	}

	return nil
}

func (r *Connection) ensureConnection() (bool, error) {
	isConnected := r.db != nil
	if !isConnected {
		return false, cerror.NewF(r.ctx,
			cerror.KindDBOther, "db connection is not initialized").LogError()
	}

	return true, nil
}
