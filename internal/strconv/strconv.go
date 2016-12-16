package strconv

import (
	"math"
	"strconv"
	"strings"
	"unicode/utf8"
)

var (
	ErrSyntax = strconv.ErrSyntax
	ErrRange  = strconv.ErrRange
)

func Atoi(s string) (i int, err error) {
	i64, err := ParseInt(s)
	return int(i64), err
}

func Itoa(i int) string {
	return FormatInt(int64(i), 10)
}

func AppendInt(dst []byte, i int64, base int) []byte {
	return strconv.AppendInt(dst, i, base)
}

func FormatInt(i int64, base int) string {
	return strconv.FormatInt(i, base)
}

func FormatUint(u uint64, base int) string {
	return strconv.FormatUint(u, base)
}

func FormatFloat(f float64, fmt byte, prec, bitSize int) string {
	s := strconv.FormatFloat(f, fmt, prec, bitSize)

	switch s {
	case "NaN":
		return "nan"
	case "-Inf":
		return "-inf"
	case "+Inf":
		return "inf"
	}

	return s
}

func ParseUint(s string) (uint64, error) {
	if len(s) == 0 {
		return 0, ErrSyntax
	}

	if s[0] == '0' && len(s) != 1 && (s[1] == 'x' || s[1] == 'X') {
		u, err := strconv.ParseUint(s[2:], 16, 64)
		return u, unwrap(err)
	}

	i, err := strconv.ParseUint(s, 10, 64)
	if err != nil {
		f, err := strconv.ParseFloat(s, 64)
		return uint64(f), unwrap(err)
	}

	return i, nil
}

func ParseInt(s string) (int64, error) {
	if len(s) == 0 {
		return 0, ErrSyntax
	}

	t := s

	var neg bool

	switch s[0] {
	case '-':
		neg = true
		t = t[1:]
	case '+':
		t = t[1:]
	}

	var i int64
	var err error

	if len(t) > 1 && t[0] == '0' && (t[1] == 'x' || t[1] == 'X') {
		var u uint64
		u, err = strconv.ParseUint(t[2:], 16, 64)
		i = int64(u)
		if neg {
			i = -i
		}
	} else {
		i, err = strconv.ParseInt(s, 10, 64)
	}

	return i, unwrap(err)
}

func ParseFloat(s string) (float64, error) {
	if len(s) == 0 {
		return 0, ErrSyntax
	}

	t := s

	var neg bool

	switch s[0] {
	case '-':
		neg = true
		t = t[1:]
	case '+':
		t = t[1:]
	}

	var f float64
	var err error

	if len(t) > 1 && t[0] == '0' && (t[1] == 'x' || t[1] == 'X') {
		f, err = parseHexFloat(t[2:])
		if neg {
			f = math.Copysign(f, -1)
		}
	} else {
		if len(t) > 0 && !(('0' <= t[0] && t[0] <= '9') || t[0] == '.') { // drop special cases. e.g "inf", "nan", ...
			err = ErrSyntax
		} else {
			f, err = strconv.ParseFloat(s, 64)
		}
	}

	return f, unwrap(err)
}

func parseHexFloat(s string) (float64, error) {
	if len(s) == 0 {
		return 0, ErrSyntax
	}

	var integer string
	var fraction string
	var exponent string

	if j := strings.IndexRune(s, '.'); j != -1 {
		integer = s[:j]
		s = s[j+1:]
		if k := strings.IndexAny(s, "pP"); k != -1 {
			fraction = s[:k]
			exponent = s[k+1:]
			if exponent == "" {
				return 0, ErrSyntax
			}
		} else {
			fraction = s
		}
	} else {
		if k := strings.IndexAny(s, "pP"); k != -1 {
			integer = s[:k]
			exponent = s[k+1:]
			if exponent == "" {
				return 0, ErrSyntax
			}
		} else {
			integer = s
		}
	}

	var f float64

	if integer != "" {
		i, err := strconv.ParseInt(integer, 16, 64)
		if err != nil {
			return 0, unwrap(err)
		}

		f = float64(i)
	}

	if fraction != "" {
		coef := 16.0

		var x int
		for _, r := range fraction {
			if r >= utf8.RuneSelf {
				return 0, ErrSyntax
			}
			x = digitVal(byte(r))
			if x == 16 {
				return 0, ErrSyntax
			}

			// do nothing
			if x == '0' {
				coef *= 16
				continue
			}

			f += float64(x) / coef

			coef *= 16
		}
	}

	if exponent != "" {
		e, err := strconv.ParseInt(exponent, 10, 64)
		if err != nil {
			return 0, unwrap(err)
		}

		f = f * math.Pow(2, float64(e))
	}

	return f, nil
}

func digitVal(c byte) int {
	switch {
	case uint(c)-'0' < 10:
		return int(c - '0')
	case uint(c)-'a' < 6:
		return int(c - 'a' + 10)
	case uint(c)-'A' < 6:
		return int(c - 'A' + 10)
	}

	return 16
}

func unwrap(err error) error {
	if err == nil {
		return nil
	}

	if nerr, ok := err.(*strconv.NumError); ok {
		return nerr.Err
	}

	return err
}
