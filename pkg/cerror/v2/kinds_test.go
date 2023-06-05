package cerror_test

import (
	"kafka-polygon/pkg/cerror/v2"
	"testing"

	"github.com/tj/assert"
)

type expectedKind struct {
	kind       cerror.Kind
	kindString string
}

var expectedKindTable = []expectedKind{
	{cerror.KindValidation, "validation"},
	{cerror.KindUnauthenticated, "unauthenticated"},
	{cerror.KindForbidden, "forbidden"},
	{cerror.KindExists, "exists"},
	{cerror.KindNotExists, "not_exists"},
	{cerror.KindConflict, "conflict"},
	{cerror.KindSyntax, "syntax"},
	{cerror.KindExternalPkg, "external_pkg"},
	{cerror.KindIO, "io"},
	{cerror.KindTimeout, "timeout"},
	{cerror.KindExternalAPI, "external_api"},
	{cerror.KindInternalAPI, "internal_api"},
	{cerror.KindExternalResource, "external_resource"},
	{cerror.KindInternalResource, "internal_resource"},
}

func TestKindString(t *testing.T) {
	for _, v := range expectedKindTable {
		assert.Equal(t, v.kindString, v.kind.String())
	}
}
