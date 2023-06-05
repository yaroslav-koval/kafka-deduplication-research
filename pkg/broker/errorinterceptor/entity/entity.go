package entity

import (
	"encoding/json"
	"time"
)

type FailedBrokerEvent struct {
	ID        string
	CreatedAt time.Time
	UpdatedAt time.Time
	Status    string
	Topic     string
	Data      json.RawMessage
	Error     string
}
