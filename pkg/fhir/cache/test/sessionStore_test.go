package test

import (
	"context"
	"fmt"
	"kafka-polygon/pkg/fhir/cache"
	"sync"
	"testing"

	"github.com/tj/assert"
)

//nolint:dupl
func TestInitSession(t *testing.T) {
	t.Parallel()

	type store struct{}

	createStoreFn := func() *store { return new(store) }

	s := cache.NewSessionStore(createStoreFn)

	// must return error if there is no session key in context
	err := s.InitSession(_cb)
	assert.Error(t, err)
	assertErrorEqual(t, errNoSessionKeyInCtx, err)

	// must return error if there is session key with empty value in context
	err = s.InitSession(ctxWithReqID(_cb, ""))
	assert.Error(t, err)
	assertErrorEqual(t, errNoSessionKeyInCtx, err)

	// must work with session key in context
	err = s.InitSession(ctxWithReqID(_cb, "123"))
	assert.NoError(t, err)
}

func TestSession(t *testing.T) {
	t.Parallel()

	type store struct{}

	expStore := new(store)

	createStoreFn := func() *store { return expStore }

	s := cache.NewSessionStore(createStoreFn)

	// must return false if there is no session key in context
	_, ok := s.Session(_cb)
	assert.False(t, ok)

	ctxWithKey := ctxWithReqID(_cb, "123")

	// must return false if session was not inited
	_, ok = s.Session(ctxWithKey)
	assert.False(t, ok)

	// must work with inited session
	err := s.InitSession(ctxWithKey)
	assert.NoError(t, err)

	actStore, ok := s.Session(ctxWithKey)

	assert.True(t, ok)
	assert.Equal(t, expStore, actStore)
}

//nolint:dupl
func TestClearSession(t *testing.T) {
	t.Parallel()

	type store struct{}

	createStoreFn := func() *store { return new(store) }

	s := cache.NewSessionStore(createStoreFn)

	// must return error if there is no session key in context
	err := s.ClearSession(_cb)
	assert.Error(t, err)
	assertErrorEqual(t, errNoSessionKeyInCtx, err)

	// must return error if there is session key with empty value in context
	err = s.ClearSession(ctxWithReqID(_cb, ""))
	assert.Error(t, err)
	assertErrorEqual(t, errNoSessionKeyInCtx, err)

	// must work with session key in context
	err = s.ClearSession(ctxWithReqID(_cb, "123"))
	assert.NoError(t, err)
}

func TestConcurrentSessions(t *testing.T) {
	type store struct{}

	createStoreFn := func() *store { return new(store) }

	s := cache.NewSessionStore(createStoreFn)

	wg := sync.WaitGroup{}
	groupCount := 10000

	wg.Add(groupCount)

	for i := 0; i < groupCount; i++ {
		go func(k string) {
			defer wg.Done()

			_ = s.InitSession(ctxWithReqID(_cb, k))
		}(fmt.Sprintf("%d", i))
	}

	wg.Add(groupCount)

	for i := 0; i < groupCount; i++ {
		go func(k string) {
			defer wg.Done()

			_, _ = s.Session(ctxWithReqID(_cb, k))
		}(fmt.Sprintf("%d", i))
	}

	wg.Add(groupCount)

	for i := 0; i < groupCount; i++ {
		go func(k string) {
			defer wg.Done()

			_ = s.ClearSession(ctxWithReqID(_cb, k))
		}(fmt.Sprintf("%d", i))
	}

	wg.Wait()
}

func TestCallWithSessionCache(t *testing.T) {
	type callResultStore struct {
		results *cache.Store[string]
	}

	createcallResultStoreFn := func() *callResultStore {
		return &callResultStore{results: cache.NewStore[string]()}
	}
	s := cache.NewSessionStore(createcallResultStoreFn)
	ctx := ctxWithReqID(_cb, "123")
	cacheSelector := func(s *callResultStore) *cache.Store[string] {
		return s.results
	}
	isOriginalFnCalled := false
	originalFn := func() (string, error) {
		isOriginalFnCalled = true
		return "", nil
	}

	_ = s.InitSession(ctx)

	inputs := []struct {
		key          string
		isDuplicated bool
	}{
		{"1", false},
		{"2", false},
		{"1", true},
	}

	for _, i := range inputs {
		if !i.isDuplicated {
			// must call original fn for the first search with given parameters
			isOriginalFnCalled = false
			_, err := cache.CallWithSessionCache(ctx, s, originalFn, i.key, cacheSelector)
			assert.NoError(t, err)
			assert.True(t, isOriginalFnCalled)
		}

		// must return from cache and not call original fn for the second call with the same parameters
		isOriginalFnCalled = false
		_, err := cache.CallWithSessionCache(ctx, s, originalFn, i.key, cacheSelector)
		assert.NoError(t, err)
		assert.False(t, isOriginalFnCalled)
	}

	// must call original fn if the same parameters were given but in other session
	isOriginalFnCalled = false
	newCtx := ctxWithReqID(_cb, "-123")
	_ = s.InitSession(newCtx)
	_, err := cache.CallWithSessionCache(newCtx, s, originalFn, inputs[0].key, cacheSelector)
	assert.NoError(t, err)
	assert.True(t, isOriginalFnCalled)

	// must always call original fn if there is no session or ctx don't have session key
	for _, ctxWithNoSession := range []context.Context{
		_cb,
		ctxWithReqID(_cb, "-999"),
	} {
		for i := 0; i < 2; i++ {
			isOriginalFnCalled = false
			_, err := cache.CallWithSessionCache(ctxWithNoSession, s, originalFn, inputs[0].key, cacheSelector)
			assert.NoError(t, err)
			assert.True(t, isOriginalFnCalled)
		}
	}
}
