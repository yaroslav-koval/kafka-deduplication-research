package event

import (
	"context"
	"encoding/json"
	"kafka-polygon/pkg/cmd/metadata"
	"time"
)

type NotificationEvent interface {
	BaseEvent
	GetWorkflowID() string
	GetChannel() string
	GetProvider() string
	GetTime() time.Time
}

type NotificationData struct {
	ID         string `json:"id"`
	WorkflowID string `bson:"workflow_id" json:"workflow_id"`
	Channel    string `bson:"channel" json:"channel"`
	Provider   string `bson:"provider" json:"provider"`
	// Header set header field with context
	Header Header `json:"header"`
	// Time is the time at which the event was generated.
	Time time.Time `bson:"time" json:"time"`
	// Debug is set to true if this is a debugging event.
	Debug    bool
	Metadata metadata.Meta `json:"metadata"`
}

// GetID of NotificationData
func (nd *NotificationData) GetID() string {
	return nd.ID
}

// GetWorkflowID of NotificationData
func (nd *NotificationData) GetWorkflowID() string {
	return nd.WorkflowID
}

// GetChannel of NotificationData
func (nd *NotificationData) GetChannel() string {
	return nd.Channel
}

// GetProvider of NotificationData
func (nd *NotificationData) GetProvider() string {
	return nd.Provider
}

// GetTime of NotificationData
func (nd *NotificationData) GetTime() time.Time {
	return nd.Time
}

// GetDebug of NotificationData
func (nd *NotificationData) GetDebug() bool {
	return nd.Debug
}

// ToByte of NotificationData
func (nd *NotificationData) ToByte() []byte {
	b, err := json.Marshal(nd)
	if err != nil {
		return nil
	}

	return b
}

// Unmarshal of NotificationData
func (nd *NotificationData) Unmarshal(msg Message) error {
	return json.Unmarshal(msg.Value, nd)
}

func (nd *NotificationData) GetHeader() Header {
	return nd.Header
}

func (nd *NotificationData) WithHeader(ctx context.Context) {
	nd.Header.XRequestIDFromContext(ctx)
}

func (nd *NotificationData) GetMeta() metadata.Meta {
	return nd.Metadata
}

func (nd *NotificationData) WithMeta(meta metadata.Meta) {
	nd.Metadata = meta
}
