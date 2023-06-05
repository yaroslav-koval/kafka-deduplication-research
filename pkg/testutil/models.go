package testutil

import (
	"github.com/jmoiron/sqlx"
)

type DockerPGDatabaseInfo struct {
	name              string
	migrationVersions map[string]uint
	DBClient          *sqlx.DB
	isMigrationUp     bool
}

type DockerCHDatabaseInfo struct {
	name             string
	migrationVersion uint
	DBClient         *sqlx.DB
	isMigrationUp    bool
}
