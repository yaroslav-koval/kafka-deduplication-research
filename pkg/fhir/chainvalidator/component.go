package chainvalidator

import "context"

type ValidationNode interface {
	Validate(ctx context.Context) error
}
