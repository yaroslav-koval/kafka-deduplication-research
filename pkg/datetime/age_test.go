package datetime_test

import (
	"kafka-polygon/pkg/datetime"
	"testing"
	"time"

	"github.com/tj/assert"
)

func TestAgeFrom(t *testing.T) {
	for _, a := range []struct {
		birthDate time.Time
		from      time.Time
		expected  int
	}{
		{time.Date(2000, 3, 14, 0, 0, 0, 0, time.UTC), time.Date(2010, 3, 14, 0, 0, 0, 0, time.UTC), 10},
		{time.Date(2001, 3, 14, 0, 0, 0, 0, time.UTC), time.Date(2009, 3, 14, 0, 0, 0, 0, time.UTC), 8},
		{time.Date(2004, 6, 18, 0, 0, 0, 0, time.UTC), time.Date(2005, 5, 12, 0, 0, 0, 0, time.UTC), 0},
	} {
		assert.Equal(t, a.expected, datetime.AgeFrom(a.birthDate, a.from))
	}
}
