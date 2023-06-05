package handlebars

import (
	"context"
	"kafka-polygon/pkg/cerror"
	"reflect"
	"strconv"

	"github.com/aymerick/raymond"
)

// CustomIfHelper is a custom helper for `if` statements in handlebars templates with more operators.
//
//	It supports the following operators:
//	   "eq"  -  compare two values for equality
//	   "noteq"  -  compare two values for inequality
//	   "lt"  -  compare if value1 is less than value2
//	   "lteq"  -  compare if value1 is less than or equal to value2
//	   "gt"  -  compare if value1 is greater than value2
//	   "gteq"  -  compare if value1 is greater than or equal to value2
//	   "and"  -  compare if value1 and value2 are true (has values, not empty)
//	   "or"  -  compare if value1 or value2 are true (has values, not empty)
//
// ------------------------------------------------------------
//
// Usage:
//
//	{{#when value1 "operator" value2}}
//		<p>Do something</p>
//	{{else}}
//		<p>Do something else</p>
//	{{/when}}
//
// Examples:
//
// ----
//
//	{{#when user_name "noteq" ""}}
//		<p>Hello, {{user_name}}</p>
//	{{else}}
//		<p>Hello,</p>
//	{{/when}}
//
// ----
//
//	{{#when count1 "gt" count2}}
//		<p>You have got, {{count1}} euro</p>
//	{{else}}
//		<p>You have got, {{count2}} euro</p>
//	{{/when}}
//
//nolint:gocyclo
func CustomIfHelper(value1, operator, value2 interface{}, options *raymond.Options) string {
	switch operator {
	case "eq":
		ok, err := assertValuesEqual(value1, value2, operator)
		if err != nil {
			return ""
		} else if ok {
			return options.Fn()
		}
	case "noteq":
		ok, err := assertValuesEqual(value1, value2, operator)
		if err != nil {
			return ""
		} else if !ok {
			return options.Fn()
		}
	case "lt":
		if Float(value1) < Float(value2) {
			return options.Fn()
		}
	case "lteq":
		if Float(value1) <= Float(value2) {
			return options.Fn()
		}
	case "gt":
		if Float(value1) > Float(value2) {
			return options.Fn()
		}
	case "gteq":
		if Float(value1) >= Float(value2) {
			return options.Fn()
		}
	case "and":
		if raymond.IsTrue(value1) && raymond.IsTrue(value2) {
			return options.Fn()
		}
	case "or":
		if raymond.IsTrue(value1) || raymond.IsTrue(value2) {
			return options.Fn()
		}
	}

	return options.Inverse()
}

// Float returns a float representation of the provided type value
// Note that is only able to parse strings or convert actual numbers
// It defaults to 0.0 if the value cannot be properly represented as a float
func Float(value interface{}) float64 {
	return floatValue(reflect.ValueOf(value))
}

// floatValue returns the float64 representation of a reflect.Value
func floatValue(value reflect.Value) float64 {
	result := float64(0.0)

	//nolint:exhaustive,lll
	switch value.Kind() {
	case reflect.String:
		result, _ = strconv.ParseFloat(value.String(), 64)
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64, reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64, reflect.Uintptr:
		result = float64(value.Int())
	case reflect.Float32, reflect.Float64:
		result = value.Float()
	}

	return result
}

func assertValuesEqual(value1, value2, operand interface{}) (bool, error) {
	if value1 != nil && value2 == nil || value1 == nil && value2 != nil {
		return false, nil
	} else if value1 == nil && value2 == nil {
		return true, nil
	}

	v1 := reflect.ValueOf(value1)
	v2 := reflect.ValueOf(value2)

	if v1.Type() != v2.Type() {
		return false, cerror.NewF(context.Background(), cerror.KindInternal,
			"values are not of the same type. v1: [value=%+v, type=%T], v2: [value=%+v, type=%T], op: %+v",
			value1, value1, value2, value2, operand).LogError()
	}

	return raymond.Str(value1) == raymond.Str(value2), nil
}
