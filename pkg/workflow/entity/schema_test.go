package entity_test

import (
	"kafka-polygon/pkg/workflow/entity"
	"testing"

	"github.com/tj/assert"
)

func TestNewWorkflowSchema(t *testing.T) {
	schemaName := entity.WorkflowSchemaName("")

	var steps []entity.WorkflowSchemaStep

	// empty name and steps
	_, actErr := entity.NewWorkflowSchema(_bgCtx, schemaName, steps...)
	assert.Error(t, actErr)
	assertMultiValidationError(t, map[string]string{
		"name":  "workflow name is empty",
		"steps": "workflow has no steps",
	}, actErr)

	// step without name, topic and worker
	schemaName = entity.WorkflowSchemaName("schema1")
	steps = []entity.WorkflowSchemaStep{entity.NewWorkflowSchemaSimpleStep("", "", nil)}
	_, actErr = entity.NewWorkflowSchema(_bgCtx, schemaName, steps...)
	assert.Error(t, actErr)
	assertMultiValidationError(t, map[string]string{
		"steps[0].name":   "step name is empty",
		"steps[0].topic":  "step topic is empty",
		"steps[0].worker": "step worker is empty",
	}, actErr)

	// steps with duplicate names and topics
	steps = []entity.WorkflowSchemaStep{
		entity.NewWorkflowSchemaSimpleStep("step1", "topic1", new(stepWorkerTest)),
		entity.NewWorkflowSchemaSimpleStep("step1", "topic1", new(stepWorkerTest)),
	}
	_, actErr = entity.NewWorkflowSchema(_bgCtx, schemaName, steps...)
	assert.Error(t, actErr)
	assertMultiValidationError(t, map[string]string{
		"steps[0].name":  "step1 step has duplicate by name(dupl step index 1)",
		"steps[0].topic": "step1 step has duplicate by topic(dupl step index 1)",
		"steps[1].name":  "step1 step has duplicate by name(dupl step index 0)",
		"steps[1].topic": "step1 step has duplicate by topic(dupl step index 0)",
	}, actErr)

	// no error
	steps = []entity.WorkflowSchemaStep{
		entity.NewWorkflowSchemaSimpleStep("step1", "topic1", new(stepWorkerTest)),
		entity.NewWorkflowSchemaSimpleStep("step2", "topic2", new(stepWorkerTest)),
	}
	actSchema, actErr := entity.NewWorkflowSchema(_bgCtx, schemaName, steps...)
	assert.NoError(t, actErr)
	assert.NotEmpty(t, actSchema)
	assert.Equal(t, schemaName, actSchema.Name())
}

func TestNewWorkflowSchemaStep(t *testing.T) {
	steps := []entity.WorkflowSchemaStep{
		entity.NewWorkflowSchemaSimpleStep("step1", "topic1", new(stepWorkerTest)),
		entity.NewWorkflowSchemaSimpleStep("step2", "topic2", new(stepWorkerTest)),
		entity.NewWorkflowSchemaSimpleStep("step3", "topic3", new(stepWorkerTest)),
	}
	schema, err := entity.NewWorkflowSchema(_bgCtx, "schema1", steps...)
	assert.NoError(t, err)

	for _, s := range steps {
		actStep, ok := schema.Step(s.Name())
		assert.True(t, ok)
		assert.Equal(t, s, actStep)

		_, ok = schema.Step(entity.WorkflowSchemaStepName(s.Name().String()) + "?")
		assert.False(t, ok)
	}
}

func TestNewWorkflowSchemaNextStep(t *testing.T) {
	steps := []entity.WorkflowSchemaStep{
		entity.NewWorkflowSchemaSimpleStep("step1", "topic1", new(stepWorkerTest)),
		entity.NewWorkflowSchemaSimpleStep("step2", "topic2", new(stepWorkerTest)),
		entity.NewWorkflowSchemaSimpleStep("step3", "topic3", new(stepWorkerTest)),
	}
	schema, err := entity.NewWorkflowSchema(_bgCtx, "schema1", steps...)
	assert.NoError(t, err)

	for i, s := range steps {
		actNext, ok := schema.NextStep(s.Name())
		if i == len(steps)-1 {
			assert.False(t, ok)
			assert.Nil(t, actNext)
		} else {
			assert.True(t, ok)
			assert.Equal(t, steps[i+1], actNext)
		}

		_, ok = schema.NextStep(entity.WorkflowSchemaStepName(s.Name().String()) + "?")
		assert.False(t, ok)
	}
}

func TestNewWorkflowSchemaFirstStep(t *testing.T) {
	steps := []entity.WorkflowSchemaStep{
		entity.NewWorkflowSchemaSimpleStep("step1", "topic1", new(stepWorkerTest)),
		entity.NewWorkflowSchemaSimpleStep("step2", "topic2", new(stepWorkerTest)),
	}
	schema, err := entity.NewWorkflowSchema(_bgCtx, "schema1", steps...)
	assert.NoError(t, err)
	assert.Equal(t, steps[0], schema.FirstStep())
}

func TestNewWorkflowSchemaName(t *testing.T) {
	expName := entity.WorkflowSchemaName("schema1")
	steps := []entity.WorkflowSchemaStep{
		entity.NewWorkflowSchemaSimpleStep("step1", "topic1", new(stepWorkerTest)),
	}
	schema, err := entity.NewWorkflowSchema(_bgCtx, expName, steps...)
	assert.NoError(t, err)
	assert.Equal(t, expName, schema.Name())
}

func TestWorkflowSchemaNameString(t *testing.T) {
	s := "hello"
	assert.Equal(t, s, entity.WorkflowSchemaName(s).String())
}

func TestPointerWorkflowSchemaName(t *testing.T) {
	s := entity.WorkflowSchemaName("hello")
	assert.Equal(t, &s, entity.PointerWorkflowSchemaName(s.String()))
}
