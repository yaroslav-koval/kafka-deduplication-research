package postgres

import (
	"fmt"
	"kafka-polygon/pkg/env"
)

func DSN(cfg *env.Postgres) string {
	return fmt.Sprintf(
		"postgres://%s:%s@%s:%s/%s?sslmode=%s&connect_timeout=%d",
		cfg.Username,
		cfg.Password,
		cfg.Host,
		cfg.Port,
		cfg.DBName,
		cfg.SSLMode,
		cfg.ConnTimeout)
}
