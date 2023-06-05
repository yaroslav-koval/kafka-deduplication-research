package event

import (
	"context"
	"encoding/json"
	"kafka-polygon/pkg/cmd/metadata"
	"time"
)

type ConnectorEvent interface {
	BaseEvent
	GetInstance() interface{}
	GetTime() time.Time
}

type ConnectorData struct {
	ID           string      `json:"id"`
	Type         string      `json:"type"`
	Notification string      `json:"notification"`
	Instance     interface{} `json:"instance"`
	// Header set header field with context
	Header Header `json:"header"`
	// Time is the time at which the event was generated.
	Time time.Time `json:"time"`
	// Debug is set to true if this is a debugging event.
	Debug    bool
	Metadata metadata.Meta `json:"metadata"`
}

func (cd *ConnectorData) GetNotification() string {
	return cd.Notification
}

func (cd *ConnectorData) GetType() string {
	return cd.Type
}

func (cd *ConnectorData) GetTime() time.Time {
	return cd.Time
}

func (cd *ConnectorData) GetInstance() interface{} {
	return cd.Instance
}

func (cd *ConnectorData) GetID() string {
	return cd.ID
}

func (cd *ConnectorData) GetDebug() bool {
	return cd.Debug
}

func (cd *ConnectorData) ToByte() []byte {
	b, _ := json.Marshal(cd) //nolint:errchkjson

	return b
}

func (cd *ConnectorData) Unmarshal(msg Message) error {
	return json.Unmarshal(msg.Value, cd)
}

func (cd *ConnectorData) GetHeader() Header {
	return cd.Header
}

func (cd *ConnectorData) WithHeader(ctx context.Context) {
	cd.Header.XRequestIDFromContext(ctx)
}

func (cd *ConnectorData) GetMeta() metadata.Meta {
	return cd.Metadata
}

func (cd *ConnectorData) WithMeta(meta metadata.Meta) {
	cd.Metadata = meta
}
