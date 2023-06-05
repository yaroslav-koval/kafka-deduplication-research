package entity

import (
	"encoding/json"
	"time"

	"github.com/uptrace/bun"
)

const (
	WorkflowHistoryTypeRestart     WorkflowHistoryType = "RESTART"
	WorkflowHistoryTypeRestartFrom WorkflowHistoryType = "RESTART_FROM"
)

type WorkflowHistoryType string

func (w WorkflowHistoryType) String() string {
	return string(w)
}

func PointerWorkflowHistoryType(s string) *WorkflowHistoryType {
	wh := WorkflowHistoryType(s)
	return &wh
}

type WorkflowHistory struct {
	bun.BaseModel     `bun:"table:workflow_history"`
	ID                ID                     `bson:"_id" json:"id" bun:"id,pk"`
	CreatedAt         time.Time              `bson:"created_at,omitempty" json:"created_at,omitempty" bun:"created_at"`
	Type              WorkflowHistoryType    `bson:"type" json:"type" bun:"type"`
	Input             json.RawMessage        `bson:"input" json:"input" bun:"input,type:jsonb"`
	InputPrevious     json.RawMessage        `bson:"input_previous" json:"input_previous" bun:"input_previous,type:jsonb"`
	StepName          WorkflowSchemaStepName `bson:"step_name" json:"step_name" bun:"step_name"`
	WorkflowID        ID                     `bson:"workflow_id" json:"workflow_id" bun:"workflow_id"`
	WorkflowStatus    WorkflowStatus         `bson:"workflow_status" json:"workflow_status" bun:"workflow_status"`
	WorkflowError     *WorkflowErrorMsg      `bson:"workflow_error" json:"workflow_error" bun:"workflow_error"`
	WorkflowErrorKind *WorkflowErrorKind     `bson:"workflow_error_kind" json:"workflow_error_kind" bun:"workflow_error_kind"`
	RequestID         *string                `bson:"request_id" json:"request_id" bun:"request_id"`
}
