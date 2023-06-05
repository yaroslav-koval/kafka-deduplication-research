package cache

import (
	"context"
	"kafka-polygon/pkg/cerror"
	"sync"
)

// SessionStore stores cache bounded to session key.
// Actually it can store any data bounded to session key, but
// it was mostly designed to store objects that have fields
// of type Store[T]. It is the most useful for using in
// CallWithSessionCache method where TCacheStore will be passed to
// CacheSelector function. For example, we want to cache results
// of some functions that return Tasks and Patients:
//
//	type MyResultStore struct {
//		Tasks    *Store[*fhir.Task]
//		Patients *Store[*fhir.Patient]
//	}
//
//	type MySessionStore struct {
//		*SessionStore[SomeResultStore]
//	}
//
// InitSession and ClearSession must be called by the layer
// that is responsible for the logic flow and knows when it
// starts and ends (e.g. usecase layer).
type SessionStore[TCacheStore interface{}] struct {
	m               sync.Mutex
	sessions        map[Key]*TCacheStore
	newCacheStoreFn func() *TCacheStore
}

func NewSessionStore[TCacheStore interface{}](newCacheStoreFn func() *TCacheStore) *SessionStore[TCacheStore] {
	return &SessionStore[TCacheStore]{
		sessions:        make(map[Key]*TCacheStore),
		newCacheStoreFn: newCacheStoreFn,
	}
}

// InitSession starts a new cache session for based on session id value
// found in a given context. If there is no session key in context
// that error is returned. Look at ContextSessionKeyName constant
// to know what key is assumed to store session id value.
// Be careful and always use defer ClearSession after InitSession.
// Must be called by the layer that is responsible for the logic flow
// and knows when it starts and ends (e.g. usecase layer).
func (s *SessionStore[TCache]) InitSession(ctx context.Context) error {
	key, ok := SessionKeyFromCtx(ctx)
	if !ok {
		return cerror.New(ctx, cerror.KindInternal, ErrNoSessionKeyInContext).LogError()
	}

	s.m.Lock()
	defer s.m.Unlock()
	s.sessions[key] = s.newCacheStoreFn()

	return nil
}

// Session returns cache session based on a given context.
// If there is no session by given key or there is no session key
// in context, the second returned value will be false.
func (s *SessionStore[TCache]) Session(ctx context.Context) (*TCache, bool) {
	k, ok := SessionKeyFromCtx(ctx)
	if !ok {
		return nil, false
	}

	s.m.Lock()
	defer s.m.Unlock()
	v, ok := s.sessions[k]

	return v, ok
}

// ClearSession removes session based on a given context.
// If there is no session by given key or there is no session key
// in context, the second returned value will be false.
// Basically must be called by the layer that initiated the session
// and knows when the logic flow ends (e.g. usecase layer).
func (s *SessionStore[TCache]) ClearSession(ctx context.Context) error {
	key, ok := SessionKeyFromCtx(ctx)
	if !ok {
		return cerror.New(ctx, cerror.KindInternal, ErrNoSessionKeyInContext).LogError()
	}

	s.m.Lock()
	defer s.m.Unlock()
	delete(s.sessions, key)

	return nil
}

type CacheSelector[TCacheStore, TData interface{}] func(*TCacheStore) *Store[TData] //nolint:revive

type OriginalFuncCaller[TData interface{}] func() (TData, error)

// CallWithSessionCache tries to find cache session based on a given context
// and get cached value from it by a given key, otherwise calls original function.
// If there is no session key in the context or no session in the SessionStore
// than it will always call the original function.
// If there is initialized session but no cached data by a given key
// it will call the original function and save its results.
// So the next call with the same key during the same session
// will not call the original function and take results from cache.
func CallWithSessionCache[TCacheStore, TData interface{}](
	ctx context.Context,
	s *SessionStore[TCacheStore],
	originalFn OriginalFuncCaller[TData],
	key Key,
	cacheSelector CacheSelector[TCacheStore, TData]) (TData, error) {
	session, ok := s.Session(ctx)

	if !ok {
		return originalFn()
	}

	cache := cacheSelector(session)

	v, ok := cache.Value(key)
	if ok {
		return v.Value()
	}

	r, err := originalFn()
	cache.SetValue(key, NewData(r, err))

	return r, err
}
