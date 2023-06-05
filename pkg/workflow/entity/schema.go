package entity

import (
	"context"
	"fmt"
	"kafka-polygon/pkg/cerror"
)

type WorkflowSchemaName string

func (w WorkflowSchemaName) String() string {
	return string(w)
}

func PointerWorkflowSchemaName(s string) *WorkflowSchemaName {
	sn := WorkflowSchemaName(s)
	return &sn
}

// WorkflowSchema represents a sequence of logical units(steps)
type WorkflowSchema struct {
	name  WorkflowSchemaName
	steps []WorkflowSchemaStep
}

// NewWorkflowSchema creates a new workflow schema with given name and set of steps.
// Execute steps validation and returns error if something is wrong.
// It is the only way to initialize workflow with steps, so later all the steps
// can be considered to be valid (e.g. at least one step, no duplicates, ...)
func NewWorkflowSchema(
	ctx context.Context, name WorkflowSchemaName, steps ...WorkflowSchemaStep) (*WorkflowSchema, error) {
	w := &WorkflowSchema{name: name, steps: steps}

	if err := w.validate(ctx); err != nil {
		return nil, err
	}

	return w, nil
}

func (w *WorkflowSchema) Name() WorkflowSchemaName {
	return w.name
}

// Step finds workflow's schema step by name.
// If step is not found, the second returned parameter will be nil.
func (w *WorkflowSchema) Step(sn WorkflowSchemaStepName) (WorkflowSchemaStep, bool) {
	for _, s := range w.steps {
		if s.Name() == sn {
			return s, true
		}
	}

	return nil, false
}

// NextStep returns the next step after step with given name.
// The second returned parameter indicates whether the next step exists.
func (w *WorkflowSchema) NextStep(current WorkflowSchemaStepName) (WorkflowSchemaStep, bool) {
	for i := range w.steps {
		if w.steps[i].Name() == current {
			if i < len(w.steps)-1 {
				return w.steps[i+1], true
			}

			return nil, false
		}
	}

	return nil, false
}

func (w *WorkflowSchema) FirstStep() WorkflowSchemaStep {
	return w.steps[0]
}

// Validate checks whether the current workflow's schema is valid
func (w *WorkflowSchema) validate(ctx context.Context) error {
	errs := make(map[string]string)

	if w.name == "" {
		errs["name"] = "workflow name is empty"
	}

	if len(w.steps) == 0 {
		errs["steps"] = "workflow has no steps"
	}

	for i := range w.steps {
		s := w.steps[i]

		if s.Name() == "" {
			errs[fmt.Sprintf("steps[%d].name", i)] = "step name is empty"
		} else {
			for iDupl := range w.steps {
				if iDupl == i {
					continue
				}

				if w.steps[iDupl].Name() == s.Name() {
					errs[fmt.Sprintf("steps[%d].name", i)] = fmt.Sprintf(
						"%s step has duplicate by name(dupl step index %d)", s.Name(), iDupl)
					break
				}
			}
		}

		if s.Topic() == "" {
			errs[fmt.Sprintf("steps[%d].topic", i)] = "step topic is empty"
		} else {
			for iDupl := range w.steps {
				if iDupl == i {
					continue
				}

				if w.steps[iDupl].Topic() == s.Topic() {
					errs[fmt.Sprintf("steps[%d].topic", i)] = fmt.Sprintf(
						"%s step has duplicate by topic(dupl step index %d)", s.Name(), iDupl)
					break
				}
			}
		}

		if s.Worker() == nil {
			errs[fmt.Sprintf("steps[%d].worker", i)] = "step worker is empty"
		}
	}

	if len(errs) > 0 {
		return cerror.NewValidationError(ctx, errs).LogError()
	}

	return nil
}
