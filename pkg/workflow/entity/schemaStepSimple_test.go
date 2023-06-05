package entity_test

import (
	"kafka-polygon/pkg/workflow/entity"
	"testing"

	"github.com/tj/assert"
)

func TestNewWorkflowSchemaSimpleStep(t *testing.T) {
	name := entity.WorkflowSchemaStepName("step1")
	topic := entity.WorkflowSchemaStepTopic("topic1")
	worker := new(stepWorkerTest)

	actStep := entity.NewWorkflowSchemaSimpleStep(name, topic, worker)

	assert.NotNil(t, actStep)
	assert.Equal(t, name, actStep.Name())
	assert.Equal(t, topic, actStep.Topic())
	assert.Equal(t, worker, actStep.Worker())
}
