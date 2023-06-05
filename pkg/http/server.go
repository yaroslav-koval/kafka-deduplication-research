package http

import "context"

type Server interface {
	ListenAndServe(ctx context.Context) chan error
	Shutdown(ctx context.Context) error
}
