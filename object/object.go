package object

import (
	"math"
	"reflect"

	"github.com/hirochachacha/plua/internal/limits"
)

func Equal(x, y Value) bool {
	switch x := x.(type) {
	case GoFunction:
		if y, ok := y.(GoFunction); ok {
			return reflect.ValueOf(x).Pointer() == reflect.ValueOf(y).Pointer()
		}

		return false
	case Integer:
		if y, ok := y.(Number); ok {
			iy := Integer(y)
			return (x == iy && y == Number(iy))
		}
	case Number:
		if y, ok := y.(Integer); ok {
			ix := Integer(x)
			return (y == ix && x == Number(ix))
		}
	}

	return x == y
}

func Repr(val Value) string {
	if val == nil {
		return "nil"
	}

	return val.String()
}

func ToType(val Value) Type {
	if val == nil {
		return TNIL
	}

	return val.Type()
}

func ToInteger(val Value) (Integer, bool) {
	switch val := val.(type) {
	case Integer:
		return val, true
	case Number:
		return numberToInteger(val)
	case String:
		return stringToInteger(val)
	}

	return 0, false
}

func ToNumber(val Value) (Number, bool) {
	switch val := val.(type) {
	case Integer:
		return Number(val), true
	case Number:
		return val, true
	case String:
		return stringToNumber(val)
	}

	return 0, false
}

func ToString(val Value) (String, bool) {
	switch val := val.(type) {
	case String:
		return val, true
	case Integer:
		return integerToString(val), true
	case Number:
		return numberToString(val), true
	}

	return "", false
}

func ToBoolean(val Value) Boolean {
	switch val := val.(type) {
	case nil:
		return false
	case Boolean:
		return val
	}
	return true
}

func ToGoInt(val Value) (int, bool) {
	i, ok := ToGoInt64(val)

	if i > limits.MaxInt || i < limits.MinInt {
		return int(i), false
	}

	return int(i), ok
}

func ToGoInt64(val Value) (int64, bool) {
	i, ok := ToInteger(val)
	return int64(i), ok
}

func ToGoFloat64(val Value) (float64, bool) {
	f, ok := ToNumber(val)
	return float64(f), ok
}

func ToGoString(val Value) (string, bool) {
	s, ok := ToString(val)
	return string(s), ok
}

func ToGoBool(val Value) bool {
	return bool(ToBoolean(val))
}

func IsNaN(val Value) bool {
	if val, ok := val.(Number); ok {
		return math.IsNaN(float64(val))
	}
	return false
}
