package cache

import "sync"

// Data is a generic type to stored single unit of cached data.
// It is assumed that a method which returned value is the desired one for caching
// will always return two values - some data and error
// If you want to cache a pointer, use a pointer as a parameter (e.g. Data[*fhir.Task]).
type Data[T interface{}] struct {
	data T
	err  error
}

func NewData[T interface{}](data T, err error) *Data[T] {
	return &Data[T]{data: data, err: err}
}

func (s *Data[T]) Value() (T, error) {
	return s.data, s.err
}

// Store is an in-memory generic cache store.
// It stores data of a provided parameter type.
// For example Store[*fhir.Task] stores values of *Data[*fhir.Task] type.
type Store[T interface{}] struct {
	m    sync.Mutex
	data map[Key]*Data[T]
}

func NewStore[T interface{}]() *Store[T] {
	return &Store[T]{data: make(map[Key]*Data[T])}
}

func (s *Store[T]) Value(key Key) (*Data[T], bool) {
	s.m.Lock()
	defer s.m.Unlock()
	v, ok := s.data[key]

	return v, ok
}

func (s *Store[T]) SetValue(key Key, v *Data[T]) {
	s.m.Lock()
	defer s.m.Unlock()
	s.data[key] = v
}
