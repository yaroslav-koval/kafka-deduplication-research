package clickhouse

import (
	"context"
	"fmt"
	"kafka-polygon/pkg/cerror"
	clickhouse "kafka-polygon/pkg/db/clickhouse/gomigrate"
	"kafka-polygon/pkg/env"
	"kafka-polygon/pkg/log"
	"kafka-polygon/pkg/log/logger"
	"time"

	bindata "github.com/golang-migrate/migrate/v4/source/go_bindata"
	"github.com/jmoiron/sqlx"
)

type MigrationCfg struct {
	pCfg       *env.ClickHouse
	mCfg       *env.Migration
	assetNames []string
	assetFn    func(name string) ([]byte, error)
}

func NewMigrationCfg(
	pCfg *env.ClickHouse,
	mCfg *env.Migration,
	assetNames []string,
	assetFn func(name string) ([]byte, error)) *MigrationCfg {
	return &MigrationCfg{
		pCfg:       pCfg,
		mCfg:       mCfg,
		assetNames: assetNames,
		assetFn:    assetFn,
	}
}

func Migrate(cfg *MigrationCfg) error {
	ctx := context.Background()

	binInstance, err := bindata.WithInstance(bindata.Resource(cfg.assetNames, cfg.assetFn))
	if err != nil {
		return cerror.New(ctx, cerror.KindInternal, err).LogError()
	}

	dsn := DSN(cfg.pCfg)
	fmt.Println(dsn)

	var conn *sqlx.DB

	for retries := 0; retries <= DefaultConnRetryCount; retries++ {
		conn, err = sqlx.Connect("clickhouse", dsn)
		if err == nil {
			break
		}

		err = cerror.New(ctx, cerror.KindInternal, err).LogError()

		time.Sleep(DefaultConnRetryTimeoutSec * time.Second)
	}

	if err != nil {
		return err
	}

	defer func() {
		if err = conn.Close(); err != nil {
			_ = cerror.New(ctx, cerror.KindInternal, err).LogError()
		}
	}()

	targetInstance, err := clickhouse.WithInstance(conn.DB, new(clickhouse.Config))
	if err != nil {
		return cerror.New(ctx, cerror.KindInternal, err).LogError()
	}

	m, err := migrate.NewWithInstance("go-bindata", binInstance, "clickhouse", targetInstance)
	if err != nil {
		return cerror.New(ctx, cerror.KindInternal, err).LogError()
	}

	err = m.Migrate(cfg.mCfg.Version)
	if err != nil && err == migrate.ErrNoChange {
		log.Log(logger.NewEvent(ctx, logger.LevelInfo, "no new migrations found"))
		return nil
	} else if err != nil {
		return cerror.New(ctx, cerror.KindInternal, err).LogError()
	}

	log.Log(logger.NewEventF(ctx, logger.LevelInfo, "migrated to version %d", cfg.mCfg.Version))

	return nil
}
