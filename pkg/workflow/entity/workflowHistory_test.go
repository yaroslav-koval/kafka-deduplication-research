package entity_test

import (
	"kafka-polygon/pkg/workflow/entity"
	"testing"

	"github.com/tj/assert"
)

func TestWorkflowHistoryTypeString(t *testing.T) {
	s := "hello"
	assert.Equal(t, s, entity.WorkflowHistoryType(s).String())
}

func TestPointerWorkflowHistoryType(t *testing.T) {
	s := entity.WorkflowHistoryType("hello")
	assert.Equal(t, &s, entity.PointerWorkflowHistoryType(s.String()))
}
