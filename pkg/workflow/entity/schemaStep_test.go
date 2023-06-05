package entity_test

import (
	"kafka-polygon/pkg/workflow/entity"
	"testing"

	"github.com/tj/assert"
)

func TestWorkflowSchemaStepNameString(t *testing.T) {
	s := "hello"
	assert.Equal(t, s, entity.WorkflowSchemaStepName(s).String())
}

func TestPointerWorkflowSchemaStepName(t *testing.T) {
	s := entity.WorkflowSchemaStepName("hello")
	assert.Equal(t, &s, entity.PointerWorkflowSchemaStepName(s.String()))
}

func TestWorkflowSchemaStepTopicString(t *testing.T) {
	s := "hello"
	assert.Equal(t, s, entity.WorkflowSchemaStepTopic(s).String())
}

func TestPointerWorkflowSchemaStepTopic(t *testing.T) {
	s := entity.WorkflowSchemaStepTopic("hello")
	assert.Equal(t, &s, entity.PointerWorkflowSchemaStepTopic(s.String()))
}
