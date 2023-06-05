package test

import (
	"fmt"
	"kafka-polygon/pkg/fhir/cache"
	"sync"
	"testing"

	"github.com/tj/assert"
)

func TestDataValue(t *testing.T) {
	t.Parallel()

	v := "1"
	err := fmt.Errorf("err")
	d := cache.NewData(v, err)
	actV, actErr := d.Value()
	assert.Equal(t, v, actV)
	assert.Equal(t, err, actErr)
}

func TestStoreValue(t *testing.T) {
	t.Parallel()

	s := cache.NewStore[string]()
	key := "1"

	_, ok := s.Value(key)
	assert.False(t, ok)

	d := cache.NewData("2", nil)
	s.SetValue(key, d)

	actD, ok := s.Value(key)
	assert.True(t, ok)
	assert.Equal(t, d, actD)
}

func TestConcurrentValue(t *testing.T) {
	s := cache.NewStore[string]()

	wg := sync.WaitGroup{}
	groupCount := 10000

	wg.Add(groupCount)

	for i := 0; i < groupCount; i++ {
		go func(k string) {
			defer wg.Done()

			s.SetValue(k, cache.NewData(k, nil))
		}(fmt.Sprintf("%d", i))
	}

	wg.Add(groupCount)

	for i := 0; i < groupCount; i++ {
		go func(k string) {
			defer wg.Done()

			_, _ = s.Value(k)
		}(fmt.Sprintf("%d", i))
	}

	wg.Wait()
}
