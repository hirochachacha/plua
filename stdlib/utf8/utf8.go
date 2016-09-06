package utf8

import (
	"bytes"
	"unicode/utf8"

	"github.com/hirochachacha/plua/object"
	"github.com/hirochachacha/plua/object/fnutil"
)

func Char(th object.Thread, args ...object.Value) ([]object.Value, *object.RuntimeError) {
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

func NextCode(th object.Thread, args ...object.Value) ([]object.Value, *object.RuntimeError) {
	ap := fnutil.NewArgParser(th, args)

	s, err := ap.ToGoString(0)
	if err != nil {
		return nil, err
	}

	off, err := ap.ToGoInt(1)
	if err != nil {
		return nil, err
	}

	r, i := utf8.DecodeRuneInString(s[off:])
	if r == utf8.RuneError {
		return nil, object.NewRuntimeError("invalid UTF-8 code")
	}

	return []object.Value{object.Integer(off + i), object.Integer(r)}, nil
}

func Codes(th object.Thread, args ...object.Value) ([]object.Value, *object.RuntimeError) {
	ap := fnutil.NewArgParser(th, args)

	s, err := ap.ToString(0)
	if err != nil {
		return nil, err
	}

	return []object.Value{object.GoFunction(NextCode), s, object.Integer(0)}, nil
}

func CodePoint(th object.Thread, args ...object.Value) ([]object.Value, *object.RuntimeError) {
	ap := fnutil.NewArgParser(th, args)

	s, err := ap.ToGoString(0)
	if err != nil {
		return nil, err
	}

	i, err := ap.OptGoInt(1, 1)
	if err != nil {
		return nil, err
	}

	j, err := ap.OptGoInt(2, i)
	if err != nil {
		return nil, err
	}

	if i == 0 {
		return nil, ap.ArgError(1, "out of range")
	} else if i < 0 {
		i = len(s) + 1 + i

		if i <= 0 {
			return nil, ap.ArgError(1, "out of range")
		}

		if i > len(s) {
			return nil, ap.ArgError(1, "out of range")
		}
	} else if i > len(s) {
		return nil, ap.ArgError(1, "out of range")
	}

	if j == 0 {
		return nil, ap.ArgError(2, "out of range")
	} else if j < 0 {
		j = len(s) + 1 + j

		if j <= 0 {
			return nil, ap.ArgError(2, "out of range")
		}

		if j > len(s) {
			return nil, ap.ArgError(2, "out of range")
		}
	} else if j > len(s) {
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

func Len(th object.Thread, args ...object.Value) ([]object.Value, *object.RuntimeError) {
	ap := fnutil.NewArgParser(th, args)

	s, err := ap.ToGoString(0)
	if err != nil {
		return nil, err
	}

	return []object.Value{object.Integer(utf8.RuneCountInString(s))}, nil
}

func Offset(th object.Thread, args ...object.Value) ([]object.Value, *object.RuntimeError) {
	ap := fnutil.NewArgParser(th, args)

	s, err := ap.ToGoString(0)
	if err != nil {
		return nil, err
	}

	n, err := ap.ToGoInt(1)
	if err != nil {
		return nil, err
	}

	i, err := ap.OptGoInt(2, 1)
	if err != nil {
		return nil, err
	}

	scount := utf8.RuneCountInString(s)

	if n == 0 {
		n = 1
	} else if n < 0 {
		n = scount + 1 + n

		if n <= 0 {
			return []object.Value{nil}, nil
		}

		if n > scount+1 {
			return []object.Value{nil}, nil
		}
	} else if n > scount+1 {
		return []object.Value{nil}, nil
	}

	if i == 0 {
		return nil, ap.ArgError(2, "position out of range")
	} else if i < 0 {
		i = len(s) + 1 + i

		if i <= 0 {
			return nil, ap.ArgError(2, "position out of range")
		}

		if i > len(s)+1 {
			return nil, ap.ArgError(2, "position out of range")
		}
	} else if i > len(s)+1 {
		return nil, ap.ArgError(2, "position out of range")
	}

	for j := 0; j < n-1; j++ {
		r, k := utf8.DecodeRuneInString(s[i-1:])
		if r == utf8.RuneError {
			return nil, object.NewRuntimeError("invalid UTF-8 code")
		}

		i += k
	}

	return []object.Value{object.Integer(i)}, nil
}

func Open(th object.Thread, args ...object.Value) ([]object.Value, *object.RuntimeError) {
	m := th.NewTableSize(0, 6)

	m.Set(object.String("charpatt"), object.String("."))

	m.Set(object.String("char"), object.GoFunction(Char))
	m.Set(object.String("codes"), object.GoFunction(Codes))
	m.Set(object.String("codepoint"), object.GoFunction(CodePoint))
	m.Set(object.String("len"), object.GoFunction(Len))
	m.Set(object.String("offset"), object.GoFunction(Offset))

	return []object.Value{m}, nil
}
