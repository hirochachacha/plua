package strconv

import (
	"io"
	"math"
	"strconv"
	"strings"
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
	return strconv.FormatFloat(f, fmt, prec, bitSize)
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

	if s[0] == '0' && len(s) != 1 && (s[1] == 'x' || s[1] == 'X') {
		i, err := strconv.ParseInt(s[2:], 16, 64)
		return i, unwrap(err)
	}

	i, err := strconv.ParseInt(s, 10, 64)
	if err != nil {
		return i, unwrap(err)
	}

	return i, nil
}

func ParseFloat(s string) (float64, error) {
	if len(s) == 0 {
		return 0, ErrSyntax
	}

	if s[0] == '0' && len(s) != 1 && (s[1] == 'x' || s[1] == 'X') {
		return parseHexFloat(s[2:])
	}

	f, err := strconv.ParseFloat(s, 64)
	if err != nil {
		return f, unwrap(err)
	}
	return f, nil
}

func ScanUint(sc io.ByteScanner) (uint64, error) {
	s := newScanner(sc)

	s.next()

	u64, err := s.scanUint()
	if err != nil {
		return 0, err
	}

	err = s.sc.UnreadByte()
	if err != nil {
		return 0, err
	}

	return u64, err
}

func ScanInt(sc io.ByteScanner) (int64, error) {
	s := newScanner(sc)

	s.next()

	i64, err := s.scanInt()
	if err != nil {
		return 0, err
	}

	err = s.sc.UnreadByte()
	if err != nil {
		return 0, err
	}

	return i64, err
}

func ScanFloat(sc io.ByteScanner) (float64, error) {
	s := newScanner(sc)

	s.next()

	f64, err := s.scanFloat()
	if err != nil {
		return 0, err
	}

	err = s.sc.UnreadByte()
	if err != nil {
		return 0, err
	}

	return f64, err
}

func parseHexFloat(s string) (float64, error) {
	if len(s) == 0 {
		return 0, ErrSyntax
	}

	j := strings.IndexRune(s, '.')
	if j != -1 {
		var i int64
		var err error

		if j != 0 {
			i, err = strconv.ParseInt(s[:j], 16, 64)
			if err != nil {
				return 0, unwrap(err)
			}
		}

		f := float64(i)

		coef := 16.0

		var x int
		for k, r := range s[j+1:] {
			x = digitVal(r)
			if x == 16 {
				if r == 'p' || r == 'P' {
					e, err := strconv.ParseInt(s[j+k+2:], 10, 64)
					if err != nil {
						return 0, unwrap(err)
					}
					return f * math.Pow(2, float64(e)), nil
				}
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
		return f, nil
	}

	k := strings.IndexAny(s, "pP")
	if k != -1 {
		i, err := strconv.ParseInt(s[:k], 16, 64)
		if err != nil {
			return 0, unwrap(err)
		}

		e, err := strconv.ParseInt(s[k+1:], 10, 64)
		if err != nil {
			return 0, unwrap(err)
		}

		return float64(i) * math.Pow(2, float64(e)), nil
	}

	i, err := strconv.ParseInt(s, 16, 64)
	if err != nil {
		return 0, unwrap(err)
	}

	return float64(i), nil
}

func digitVal(r rune) int {
	switch {
	case uint(r)-'0' < 10:
		return int(r - '0')
	case uint(r)-'a' < 6:
		return int(r - 'a' + 10)
	case uint(r)-'A' < 6:
		return int(r - 'A' + 10)
	}

	return 16 // larger than any legal digit val
}

func unwrap(err error) error {
	if err == nil {
		return nil
	}

	return err.(*strconv.NumError).Err
}
