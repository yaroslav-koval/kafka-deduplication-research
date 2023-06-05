package event

import (
	"context"
	"encoding/json"
	"kafka-polygon/pkg/cmd/metadata"
)

// WorkflowEvent workflow event abstraction
type WorkflowEvent interface {
	BaseEvent
	GetWorkflow() Workflow
}

// WorkflowData implements WorkflowEvent
type WorkflowData struct {
	ID       string   `json:"id"`
	Header   Header   `json:"header"`
	Workflow Workflow `json:"workflow"`
	Debug    bool
	Metadata metadata.Meta `json:"metadata"`
}

// Workflow a workflow related data
type Workflow struct {
	ID     string `json:"id"`
	Schema string `json:"schema"`
	Step   string `json:"step"`
	// StepPayload is a portion of business data
	// that will be passed to workflow's step handler.
	// So it makes sense to passthrough it to the handler as a raw message.
	StepPayload json.RawMessage `json:"step_payload"`
}

func (w *WorkflowData) GetID() string {
	return w.ID
}

func (w *WorkflowData) GetWorkflow() Workflow {
	return w.Workflow
}

func (w *WorkflowData) GetDebug() bool {
	return w.Debug
}

func (w *WorkflowData) ToByte() []byte {
	b, err := json.Marshal(w)
	if err != nil {
		return nil
	}

	return b
}

func (w *WorkflowData) Unmarshal(msg Message) error {
	return json.Unmarshal(msg.Value, w)
}

func (w *WorkflowData) GetHeader() Header {
	return w.Header
}

func (w *WorkflowData) WithHeader(ctx context.Context) {
	w.Header.XRequestIDFromContext(ctx)
}

func (w *WorkflowData) GetMeta() metadata.Meta {
	return w.Metadata
}

func (w *WorkflowData) WithMeta(meta metadata.Meta) {
	w.Metadata = meta
}
