package http_test

import (
	"kafka-polygon/pkg/workflow/entrypoint/controller/http"
	"testing"

	"github.com/tj/assert"
)

func TestOptions(t *testing.T) {
	t.Parallel()

	opts := http.GetOptions()
	assert.Equal(t, http.DefaultWorkflowRoutePrefix, opts.GetPrefix())

	opts = http.GetOptions(http.WithPrefix("/test"))
	assert.Equal(t, "/test", opts.GetPrefix())
}
