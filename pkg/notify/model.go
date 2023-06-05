package notify

import (
	"time"
)

type MessageResponse struct {
	SubmittedAt        time.Time
	MessageID          string
	ErrorCode          int64
	Message            string
	ResponseStatusCode int
}

type MessageStatusRequest struct {
	ID    string
	MsgID string
}

type StatusResponse struct {
	Status  string
	Message string
}
