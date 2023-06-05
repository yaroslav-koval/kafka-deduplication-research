package entity

import (
	"encoding/json"
	"time"

	"github.com/uptrace/bun"
)

const (
	WorkflowStatusInProgress WorkflowStatus = "IN_PROGRESS"
	WorkflowStatusFailed     WorkflowStatus = "FAILED"
	WorkflowStatusSuccess    WorkflowStatus = "SUCCESS"
)

type WorkflowStatus string

func (w WorkflowStatus) String() string {
	return string(w)
}

func PointerWorkflowStatus(s string) *WorkflowStatus {
	ss := WorkflowStatus(s)
	return &ss
}

type WorkflowErrorMsg string

func (w WorkflowErrorMsg) String() string {
	return string(w)
}

func PointerWorkflowErrorMsg(s string) *WorkflowErrorMsg {
	e := WorkflowErrorMsg(s)
	return &e
}

type WorkflowErrorKind string

func (w WorkflowErrorKind) String() string {
	return string(w)
}

func PointerWorkflowErrorKind(s string) *WorkflowErrorKind {
	e := WorkflowErrorKind(s)
	return &e
}

type Workflow struct {
	bun.BaseModel `bun:"table:workflow"`
	ID            ID                 `bson:"_id" json:"id" bun:"id,pk"`
	CreatedAt     time.Time          `bson:"created_at,omitempty" json:"created_at,omitempty" bun:"created_at"`
	UpdatedAt     time.Time          `bson:"updated_at,omitempty" json:"updated_at,omitempty" bun:"updated_at"`
	ParentID      *ID                `bson:"parent_id" json:"parent_id" bun:"parent_id"`
	SchemaName    WorkflowSchemaName `bson:"schema_name" json:"schema_name" bun:"schema_name"`
	Status        WorkflowStatus     `bson:"status" json:"status" bun:"status"`
	Input         json.RawMessage    `bson:"input" json:"input" bun:"input,type:jsonb"`
	Steps         []*WorkflowStep    `bson:"steps" json:"steps,omitempty" bun:"steps"`
	Error         *WorkflowErrorMsg  `bson:"error" json:"error" bun:"error"`
	ErrorKind     *WorkflowErrorKind `bson:"error_kind" json:"error_kind" bun:"error_kind"`
	RequestID     *string            `bson:"request_id" json:"request_id" bun:"request_id"`
}

type WorkflowStep struct {
	CreatedAt time.Time              `bson:"created_at,omitempty" json:"created_at,omitempty"`
	Name      WorkflowSchemaStepName `bson:"name" json:"name"`
	Data      json.RawMessage        `bson:"data" json:"data"`
	Metadata  WorkflowStepMetadata   `bson:"metadata" json:"metadata"`
}

type WorkflowStepMetadata struct {
	Version string `bson:"version" json:"version"`
}
