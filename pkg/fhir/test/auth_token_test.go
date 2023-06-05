package test

import (
	"kafka-polygon/pkg/fhir"
	"testing"

	"github.com/tj/assert"
)

func TestParse(t *testing.T) {
	jwtCli := fhir.NewJWTClient()

	_, err := jwtCli.ParseUnverified(_cb, "token")
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "failed to parse token")

	res, err := jwtCli.ParseUnverified(_cb, _accessToken)
	assert.NoError(t, err)
	assert.Equal(t, "admin-cli", res.RequestOrganizationID)
}
