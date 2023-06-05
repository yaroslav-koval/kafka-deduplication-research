package test

import (
	"encoding/json"
	"kafka-polygon/pkg/fhir/cache"
	"testing"

	"github.com/edenlabllc/go-fhir-adapter/model/fhir"
	"github.com/tj/assert"
)

func TestErrNoSessionKeyInContext(t *testing.T) {
	t.Parallel()

	assert.Equal(t, errNoSessionKeyInCtx.Error(), cache.ErrNoSessionKeyInContext.Error())
}

func TestContextSessionKeyName(t *testing.T) {
	t.Parallel()

	assert.Equal(t, "X-Request-ID", cache.ContextSessionKeyName)
}

func TestKeyFromID(t *testing.T) {
	t.Parallel()

	id := "123"
	assert.Equal(t, id, cache.KeyFromID(fhir.ID(id)))
}

func TestKeyFromStruct(t *testing.T) {
	t.Parallel()

	s := struct {
		S  string
		I  int
		F  float64
		B  bool
		Bt []byte
	}{"s", 1, 2.2, true, []byte("byte")}
	b, _ := json.Marshal(s)
	assert.Equal(t, string(b), cache.KeyFromStruct(s))
}

func TestSessionKeyFromCtx(t *testing.T) {
	t.Parallel()

	_, ok := cache.SessionKeyFromCtx(_cb)
	assert.False(t, ok)

	v := "123"
	actualV, ok := cache.SessionKeyFromCtx(ctxWithReqID(_cb, v))
	assert.True(t, ok)
	assert.Equal(t, v, actualV)
}
