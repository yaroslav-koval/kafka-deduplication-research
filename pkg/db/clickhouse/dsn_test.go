package clickhouse_test

import (
	"kafka-polygon/pkg/db/clickhouse"
	"kafka-polygon/pkg/env"
	"testing"

	"github.com/tj/assert"
)

func TestDSN(t *testing.T) {
	t.Parallel()

	chConf := &env.ClickHouse{
		Host:     "test",
		Port:     "8080",
		Username: "test-username",
		Password: "test-password",
		DBName:   "test-dbname",
	}

	dsn := clickhouse.DSN(chConf)
	assert.Equal(t,
		"clickhouse://test-username:test-password@test:8080/test-dbname?dial_timeout=&compress=false&debug=false&max_execution_time=60",
		dsn)
}
