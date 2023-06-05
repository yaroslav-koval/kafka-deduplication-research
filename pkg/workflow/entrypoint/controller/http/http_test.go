package http_test

import (
	"context"
	"kafka-polygon/pkg/workflow/entrypoint/controller/http"
	"testing"

	"github.com/tj/assert"
)

var (
	_bgCtx = context.Background()
)

type mockAdapter struct {
	registerWorkflowRoutesFunc func(ctx context.Context, opts http.Option) error
}

func (m *mockAdapter) RegisterWorkflowRoutes(ctx context.Context, opts http.Option) error {
	return m.registerWorkflowRoutesFunc(ctx, opts)
}

func TestRegisterWorkflowRoutes(t *testing.T) {
	option := http.WithPrefix("123")
	adapter := &mockAdapter{
		registerWorkflowRoutesFunc: func(ctx context.Context, opts http.Option) error {
			assert.Equal(t, http.GetOptions(option), opts)

			return nil
		},
	}
	_ = http.RegisterWorkflowRoutes(_bgCtx, adapter, option)
}
