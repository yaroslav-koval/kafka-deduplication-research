package entity

import (
	"context"
	"encoding/json"
	"kafka-polygon/pkg/broker/event"
)

type WorkflowSchemaStepName string

func (w WorkflowSchemaStepName) String() string {
	return string(w)
}

func PointerWorkflowSchemaStepName(s string) *WorkflowSchemaStepName {
	sn := WorkflowSchemaStepName(s)
	return &sn
}

type WorkflowSchemaStepTopic string

func (w WorkflowSchemaStepTopic) String() string {
	return string(w)
}

func PointerWorkflowSchemaStepTopic(s string) *WorkflowSchemaStepTopic {
	st := WorkflowSchemaStepTopic(s)
	return &st
}

// WorkflowSchemaStepWorker is a business logic unit abstraction.
type WorkflowSchemaStepWorker interface {
	// Implementor will recieve a workflow event.
	// Then it should handle it according to a business logic and
	// return a result for the next step as event.Payload.
	Run(ctx context.Context, e event.WorkflowEvent) (json.RawMessage, error)
}

// WorkflowSchemaStep is a workflow's schema step abstraction.
type WorkflowSchemaStep interface {
	// Unique step name
	Name() WorkflowSchemaStepName
	// Message queue topic name with events for the step
	Topic() WorkflowSchemaStepTopic
	// Business logic handler
	Worker() WorkflowSchemaStepWorker
}
