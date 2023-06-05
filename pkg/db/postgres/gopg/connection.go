package gopg

import (
	"context"
	"kafka-polygon/pkg/cerror"
	"kafka-polygon/pkg/db/postgres"
	"kafka-polygon/pkg/env"
	"kafka-polygon/pkg/log"
	"kafka-polygon/pkg/log/logger"
	"strings"

	"github.com/go-pg/pg/v10"
)

type Connection struct {
	db  *pg.DB
	cfg *env.Postgres
}

func NewConnection(p *env.Postgres) *Connection {
	return &Connection{cfg: p}
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

func (l *qLogger) BeforeQuery(ctx context.Context, _ *pg.QueryEvent) (context.Context, error) {
	return ctx, nil
}

func (l *qLogger) AfterQuery(ctx context.Context, q *pg.QueryEvent) error {
	if q.Err != nil && (l.logErr() || l.logAll()) {
		data, _ := q.FormattedQuery()
		_ = cerror.NewF(ctx, cerror.KindDBOther, "Connection query error: %+v. Query: %s", q.Err, string(data)).
			LogError()

		return nil
	}

	if l.logAll() {
		data, _ := q.FormattedQuery()
		log.Log(logger.NewEventF(ctx, logger.LevelDebug, "Connection query: %s", string(data)))

		return nil
	}

	return nil
}

func (r *Connection) DB() *pg.DB {
	return r.db
}

// Connect initializes db connection
func (r *Connection) Connect() error {
	dsn := postgres.DSN(r.cfg)
	opt, err := pg.ParseURL(dsn)

	if err != nil {
		return cerror.New(context.Background(), cerror.KindDBOther, err).LogError()
	}

	opt.PoolSize = r.cfg.MaxOpenConns
	opt.ReadTimeout = r.cfg.MaxConnLifetime
	opt.WriteTimeout = r.cfg.MaxConnLifetime
	db := pg.Connect(opt)

	ctx := context.Background()

	err = db.Ping(ctx)
	if err != nil {
		return cerror.New(ctx, cerror.DBToKind(err), err).LogError()
	}

	l := qLogger{level: r.cfg.DBQueryLogLevel}

	if !l.logNone() {
		db.AddQueryHook(&l)
	}

	r.db = db

	return nil
}

// Close closes db connection
func (r *Connection) Close() error {
	if success, err := r.ensureConnection(); !success {
		return err
	}

	if err := r.db.Close(); err != nil {
		return cerror.New(context.Background(), cerror.KindDBOther, err).LogError()
	}

	return nil
}

func (r *Connection) ensureConnection() (bool, error) {
	isConnected := r.db != nil
	if !isConnected {
		return false, cerror.NewF(
			context.Background(), cerror.KindDBOther, "db connection is not initialized").LogError()
	}

	return true, nil
}
