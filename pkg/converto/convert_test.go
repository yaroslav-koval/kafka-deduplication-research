package converto_test

import (
	"context"
	"kafka-polygon/pkg/cerror"
	"kafka-polygon/pkg/converto"
	"kafka-polygon/pkg/testutil"
	"math"
	"testing"
	"time"

	uuid "github.com/satori/go.uuid"
	"github.com/tj/assert"
)

var _cb = context.Background()

// SliceStringPointer
func TestSliceStringPointer(t *testing.T) {
	testSlice := []string{
		"test1", "test2",
	}
	actualSlice := converto.SliceStringPointer(testSlice)
	assert.IsType(t, &testSlice, actualSlice)
	assert.Equal(t, &testSlice, actualSlice)
}

// testing converting string
// StringPointer, StringPointerNotZero, StringValue
func TestString(t *testing.T) {
	inputVals := []string{
		"test string",
		"",
	}

	t.Run("check string pointer", func(t *testing.T) {
		for key, input := range inputVals {
			assert.IsType(t, &inputVals[key], converto.StringPointer(input))
			assert.Equal(t, &inputVals[key], converto.StringPointer(input))
		}
	})
	t.Run("check string pointer not zero", func(t *testing.T) {
		for key, input := range inputVals {
			assert.IsType(t, &inputVals[key], converto.StringPointerNotZero(input))
			if input == "" {
				assert.Nil(t, converto.StringPointerNotZero(input))
			} else {
				assert.Equal(t, &inputVals[key], converto.StringPointerNotZero(input))
			}
		}
	})
	t.Run("check string value", func(t *testing.T) {
		for key, input := range inputVals {
			assert.IsType(t, input, converto.StringValue(&inputVals[key]))
			assert.Equal(t, input, converto.StringValue(&inputVals[key]))
		}
	})
}

// BoolPointer
func TestBoolPointer(t *testing.T) {
	inputT, inputF := true, false
	assert.IsType(t, &inputT, converto.BoolPointer(inputT))
	assert.Equal(t, &inputT, converto.BoolPointer(inputT))
	assert.IsType(t, &inputF, converto.BoolPointer(inputF))
	assert.Equal(t, &inputF, converto.BoolPointer(inputF))
}

// BoolValue
func TestBoolValue(t *testing.T) {
	inputT, inputF := true, false
	assert.True(t, converto.BoolValue(&inputT))
	assert.False(t, converto.BoolValue(&inputF))
}

// BoolString
func TestBoolString(t *testing.T) {
	assert.Equal(t, "true", converto.BoolString(true))
}

// BoolValueString
func TestBoolValueString(t *testing.T) {
	inputT := true
	assert.Equal(t, "true", converto.BoolValueString(&inputT))
}

// IntPointer
func TestIntPointer(t *testing.T) {
	inputT := 10
	assert.IsType(t, &inputT, converto.IntPointer(inputT))
	assert.Equal(t, &inputT, converto.IntPointer(inputT))
}

// IntValue
func TestIntValue(t *testing.T) {
	inputT := 10
	assert.IsType(t, inputT, converto.IntValue(&inputT))
	assert.Equal(t, inputT, converto.IntValue(&inputT))
	assert.Equal(t, 0, converto.IntValue(nil))
}

// Int32Pointer
func TestInt32Pointer(t *testing.T) {
	inputT := int32(10)
	assert.IsType(t, &inputT, converto.Int32Pointer(inputT))
	assert.Equal(t, &inputT, converto.Int32Pointer(inputT))
}

// Int32Value
func TestInt32Value(t *testing.T) {
	inputT := int32(10)
	assert.IsType(t, inputT, converto.Int32Value(&inputT))
	assert.Equal(t, inputT, converto.Int32Value(&inputT))
	assert.Equal(t, int32(0), converto.Int32Value(nil))
}

// Int64Pointer
func TestInt64Pointer(t *testing.T) {
	inputT := int64(10)
	assert.IsType(t, &inputT, converto.Int64Pointer(inputT))
	assert.Equal(t, &inputT, converto.Int64Pointer(inputT))
}

// Int64Value
func TestInt64Value(t *testing.T) {
	inputT := int64(10)
	assert.IsType(t, inputT, converto.Int64Value(&inputT))
	assert.Equal(t, inputT, converto.Int64Value(&inputT))
	assert.Equal(t, int64(0), converto.Int64Value(nil))
}

// Uint16Pointer
func TestUint16Pointer(t *testing.T) {
	inputT := uint16(10)
	assert.IsType(t, &inputT, converto.Uint16Pointer(inputT))
	assert.Equal(t, &inputT, converto.Uint16Pointer(inputT))
}

// Uint16Value
func TestUint16Value(t *testing.T) {
	inputT := uint16(10)
	assert.IsType(t, inputT, converto.Uint16Value(&inputT))
	assert.Equal(t, inputT, converto.Uint16Value(&inputT))
	assert.Equal(t, uint16(0), converto.Uint16Value(nil))
}

// UintPointer
func TestUintPointer(t *testing.T) {
	inputT := uint(10)
	assert.IsType(t, &inputT, converto.UintPointer(inputT))
	assert.Equal(t, &inputT, converto.UintPointer(inputT))
}

// UintValue
func TestUintValue(t *testing.T) {
	inputT := uint(10)
	assert.IsType(t, inputT, converto.UintValue(&inputT))
	assert.Equal(t, inputT, converto.UintValue(&inputT))
	assert.Equal(t, uint(0), converto.UintValue(nil))
}

// Uint32Pointer
func TestUint32Pointer(t *testing.T) {
	inputT := uint32(10)
	assert.IsType(t, &inputT, converto.Uint32Pointer(inputT))
	assert.Equal(t, &inputT, converto.Uint32Pointer(inputT))
}

// Uint32Value
func TestUint32Value(t *testing.T) {
	inputT := uint32(10)
	assert.IsType(t, inputT, converto.Uint32Value(&inputT))
	assert.Equal(t, inputT, converto.Uint32Value(&inputT))
	assert.Equal(t, uint32(0), converto.Uint32Value(nil))
}

// Uint64Pointer
func TestUint64Pointer(t *testing.T) {
	inputT := uint64(10)
	assert.IsType(t, &inputT, converto.Uint64Pointer(inputT))
	assert.Equal(t, &inputT, converto.Uint64Pointer(inputT))
}

// Uint64Value
func TestUint64Value(t *testing.T) {
	inputT := uint64(10)
	assert.IsType(t, inputT, converto.Uint64Value(&inputT))
	assert.Equal(t, inputT, converto.Uint64Value(&inputT))
	assert.Equal(t, uint64(0), converto.Uint64Value(nil))
}

// Float32Pointer
func TestFloat32Pointer(t *testing.T) {
	inputT := float32(10)
	assert.IsType(t, &inputT, converto.Float32Pointer(inputT))
	assert.Equal(t, &inputT, converto.Float32Pointer(inputT))
}

// Float32Value
func TestFloat32Value(t *testing.T) {
	inputT := float32(10)
	assert.IsType(t, inputT, converto.Float32Value(&inputT))
	assert.Equal(t, inputT, converto.Float32Value(&inputT))
	assert.Equal(t, float32(0), converto.Float32Value(nil))
}

// Float64Pointer
func TestFloat64Pointer(t *testing.T) {
	inputT := float64(10)
	assert.IsType(t, &inputT, converto.Float64Pointer(inputT))
	assert.Equal(t, &inputT, converto.Float64Pointer(inputT))
}

// Float64Value
func TestFloat64Value(t *testing.T) {
	inputT := float64(10)
	assert.IsType(t, inputT, converto.Float64Value(&inputT))
	assert.Equal(t, inputT, converto.Float64Value(&inputT))
	assert.Equal(t, float64(0), converto.Float64Value(nil))
}

// TimePointer
func TestTimePointer(t *testing.T) {
	inputTimeNow := time.Now()
	assert.IsType(t, &inputTimeNow, converto.TimePointer(inputTimeNow))
	assert.Equal(t, &inputTimeNow, converto.TimePointer(inputTimeNow))
}

// TimeValue
func TestTimeValue(t *testing.T) {
	inputTimeNow := time.Now()
	assert.IsType(t, inputTimeNow, converto.TimeValue(&inputTimeNow))
	assert.Equal(t, inputTimeNow, converto.TimeValue(&inputTimeNow))
	assert.Equal(t, time.Time{}, converto.TimeValue(nil))
}

// NullUUIDFromStringPointer
func TestNullUUIDFromStringPointer(t *testing.T) {
	inputUUID := uuid.NewV4()
	uuidStr := inputUUID.String()
	expNULLUUID := uuid.NullUUID{UUID: inputUUID, Valid: true}

	assert.IsType(t, uuid.NullUUID{}, converto.NullUUIDFromStringPointer(&uuidStr))
	assert.Equal(t, uuid.NullUUID{}, converto.NullUUIDFromStringPointer(nil))
	assert.Equal(t, expNULLUUID, converto.NullUUIDFromStringPointer(&uuidStr))
}

// UUIDFromStringPointerOrNil
func TestUUIDFromStringPointerOrNil(t *testing.T) {
	inputUUID := uuid.NewV4()
	uuidStr := inputUUID.String()
	assert.IsType(t, uuid.UUID{}, converto.UUIDFromStringPointerOrNil(&uuidStr))
	assert.Equal(t, uuid.UUID{}, converto.UUIDFromStringPointerOrNil(nil))
	assert.Equal(t, inputUUID, converto.UUIDFromStringPointerOrNil(&uuidStr))
}

// PhoneNumber
func TestPhoneNumber(t *testing.T) {
	withPlusPhoneNumber := "+380123456789"
	withOutPlusPhoneNumber := "380123456789"

	assert.Equal(t, withPlusPhoneNumber, converto.PhoneNumber(withPlusPhoneNumber))
	assert.Equal(t, withPlusPhoneNumber, converto.PhoneNumber(withOutPlusPhoneNumber))
}

// MapToStruct
func TestMapToStruct(t *testing.T) {
	t.Parallel()

	err := converto.MapToStruct(_cb, map[string]interface{}{"1": math.Inf(1)}, nil)
	testutil.AssertCError(t, "failed marshal: json: unsupported value: +Inf", cerror.KindInternal.String(), err)

	err = converto.MapToStruct(_cb, map[string]interface{}{"1": 123}, nil)
	testutil.AssertCError(t, "failed unmarshal <nil>: json: Unmarshal(nil)", cerror.KindInternal.String(), err)

	src := map[string]interface{}{"key": "value"}
	dst := struct {
		Key string `json:"key"`
	}{}
	err = converto.MapToStruct(_cb, src, &dst)
	assert.NoError(t, err)
	assert.EqualValues(t, "value", dst.Key)
}

// StructToMap
func TestStructToMap(t *testing.T) {
	t.Parallel()

	err := converto.StructToMap(_cb, math.Inf(1), nil)
	testutil.AssertCError(t, "failed to marshal float64: json: unsupported value: +Inf", cerror.KindInternal.String(), err)

	src := struct {
		Key string `json:"key"`
	}{Key: "value"}
	dst := map[string]interface{}{}

	err = converto.StructToMap(_cb, src, dst)
	assert.NoError(t, err)
	assert.EqualValues(t, "value", dst["key"])
}
