package entity_test

import (
	"kafka-polygon/pkg/workflow/entity"
	"testing"

	"github.com/tj/assert"
)

func TestWorkflowStatusString(t *testing.T) {
	s := "hello"
	assert.Equal(t, s, entity.WorkflowStatus(s).String())
}

func TestPointerWorkflowStatus(t *testing.T) {
	s := entity.WorkflowStatus("hello")
	assert.Equal(t, &s, entity.PointerWorkflowStatus(s.String()))
}

func TestWorkflowErrorMsgString(t *testing.T) {
	s := "hello"
	assert.Equal(t, s, entity.WorkflowErrorMsg(s).String())
}

func TestPointerWorkflowErrorMsg(t *testing.T) {
	s := entity.WorkflowErrorMsg("hello")
	assert.Equal(t, &s, entity.PointerWorkflowErrorMsg(s.String()))
}

func TestWorkflowErrorKindString(t *testing.T) {
	s := "hello"
	assert.Equal(t, s, entity.WorkflowErrorKind(s).String())
}

func TestPointerWorkflowErrorKind(t *testing.T) {
	s := entity.WorkflowErrorKind("hello")
	assert.Equal(t, &s, entity.PointerWorkflowErrorKind(s.String()))
}
