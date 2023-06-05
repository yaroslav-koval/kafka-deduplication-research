package errtransformer

import "context"

type ErrorTransformer interface {
	Transform(ctx context.Context, err error) error
}

type TransformersChain struct {
	transformers []ErrorTransformer
}

func NewChain(c []ErrorTransformer) *TransformersChain {
	return &TransformersChain{
		transformers: c,
	}
}

func (tc *TransformersChain) Transform(ctx context.Context, err error) error {
	if err == nil {
		return nil
	}

	for _, c := range tc.transformers {
		if newErr := c.Transform(ctx, err); newErr != err {
			return newErr
		}
	}

	return err
}
