package reflect

import (
	"github.com/hirochachacha/plua/internal/strconv"

	"github.com/hirochachacha/plua/object"
)

func integerToString(i object.Integer) object.String {
	return object.String(strconv.FormatInt(int64(i), 10))
}

func numberToInteger(n object.Number) (object.Integer, bool) {
	ival := object.Integer(n)
	if n == object.Number(ival) {
		return ival, true
	}
	return ival, false
}

func numberToString(n object.Number) object.String {
	return object.String(strconv.FormatFloat(float64(n), 'f', 1, 64))
}

func numberToGoUint(n object.Number) (uint64, bool) {
	u := uint64(n)
	if n == object.Number(u) {
		return u, true
	}
	return u, false
}

func stringToInteger(s object.String) (object.Integer, bool) {
	i, err := strconv.ParseInt(string(s))
	if err != nil {
		return 0, false
	}
	return object.Integer(i), true
}

func stringToNumber(s object.String) (object.Number, bool) {
	f, err := strconv.ParseFloat(string(s))
	if err != nil {
		if err == strconv.ErrRange {
			return object.Number(f), true
		}
		return 0, false
	}
	return object.Number(f), true
}

func stringToGoUint(s object.String) (uint64, bool) {
	u, err := strconv.ParseUint(string(s))
	if err != nil {
		return 0, false
	}
	return u, true
}
