package event

import (
	"context"
	"kafka-polygon/pkg/cmd/metadata"
)

type Message struct {
	Key   *[]byte
	Value []byte
}

type Metadata interface {
	GetMeta() metadata.Meta
	WithMeta(meta metadata.Meta)
}

type BaseEvent interface {
	Metadata
	GetID() string
	GetDebug() bool
	WithHeader(ctx context.Context)
	GetHeader() Header
	ToByte() []byte
	Unmarshal(msg Message) error
}
