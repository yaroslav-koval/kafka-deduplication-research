package cerror

import (
	"database/sql"
	"net/http"
	"strings"

	"github.com/go-pg/pg/v10"
)

type DBKind uint8

const (
	_ DBKind = iota
	KindDBOther
	KindDBNoRows
	KindDBMultiRows
	KindDBPermission // KindPermission denied
	KindDBIO         // External I/O error such as network failure

	dbNoRows           = "db_no_rows_error"
	dbMultiRows        = "db_multi_rows_error"
	dbUnknown          = "db_unknown_error"
	dbOther            = "db_other_error" // king message
	dbPermissionDenied = "db_permission_denied"
	dbIO               = "db_io_error"
)

// List postgres code https://www.postgresql.org/docs/current/errcodes-appendix.html
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

func (k DBKind) String() string {
	switch k {
	case KindDBNoRows:
		return dbNoRows
	case KindDBMultiRows:
		return dbMultiRows
	case KindDBPermission:
		return dbPermissionDenied
	case KindDBIO:
		return dbIO
	case KindDBOther:
		return dbOther
	}

	return dbUnknown
}

func (k DBKind) HTTPCode() int {
	switch k {
	case KindDBNoRows:
		return http.StatusNotFound
	default:
		return http.StatusInternalServerError
	}
}

func (k DBKind) Group() KindGroup {
	return GroupDB
}

func DBToKind(err error) DBKind {
	switch {
	case err == pg.ErrNoRows, err == sql.ErrNoRows:
		return KindDBNoRows
	case err == pg.ErrMultiRows:
		return KindDBMultiRows
	case checkCode(err.Error(), _dbPermissionCodes):
		return KindDBPermission
	case checkCode(err.Error(), _dbIOCodes):
		return KindDBIO
	default:
		return KindDBOther
	}
}

func checkCode(msg string, list []string) bool {
	for _, code := range list {
		if strings.Contains(strings.ToLower(msg), strings.ToLower(code)) {
			return true
		}
	}

	return false
}
