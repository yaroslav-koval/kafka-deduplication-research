package postgres_test

import (
	"kafka-polygon/pkg/db/postgres"
	"testing"

	"github.com/tj/assert"
)

func TestUseZeroInt(t *testing.T) {
	t.Parallel()

	var expInt, newExpInt int

	expInt = 10
	newExpInt = 20

	uzInt := postgres.NewUseZeroInt(expInt)

	t.Run("user zero to int", func(t *testing.T) {
		assert.IsType(t, expInt, uzInt.ToInt())
		assert.Equal(t, expInt, uzInt.ToInt())
	})

	t.Run("user zero scan", func(t *testing.T) {
		err := uzInt.Scan(newExpInt)
		assert.NoError(t, err)

		assert.IsType(t, newExpInt, uzInt.ToInt())
		assert.Equal(t, newExpInt, uzInt.ToInt())
	})

	t.Run("user zero scan error", func(t *testing.T) {
		err := uzInt.Scan(float64(30))
		assert.Error(t, err)
		assert.Equal(t, "unsupported data type: float64", err.Error())
	})

	t.Run("user zero sql driver.Value", func(t *testing.T) {
		sqlVal, err := uzInt.Value()
		assert.NoError(t, err)
		assert.IsType(t, newExpInt, sqlVal)
		assert.Equal(t, newExpInt, sqlVal.(int))
	})
}

func TestUseZeroBool(t *testing.T) {
	t.Parallel()

	uzBool := postgres.NewUseZeroBool(true)

	t.Run("user zero to bool", func(t *testing.T) {
		assert.IsType(t, true, uzBool.ToBool())
		assert.Equal(t, true, uzBool.ToBool())
	})

	t.Run("user zero scan", func(t *testing.T) {
		err := uzBool.Scan(false)
		assert.NoError(t, err)

		assert.IsType(t, false, uzBool.ToBool())
		assert.Equal(t, false, uzBool.ToBool())
	})

	t.Run("user zero scan error", func(t *testing.T) {
		err := uzBool.Scan(float64(30))
		assert.Error(t, err)
		assert.Equal(t, "unsupported data type: float64", err.Error())
	})

	t.Run("user zero sql driver.Value", func(t *testing.T) {
		sqlVal, err := uzBool.Value()
		assert.NoError(t, err)
		assert.IsType(t, false, sqlVal)
		assert.Equal(t, false, sqlVal.(bool))
	})
}
