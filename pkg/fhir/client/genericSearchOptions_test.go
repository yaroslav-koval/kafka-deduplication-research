package client_test

import (
	"kafka-polygon/pkg/fhir/client"
	"testing"

	"github.com/tj/assert"
)

func TestSortParamsOptionString(t *testing.T) {
	t.Parallel()

	sps := client.SortParamsOption{{Column: "0", Order: client.Desc}, {Column: "1", Order: client.Asc}, {Column: "2", Order: client.Desc}}

	assert.Equal(t, "-0,1,-2", sps.String())

	sps = nil

	assert.Equal(t, "", sps.String())
}

func TestSortParamsOptionFromString(t *testing.T) {
	t.Parallel()

	exp := client.SortParamsOption{
		{Column: "1", Order: client.Desc},
		{Column: "2", Order: client.Asc},
		{Column: "3", Order: client.Desc},
		{Column: "4", Order: client.Desc},
		{Column: "5", Order: client.Asc},
	}
	act := client.SortParamsOptionFromString("-1,2,-3,-4,5")

	assert.EqualValues(t, exp, act)

	exp = nil
	act = client.SortParamsOptionFromString("")

	assert.EqualValues(t, exp, act)
}
