package testutil

import "time"

const (
	defaultTemplateContainerName            = "test_api_%v_container"
	defaultTemplateConnectionStringPostgres = "postgres://%s:%s@%s:%s/%s?sslmode=%s"

	DockerContainerResourceLifeTimeInSec = 60
	defaultTimeout                       = 10 * time.Second
)
