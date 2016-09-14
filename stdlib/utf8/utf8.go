package utf8

import (
	"bytes"
	"unicode/utf8"

	"github.com/hirochachacha/plua/object"
	"github.com/hirochachacha/plua/object/fnutil"
)

func char(th object.Thread, args ...object.Value) ([]object.Value, *object.RuntimeError) {
	if len(args) == 0 {
		return []object.Value{object.String("")}, nil
	}

	ap := fnutil.NewArgParser(th, args)

	var buf bytes.Buffer

	for i := range args {
		i64, err := ap.ToGoInt64(i)
		if err != nil {
			return nil, err
		}

		if i64 > utf8.MaxRune {
			return nil, ap.ArgError(i, "value out of range")
		}

		buf.WriteRune(rune(i64))
	}

	return []object.Value{object.String(buf.String())}, nil
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
	if off < 0 {
		off = len(s) + 1 + off
	}

	if off <= 0 {
		r, _ := utf8.DecodeRuneInString(s)
		if r == utf8.RuneError {
			return nil, object.NewRuntimeError("invalid UTF-8 code")
		}

		return []object.Value{object.Integer(1), object.Integer(r)}, nil
	}

	if off >= len(s) {
		return nil, nil
	}

	r, i := utf8.DecodeRuneInString(s[off-1:])
	if r == utf8.RuneError {
		return nil, object.NewRuntimeError("invalid UTF-8 code")
	}

	if off+i >= len(s) {
		return nil, nil
	}

	r, _ = utf8.DecodeRuneInString(s[off-1+i:])
	if r == utf8.RuneError {
		return nil, object.NewRuntimeError("invalid UTF-8 code")
	}

	return []object.Value{object.Integer(off + i), object.Integer(r)}, nil
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

	scount := utf8.RuneCountInString(s)

	n, err := ap.ToGoInt(1)
	if err != nil {
		return nil, err
	}
	if n < 0 {
		n = scount + 1 + n
	}
	if n < 0 || n > scount {
		return []object.Value{nil}, nil
	}

	i, err := ap.OptGoInt(2, 1)
	if err != nil {
		return nil, err
	}
	if i < 0 {
		i = len(s) + 1 + i
	}
	if i <= 0 || i > len(s) {
		return nil, ap.ArgError(2, "position out of range")
	}

	if n == 0 {
		j := 1
		for {
			r, rsize := utf8.DecodeRuneInString(s[j-1:])
			if r == utf8.RuneError {
				return nil, object.NewRuntimeError("invalid UTF-8 code")
			}

			if j <= i && i < j+rsize {
				return []object.Value{object.Integer(j)}, nil
			}

			j += rsize
		}
	}

	for j := 0; j < n-1; j++ {
		r, rsize := utf8.DecodeRuneInString(s[i-1:])
		if r == utf8.RuneError {
			return nil, object.NewRuntimeError("invalid UTF-8 code")
		}

		i += rsize
	}

	return []object.Value{object.Integer(i)}, nil
}

func Open(th object.Thread, args ...object.Value) ([]object.Value, *object.RuntimeError) {
	m := th.NewTableSize(0, 6)

	m.Set(object.String("charpatt"), object.String("."))

	m.Set(object.String("char"), object.GoFunction(char))
	m.Set(object.String("codes"), object.GoFunction(codes))
	m.Set(object.String("codepoint"), object.GoFunction(codepoint))
	m.Set(object.String("len"), object.GoFunction(_len))
	m.Set(object.String("offset"), object.GoFunction(offset))

	return []object.Value{m}, nil
}
