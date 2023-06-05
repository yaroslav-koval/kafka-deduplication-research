package clickhouse

import (
	"fmt"
	"kafka-polygon/pkg/env"
)

const (
	DefaultConnRetryCount      = 20
	DefaultConnRetryTimeoutSec = 3
)

func DSN(cfg *env.ClickHouse) string {
	return fmt.Sprintf(
		"clickhouse://%s:%s@%s:%s/%s?dial_timeout=%s&compress=%v&debug=%v&max_execution_time=60",
		cfg.Username,
		cfg.Password,
		cfg.Host,
		cfg.Port,
		cfg.DBName,
		cfg.ConnTimeout,
		cfg.Compress,
		cfg.Debug)
}
