package string

import (
	"bytes"
	"errors"
	"strings"

	"github.com/hirochachacha/plua/internal/pattern"
	"github.com/hirochachacha/plua/object"
	"github.com/hirochachacha/plua/object/fnutil"
)

// gsub(s, pattern, repl, [, n])
func gsub(th object.Thread, args ...object.Value) ([]object.Value, *object.RuntimeError) {
	ap := fnutil.NewArgParser(th, args)

	s, err := ap.ToGoString(0)
	if err != nil {
		return nil, err
	}

	pat, err := ap.ToGoString(1)
	if err != nil {
		return nil, err
	}

	repl, err := ap.ToValue(2)
	if err != nil {
		return nil, err
	}

	n, err := ap.OptGoInt(3, -1)
	if err != nil {
		return nil, err
	}

	switch repl := repl.(type) {
	case object.GoFunction, object.Closure:
		ret, count, e := gsubFunc(th, s, pat, repl, n)
		if e != nil {
			return nil, object.NewRuntimeError(e.Error())
		}

		return []object.Value{object.String(ret), object.Integer(count)}, nil
	case object.Table:
		ret, count, e := gsubTable(s, pat, repl, n)
		if e != nil {
			return nil, object.NewRuntimeError(e.Error())
		}

		return []object.Value{object.String(ret), object.Integer(count)}, nil
	default:
		if repl, ok := object.ToGoString(repl); ok {
			ret, count, e := gsubStr(s, pat, repl, n)
			if e != nil {
				return nil, object.NewRuntimeError(e.Error())
			}

			return []object.Value{object.String(ret), object.Integer(count)}, nil
		}

		return nil, ap.TypeError(2, "string/function/table")
	}
}

func gsubFunc(th object.Thread, input, pat string, fn object.Value, n int) (string, int, error) {
	p, e := pattern.Compile(pat)
	if e != nil {
		return "", -1, e
	}

	off := 0

	captures := p.FindIndex(input, off)
	if captures == nil {
		return input, 0, nil
	}

	var i int
	var buf bytes.Buffer

	for i = n; i != 0; i-- {
		loc := captures[0]

		buf.WriteString(input[off:loc.Begin])

		argIndex := captures
		if len(captures) > 1 {
			argIndex = argIndex[1:]
		}

		rargs := make([]object.Value, len(argIndex))
		for i, cap := range argIndex {
			rargs[i] = cap.Value(input)
		}

		rets, rerr := th.Call(fn, nil, rargs...)
		if rerr != nil {
			return "", -1, rerr
		}

		repl := loc.Value(input)

		if len(rets) > 0 {
			if s, ok := object.ToString(rets[0]); ok {
				repl = s
			}
		}

		buf.WriteString(repl.String())

		off = loc.End
		if off == len(input) {
			i--

			break
		}
		if off == loc.Begin {
			buf.WriteByte(input[off])

			off++
		}

		captures = p.FindIndex(input, off)
		if captures == nil {
			i--

			break
		}
	}

	buf.WriteString(input[off:])

	return buf.String(), n - i, nil
}

func gsubTable(input, pat string, t object.Table, n int) (string, int, error) {
	p, e := pattern.Compile(pat)
	if e != nil {
		return "", -1, e
	}

	off := 0

	captures := p.FindIndex(input, off)
	if captures == nil {
		return input, 0, nil
	}

	var i int
	var buf bytes.Buffer

	for i = n; i != 0; i-- {
		loc := captures[0]

		buf.WriteString(input[off:loc.Begin])

		repl := loc.Value(input)

		key := repl
		if len(captures) > 1 {
			key = captures[1].Value(input)
		}

		if s, ok := object.ToString(t.Get(key)); ok {
			repl = s
		}

		buf.WriteString(repl.String())

		off = loc.End
		if off == len(input) {
			i--

			break
		}
		if off == loc.Begin {
			buf.WriteByte(input[off])

			off++
		}

		captures = p.FindIndex(input, off)
		if captures == nil {
			i--

			break
		}
	}

	buf.WriteString(input[off:])

	return buf.String(), n - i, nil
}

func gsubStr(input, pat, repl string, n int) (string, int, error) {
	p, e := pattern.Compile(pat)
	if e != nil {
		return "", -1, e
	}

	off := 0

	captures := p.FindIndex(input, off)
	if captures == nil {
		return input, 0, nil
	}

	var i int
	var buf bytes.Buffer

	parts, e := gsubParseRepl(repl)
	if e != nil {
		return "", -1, e
	}

	for i = n; i != 0; i-- {
		loc := captures[0]

		buf.WriteString(input[off:loc.Begin])

		if len(captures) == 1 {
			captures = append(captures, loc)
		}

		for _, part := range parts {
			if part[0] == '%' {
				j := int(part[1] - '0')
				if j >= len(captures) {
					return "", -1, errors.New("invalid use of '%' in replacement string")
				}

				cap := captures[j]

				buf.WriteString(cap.Value(input).String())
			} else {
				buf.WriteString(part)
			}
		}

		off = loc.End
		if off == len(input) {
			i--

			break
		}
		if off == loc.Begin {
			buf.WriteByte(input[off])

			off++
		}

		captures = p.FindIndex(input, off)
		if captures == nil {
			i--

			break
		}
	}

	buf.WriteString(input[off:])

	return buf.String(), n - i, nil
}

func gsubParseRepl(repl string) (parts []string, err error) {
	for {
		i := strings.IndexByte(repl, '%')
		if i == -1 {
			if repl != "" {
				parts = append(parts, repl)
			}

			break
		}

		if i != 0 {
			parts = append(parts, repl[:i])
		}

		if i == len(repl)-1 {
			return nil, errors.New("invalid use of '%' in replacement string")
		}

		d := repl[i+1]
		if !('0' <= d && d <= '9') {
			return nil, errors.New("invalid use of '%' in replacement string")
		}

		parts = append(parts, repl[i:i+2])

		repl = repl[i+2:]
	}

	return parts, nil
}
