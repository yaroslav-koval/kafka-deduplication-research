package converto

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"kafka-polygon/pkg/cerror"
	"strconv"
	"strings"
	"time"

	uuid "github.com/satori/go.uuid"
)

func BytePointer(v []byte) *[]byte {
	return &v
}

func SliceStringPointer(v []string) *[]string {
	return &v
}

// StringPointer returns a pointer to the string value passed in.
func StringPointer(v string) *string {
	return &v
}

// StringPointerNotZero returns a pointer to the string value passed in
// or nil if string value is ""
func StringPointerNotZero(v string) *string {
	if v == "" {
		return nil
	}

	return &v
}

// StringValue returns the value of the string pointer passed in or
// "" if the pointer is nil.
func StringValue(v *string) string {
	if v != nil {
		return *v
	}

	return ""
}

// BoolPointer returns a pointer to of the bool value passed in.
func BoolPointer(v bool) *bool {
	return &v
}

// BoolValue returns the value of the bool pointer passed in or
// false if the pointer is nil.
func BoolValue(v *bool) bool {
	if v != nil {
		return *v
	}

	return false
}

// BoolString return the string value of the bool
func BoolString(v bool) string {
	return strconv.FormatBool(v)
}

// BoolValueString returns the value of the bool as a string pointer passed in or
// false if the pointer is nil.
func BoolValueString(v *bool) string {
	return strconv.FormatBool(BoolValue(v))
}

// IntPointer returns a pointer to of the int value passed in.
func IntPointer(v int) *int {
	return &v
}

// IntValue returns the value of the int pointer passed in or
// 0 if the pointer is nil.
func IntValue(v *int) int {
	if v != nil {
		return *v
	}

	return 0
}

// Int32Pointer returns a pointer to of the int32 value passed in.
func Int32Pointer(v int32) *int32 {
	return &v
}

// Int32Value returns the value of the int32 pointer passed in or
// 0 if the pointer is nil.
func Int32Value(v *int32) int32 {
	if v != nil {
		return *v
	}

	return 0
}

// Int64Pointer returns a pointer to of the int64 value passed in.
func Int64Pointer(v int64) *int64 {
	return &v
}

// Int64Value returns the value of the int64 pointer passed in or
// 0 if the pointer is nil.
func Int64Value(v *int64) int64 {
	if v != nil {
		return *v
	}

	return 0
}

// Uint16Pointer returns a pointer to of the uint16 value passed in.
func Uint16Pointer(v uint16) *uint16 {
	return &v
}

// Uint16Value returns the value of the uint16 pointer passed in or
// 0 if the pointer is nil.
func Uint16Value(v *uint16) uint16 {
	if v != nil {
		return *v
	}

	return 0
}

// UintPointer returns a pointer to of the uint value passed in.
func UintPointer(v uint) *uint {
	return &v
}

// UintValue returns the value of the uint pointer passed in or
// 0 if the pointer is nil.
func UintValue(v *uint) uint {
	if v != nil {
		return *v
	}

	return 0
}

// Uint32Pointer returns a pointer to of the uint32 value passed in.
func Uint32Pointer(v uint32) *uint32 {
	return &v
}

// Uint32Value returns the value of the uint32 pointer passed in or
// 0 if the pointer is nil.
func Uint32Value(v *uint32) uint32 {
	if v != nil {
		return *v
	}

	return 0
}

// Uint64Pointer returns a pointer to of the uint64 value passed in.
func Uint64Pointer(v uint64) *uint64 {
	return &v
}

// Uint64Value returns the value of the uint64 pointer passed in or
// 0 if the pointer is nil.
func Uint64Value(v *uint64) uint64 {
	if v != nil {
		return *v
	}

	return 0
}

// Float32Pointer returns a pointer to of the float32 value passed in.
func Float32Pointer(v float32) *float32 {
	return &v
}

// Float32Value returns the value of the float32 pointer passed in or
// 0 if the pointer is nil.
func Float32Value(v *float32) float32 {
	if v != nil {
		return *v
	}

	return 0
}

// Float64Pointer returns a pointer to of the float64 value passed in.
func Float64Pointer(v float64) *float64 {
	return &v
}

// Float64Value returns the value of the float64 pointer passed in or
// 0 if the pointer is nil.
func Float64Value(v *float64) float64 {
	if v != nil {
		return *v
	}

	return 0
}

// TimePointer returns a pointer to of the time.Time value passed in.
func TimePointer(v time.Time) *time.Time {
	return &v
}

// TimeValue returns the value of the time.Time pointer passed in or
// time.Time{} if the pointer is nil.
func TimeValue(v *time.Time) time.Time {
	if v != nil {
		return *v
	}

	return time.Time{}
}

func NullUUIDFromStringPointer(v *string) uuid.NullUUID {
	if v == nil {
		return uuid.NullUUID{Valid: false}
	}

	vUUID, err := uuid.FromString(StringValue(v))

	return uuid.NullUUID{UUID: vUUID, Valid: err == nil}
}

// UUIDFromStringPointerOrNil Convert *string to uuid.UUID
func UUIDFromStringPointerOrNil(value *string) uuid.UUID {
	result := uuid.Nil

	if value != nil {
		result = uuid.FromStringOrNil(*value)
	}

	return result
}

// UUIDFromUUIDPointer Convert *uuid.UUID to uuid.UUID
func UUIDFromUUIDPointer(value *uuid.UUID) uuid.UUID {
	result := uuid.Nil

	if value != nil {
		result = uuid.FromStringOrNil(value.String())
	}

	return result
}

// SQLNullFloat64 returns a null sql float64 based on float64 value passed in.
func SQLNullFloat64(v float64) sql.NullFloat64 {
	return sql.NullFloat64{Float64: v, Valid: true}
}

// SQLNullFloat64FromPointer returns a null sql float64 based on float64 pointer passed in.
func SQLNullFloat64FromPointer(v *float64) sql.NullFloat64 {
	if v == nil {
		return sql.NullFloat64{Valid: false}
	}

	return sql.NullFloat64{Float64: *v, Valid: true}
}

// SQLNullInt64FromPointer returns a null sql in64 based on int64 pointer passed in.
func SQLNullInt64FromPointer(v *int64) sql.NullInt64 {
	if v == nil {
		return sql.NullInt64{Valid: false}
	}

	return sql.NullInt64{Int64: *v, Valid: true}
}

// PhoneNumber convert phoneNumber, add or not add plus
func PhoneNumber(number string) string {
	number = strings.ReplaceAll(number, " ", "")

	plusSign := false
	if strings.HasPrefix(number, "+") {
		plusSign = true
	}

	if !plusSign {
		return fmt.Sprintf("+%s", number)
	}

	return number
}

// MapToStruct converts map[string]interface{} to struct of any type
func MapToStruct(ctx context.Context, m map[string]interface{}, dst interface{}) error {
	b, err := json.Marshal(m)
	if err != nil {
		return cerror.NewF(ctx, cerror.KindInternal, "failed marshal: %s", err.Error()).LogError()
	}

	err = json.Unmarshal(b, dst)
	if err != nil {
		return cerror.NewF(ctx, cerror.KindInternal, "failed unmarshal %T: %s", dst, err.Error()).LogError()
	}

	return nil
}

// StructToMap converts struct of any type to map[string]interface{}
func StructToMap(ctx context.Context, s interface{}, m map[string]interface{}) error {
	b, err := json.Marshal(s)
	if err != nil {
		return cerror.NewF(ctx, cerror.KindInternal, "failed to marshal %T: %s", s, err.Error()).LogError()
	}

	err = json.Unmarshal(b, &m)
	if err != nil {
		return cerror.NewF(ctx, cerror.KindInternal, "failed to unmarshal: %s", err.Error()).LogError()
	}

	return nil
}
