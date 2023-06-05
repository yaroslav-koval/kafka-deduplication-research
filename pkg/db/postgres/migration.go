package postgres

import (
	"context"
	"fmt"
	"kafka-polygon/pkg/cerror"
	"kafka-polygon/pkg/env"
	"kafka-polygon/pkg/log"
	"kafka-polygon/pkg/log/logger"
	"time"

	"github.com/golang-migrate/migrate"
	"github.com/golang-migrate/migrate/database/postgres"
	bindata "github.com/golang-migrate/migrate/source/go_bindata"
	"github.com/jmoiron/sqlx"
)

type Setting struct {
	Name        string
	Version     uint
	SchemaTable string
	AssetNames  []string
	AssetFn     func(name string) ([]byte, error)
}

type Migration struct {
	conn     *sqlx.DB
	pCfg     *env.Postgres
	settings []Setting
}

func NewMigration(pCfg *env.Postgres, st ...Setting) *Migration {
	return &Migration{
		pCfg:     pCfg,
		settings: st,
	}
}

func (mc *Migration) Migrate(ctx context.Context) error {
	defer func() {
		if err := mc.close(); err != nil {
			_ = cerror.New(ctx, cerror.KindInternal, err).LogError()
		}
	}()

	err := mc.connect(ctx)
	if err != nil {
		return err
	}

	for _, st := range mc.settings {
		err = mc.binData(ctx, st)
		if err != nil {
			return err
		}
	}

	return nil
}

func (mc *Migration) connect(ctx context.Context) error {
	dsn := DSN(mc.pCfg)

	var (
		conn *sqlx.DB
		err  error
	)

	for retries := 0; retries <= DefaultConnRetryCount; retries++ {
		conn, err = sqlx.Connect("postgres", dsn)
		if err == nil {
			break
		}

		err = cerror.New(ctx, cerror.KindInternal, err).LogError()

		time.Sleep(DefaultConnRetryTimeoutSec * time.Second)
	}

	if err != nil {
		return err
	}

	mc.conn = conn

	return nil
}

func (mc *Migration) close() error {
	return mc.conn.Close()
}

func (mc *Migration) binData(ctx context.Context, st Setting) error {
	binInstance, err := bindata.WithInstance(bindata.Resource(st.AssetNames, st.AssetFn))
	if err != nil {
		return cerror.New(ctx, cerror.KindInternal, err).LogError()
	}

	targetInstance, err := postgres.WithInstance(mc.conn.DB, &postgres.Config{
		MigrationsTable: st.SchemaTable,
	})

	if err != nil {
		return cerror.New(ctx, cerror.KindInternal, err).LogError()
	}

	m, err := migrate.NewWithInstance(fmt.Sprintf("go-bindata-%s", st.Name), binInstance, "postgres", targetInstance)
	if err != nil {
		return cerror.New(ctx, cerror.KindInternal, err).LogError()
	}

	err = m.Migrate(st.Version)
	if err != nil && err == migrate.ErrNoChange {
		log.Log(logger.NewEvent(ctx, logger.LevelInfo, fmt.Sprintf("[%s] no new migrations found", st.Name)))
		return nil
	} else if err != nil {
		return cerror.New(ctx, cerror.KindInternal, err).LogError()
	}

	log.Log(logger.NewEventF(ctx, logger.LevelInfo, "[%s] migrated to version %d", st.Name, st.Version))

	return nil
}
