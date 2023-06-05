package store

import "context"

const (
	EventStatusNew              = "new"
	EventStatusHandled          = "handled"
	EventStatusHandledWithError = "handled_with_error"
)

type EventProcessData struct {
	Status string `bson:"status" json:"status"`
}

type Store interface {
	GetEventInfoByID(ctx context.Context, id string) (EventProcessData, error)
	PutEventInfo(ctx context.Context, id string, data EventProcessData) error
}
