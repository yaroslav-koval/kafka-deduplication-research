package cache

import (
	"context"
	"encoding/json"
	"fmt"
	"kafka-polygon/pkg/http/consts"

	"github.com/edenlabllc/go-fhir-adapter/model/fhir"
)

// Key serves as a cache key.
type Key = string

// ErrNoSessionKeyInContext is an error that is returned if context has no session key.
var ErrNoSessionKeyInContext = fmt.Errorf("no session key in context")

// ContextSessionKeyName is the name of the key that holds session id value in context.
const ContextSessionKeyName = consts.HeaderXRequestID

// KeyFromID transforms id to cache key.
func KeyFromID(id fhir.ID) Key {
	return id.String()
}

// KeyFromStruct transforms struct to cache key by marshaling it to json.
func KeyFromStruct(s interface{}) Key {
	b, _ := json.Marshal(s) //nolint:errchkjson
	return string(b)
}

// SessionKeyFromCtx extracts a session key value from context.
func SessionKeyFromCtx(ctx context.Context) (Key, bool) {
	if s, ok := ctx.Value(ContextSessionKeyName).(string); ok && s != "" {
		return s, true
	}

	return "", false
}
