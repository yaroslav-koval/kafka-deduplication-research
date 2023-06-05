package postgres

type key string

const (
	DefaultConnRetryCount      = 20
	DefaultConnRetryTimeoutSec = 3

	TxContextKey key = "DB_CONTEXT_TRANSACTION_KEY"

	MigrateNameWorkflow = "migrate_workflow"
	MigrateName         = "migrate"
)
