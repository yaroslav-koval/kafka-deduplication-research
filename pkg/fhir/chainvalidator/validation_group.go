package chainvalidator

import (
	"context"
	"kafka-polygon/pkg/cerror"
	"sync"
)

type ValidationGroup struct {
	validations []ValidationNode
	parallel    bool
	wg          sync.WaitGroup
	mutex       sync.Mutex
}

func NewGroup(comps ...ValidationNode) *ValidationGroup {
	return &ValidationGroup{
		validations: comps,
	}
}

func (g *ValidationGroup) Validate(ctx context.Context) error {
	if g.parallel {
		//nolint:cerrl
		mErr := cerror.NewMultiError(ctx)

		for _, validation := range g.validations {
			g.wg.Add(1)

			go func(c ValidationNode) {
				defer g.wg.Done()

				err := c.Validate(ctx)
				if err != nil {
					g.mutex.Lock()
					mErr.WithErrors(err)
					g.mutex.Unlock()
				}
			}(validation)
		}

		g.wg.Wait()

		if len(mErr.Errors()) > 0 {
			return mErr
		}

		return nil
	}

	for _, validation := range g.validations {
		err := validation.Validate(ctx)

		if err != nil {
			return err
		}
	}

	return nil
}

func (g *ValidationGroup) Add(c ValidationNode) {
	g.validations = append(g.validations, c)
}

func (g *ValidationGroup) EnableParallel() {
	g.parallel = true
}
