package cerror

import (
	"database/sql"

	"github.com/go-pg/pg/v10"
)

type KindPostgres uint8

const (
	_ KindPostgres = iota
	KindPostgresOther
	KindPostgresNoRows
	KindPostgresMultiRows
	KindPostgresPermission
	KindPostgresIO

	_kindPostgresNoRows     = "no_rows"
	_kindPostgresMultiRows  = "multi_rows"
	_kindPostgresOther      = "other"
	_kindPostgresPermission = "permission"
	_kindPostgresIO         = "io"
)

// Postgres error codes https://www.postgresql.org/docs/current/errcodes-appendix.html
var (
	_dbPermissionCodes = []string{
		"0L000", // invalid_grantor
		"0LP01", // invalid_grant_operation
		"0P000", // invalid_role_specification
		"28000", // invalid_authorization_specification
		"28P01", // invalid_password
		"42000", // syntax_error_or_access_rule_violation
		"42501", // insufficient_privilege
	}

	_dbIOCodes = []string{
		"25006", // read_only_sql_transaction
		"25P03", // idle_in_transaction_session_timeout
		"53300", // too_many_connections
		"58030", // io_error
	}
)

func (k KindPostgres) String() string {
	switch k {
	case KindPostgresNoRows:
		return _kindPostgresNoRows
	case KindPostgresMultiRows:
		return _kindPostgresMultiRows
	case KindPostgresPermission:
		return _kindPostgresPermission
	case KindPostgresIO:
		return _kindPostgresIO
	default:
		return _kindPostgresOther
	}
}

func KindFromPostgres(err error) KindPostgres {
	switch {
	case err == pg.ErrNoRows, err == sql.ErrNoRows:
		return KindPostgresNoRows
	case err == pg.ErrMultiRows:
		return KindPostgresMultiRows
	case isStrContainsInsensitive(err.Error(), _dbPermissionCodes):
		return KindPostgresPermission
	case isStrContainsInsensitive(err.Error(), _dbIOCodes):
		return KindPostgresIO
	}

	return KindPostgresOther
}
