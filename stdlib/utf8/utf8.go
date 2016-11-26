package utf8

import (
	"unicode/utf8"

	"github.com/hirochachacha/plua/object"
	"github.com/hirochachacha/plua/object/fnutil"
)

func char(th object.Thread, args ...object.Value) ([]object.Value, *object.RuntimeError) {
	if len(args) == 0 {
		return []object.Value{object.String("")}, nil
	}

	ap := fnutil.NewArgParser(th, args)

	rs := make([]rune, len(args))

	for i := range args {
		i64, err := ap.ToGoInt64(i)
		if err != nil {
			return nil, err
		}

		if i64 > utf8.MaxRune {
			return nil, ap.ArgError(i, "value out of range")
		}

		rs[i] = rune(i64)
	}

	return []object.Value{object.String(string(rs))}, nil
}

func nextcode(th object.Thread, args ...object.Value) ([]object.Value, *object.RuntimeError) {
	ap := fnutil.NewArgParser(th, args)

	s, err := ap.ToGoString(0)
	if err != nil {
		return nil, err
	}

	off, err := ap.ToGoInt(1)
	if err != nil {
		return nil, err
	}

	if off >= len(s) {
		return nil, nil
	}

	if off <= 0 {
		r, _ := utf8.DecodeRuneInString(s)
		if r == utf8.RuneError {
			return nil, object.NewRuntimeError("invalid UTF-8 code")
		}

		return []object.Value{object.Integer(1), object.Integer(r)}, nil
	}

	r, rsize := utf8.DecodeRuneInString(s[off-1:])
	if r == utf8.RuneError {
		return nil, object.NewRuntimeError("invalid UTF-8 code")
	}

	off += rsize

	if off > len(s) {
		return nil, nil
	}

	r, _ = utf8.DecodeRuneInString(s[off-1:])
	if r == utf8.RuneError {
		return nil, object.NewRuntimeError("invalid UTF-8 code")
	}

	return []object.Value{object.Integer(off), object.Integer(r)}, nil
}

func codes(th object.Thread, args ...object.Value) ([]object.Value, *object.RuntimeError) {
	ap := fnutil.NewArgParser(th, args)

	s, err := ap.ToString(0)
	if err != nil {
		return nil, err
	}

	return []object.Value{object.GoFunction(nextcode), s, object.Integer(0)}, nil
}

func codepoint(th object.Thread, args ...object.Value) ([]object.Value, *object.RuntimeError) {
	ap := fnutil.NewArgParser(th, args)

	s, err := ap.ToGoString(0)
	if err != nil {
		return nil, err
	}

	i, err := ap.OptGoInt(1, 1)
	if err != nil {
		return nil, err
	}
	if i < 0 {
		i = len(s) + 1 + i
	}
	if i <= 0 || i > len(s) {
		return nil, ap.ArgError(1, "out of range")
	}

	j, err := ap.OptGoInt(2, i)
	if err != nil {
		return nil, err
	}
	if j < 0 {
		j = len(s) + 1 + j
	}
	if j <= 0 || j > len(s) {
		return nil, ap.ArgError(2, "out of range")
	}

	if i > j {
		return nil, nil
	}

	var rets []object.Value

	for {
		r, k := utf8.DecodeRuneInString(s[i-1:])

		if r == utf8.RuneError {
			return nil, object.NewRuntimeError("invalid UTF-8 code")
		}

		rets = append(rets, object.Integer(r))

		i += k

		if i > j {
			break
		}
	}

	return rets, nil
}

func _len(th object.Thread, args ...object.Value) ([]object.Value, *object.RuntimeError) {
	ap := fnutil.NewArgParser(th, args)

	s, err := ap.ToGoString(0)
	if err != nil {
		return nil, err
	}

	i, err := ap.OptGoInt(1, 1)
	if err != nil {
		return nil, err
	}
	if i < 0 {
		i = len(s) + 1 + i
	}
	if i <= 0 || i > len(s) {
		return nil, ap.ArgError(1, "out of range")
	}

	j, err := ap.OptGoInt(2, -1)
	if err != nil {
		return nil, err
	}
	if j < 0 {
		j = len(s) + 1 + j
	}
	if j <= 0 || j > len(s) {
		return nil, ap.ArgError(2, "out of range")
	}

	if i > j {
		return nil, nil
	}

	var n int
	for i <= j {
		r, rsize := utf8.DecodeRuneInString(s[i-1:])
		if r == utf8.RuneError {
			return []object.Value{object.False, object.Integer(i)}, nil
		}
		i += rsize
		n++
	}

	return []object.Value{object.Integer(n)}, nil
}

func offset(th object.Thread, args ...object.Value) ([]object.Value, *object.RuntimeError) {
	ap := fnutil.NewArgParser(th, args)

	s, err := ap.ToGoString(0)
	if err != nil {
		return nil, err
	}

	n, err := ap.ToGoInt(1)
	if err != nil {
		return nil, err
	}
	i := 1
	if n < 0 {
		i = len(s) + 1
	}

	i, err = ap.OptGoInt(2, i)
	if err != nil {
		return nil, err
	}
	if i < 0 {
		i = len(s) + 1 + i
	}
	if i <= 0 || i > len(s)+1 {
		return nil, ap.ArgError(2, "position out of range")
	}

	var r rune
	var rsize int

	switch {
	case n < 0:
		if i != len(s)+1 && !utf8.RuneStart(s[i-1]) {
			return nil, object.NewRuntimeError("initial position is a continuation byte")
		}
		for n < 0 && i > 0 {
			r, rsize = utf8.DecodeLastRuneInString(s[:i-1])
			if r == utf8.RuneError {
				break
			}
			i -= rsize
			n++
		}
	case n > 0:
		if i != len(s)+1 && !utf8.RuneStart(s[i-1]) {
			return nil, object.NewRuntimeError("initial position is a continuation byte")
		}
		n--
		for n > 0 && i < len(s)+1 {
			r, rsize = utf8.DecodeRuneInString(s[i-1:])
			if r == utf8.RuneError {
				break
			}
			i += rsize
			n--
		}
	default:
		if len(s) == i-1 {
			return []object.Value{object.Integer(i)}, nil
		}
		for i > 0 && !utf8.RuneStart(s[i-1]) {
			i--
		}
	}

	if n == 0 {
		return []object.Value{object.Integer(i)}, nil
	}

	return nil, nil
}

func Open(th object.Thread, args ...object.Value) ([]object.Value, *object.RuntimeError) {
	m := th.NewTableSize(0, 6)

	m.Set(object.String("charpattern"), object.String("[\x00-\x7F\xC2-\xF4][\x80-\xBF]*"))

	m.Set(object.String("char"), object.GoFunction(char))
	m.Set(object.String("codes"), object.GoFunction(codes))
	m.Set(object.String("codepoint"), object.GoFunction(codepoint))
	m.Set(object.String("len"), object.GoFunction(_len))
	m.Set(object.String("offset"), object.GoFunction(offset))

	return []object.Value{m}, nil
}
