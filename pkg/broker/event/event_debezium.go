package event

import (
	"context"
	"encoding/json"
	"kafka-polygon/pkg/cerror"
	"kafka-polygon/pkg/cmd/metadata"
)

type DebeziumEvent interface {
	BaseEvent
	WithContext(context.Context)
	GetPayload() *DebeziumPayload
}

type DebeziumData struct {
	ctx     context.Context
	Key     *KeyData
	Payload *DebeziumPayload `json:"payload"`
	// Header set header field with context
	Header Header `json:"header"`
	// Debug is set to true if this is a debugging event.
	Debug    bool
	Metadata metadata.Meta `json:"metadata"`
}

type KeyData struct {
	Payload KeyPayload `json:"payload"`
}

type KeyPayload struct {
	ID string `json:"id"`
}

type DebeziumPayload struct {
	Source       DebeziumSource         `json:"source"`
	BeforeValues map[string]interface{} `json:"before"`
	AfterValues  map[string]interface{} `json:"after"`
	Op           string                 `json:"op"`
}

type DebeziumSource struct {
	Table string `json:"table"`
}

func NewDebeziumData() DebeziumEvent {
	return &DebeziumData{
		ctx:      context.Background(),
		Key:      &KeyData{},
		Metadata: metadata.New(),
	}
}

func (dd *DebeziumData) WithContext(ctx context.Context) {
	dd.ctx = ctx
}

// GetID of DebeziumData
func (dd *DebeziumData) GetID() string {
	if dd.Key != nil {
		return dd.Key.Payload.ID
	}

	return ""
}

// ToByte of DebeziumData
func (dd *DebeziumData) ToByte() []byte {
	b, err := json.Marshal(dd)
	if err != nil {
		_ = cerror.New(dd.ctx, cerror.KindInternal, err).LogError()
		return nil
	}

	return b
}

// Unmarshal of DebeziumData
func (dd *DebeziumData) Unmarshal(msg Message) error {
	err := json.Unmarshal(msg.Value, dd)
	if err != nil {
		return cerror.New(dd.ctx, cerror.KindInternal, err).LogError()
	}

	if msg.Key != nil {
		err = json.Unmarshal(*msg.Key, dd.Key)
		if err != nil {
			return cerror.New(dd.ctx, cerror.KindInternal, err).LogError()
		}
	}

	return nil
}

// GetPayload of DebeziumData
func (dd *DebeziumData) GetPayload() *DebeziumPayload {
	return dd.Payload
}

// GetDebug of DebeziumData
func (dd *DebeziumData) GetDebug() bool {
	return dd.Debug
}

func (dd *DebeziumData) GetHeader() Header {
	return dd.Header
}

func (dd *DebeziumData) WithHeader(ctx context.Context) {
	dd.Header.XRequestIDFromContext(ctx)
}

func (dd *DebeziumData) GetMeta() metadata.Meta {
	return dd.Metadata
}

func (dd *DebeziumData) WithMeta(meta metadata.Meta) {
	dd.Metadata = meta
}
