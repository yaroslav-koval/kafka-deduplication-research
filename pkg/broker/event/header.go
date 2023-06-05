package event

import (
	"context"
	"fmt"
	"kafka-polygon/pkg/http/consts"
)

type Header struct {
	RequestID string `json:"request_id"`
}

func (h *Header) XRequestIDFromContext(c context.Context) {
	reqID := c.Value(consts.HeaderXRequestID)

	if reqID != nil {
		switch v := reqID.(type) {
		case fmt.Stringer:
			h.RequestID = v.String()
		default:
			h.RequestID = fmt.Sprintf("%+v", v)
		}
	}
}
