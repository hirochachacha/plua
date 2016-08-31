package object

import (
	"github.com/hirochachacha/plua/internal/strconv"
)

func integerToString(i Integer) String {
	return String(strconv.FormatInt(int64(i), 10))
}

func numberToInteger(n Number) (Integer, bool) {
	ival := Integer(n)
	if n == Number(ival) {
		return ival, true
	}
	return ival, false
}

func numberToString(n Number) String {
	return String(strconv.FormatFloat(float64(n), 'f', 1, 64))
}

func numberToGoUint(n Number) (uint64, bool) {
	u := uint64(n)
	if n == Number(u) {
		return u, true
	}
	return u, false
}

func stringToInteger(s String) (Integer, bool) {
	i, err := strconv.ParseInt(string(s))
	if err != nil {
		return 0, false
	}
	return Integer(i), true
}

func stringToNumber(s String) (Number, bool) {
	f, err := strconv.ParseFloat(string(s))
	if err != nil {
		if err == strconv.ErrRange {
			return Number(f), true
		}
		return 0, false
	}
	return Number(f), true
}

func stringToGoUint(s String) (uint64, bool) {
	u, err := strconv.ParseUint(string(s))
	if err != nil {
		return 0, false
	}
	return u, true
}
