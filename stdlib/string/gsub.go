package string

import (
	"bytes"
	"errors"
	"fmt"
	"strings"

	"github.com/hirochachacha/plua/internal/arith"
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
		ret, count, e := gsubTable(th, s, pat, repl, n)
		if e != nil {
			return nil, object.NewRuntimeError(e.Error())
		}

		return []object.Value{object.String(ret), object.Integer(count)}, nil
	default:
		if repl, ok := object.ToGoString(repl); ok {
			ret, count, e := gsubStr(th, s, pat, repl, n)
			if e != nil {
				return nil, object.NewRuntimeError(e.Error())
			}

			return []object.Value{object.String(ret), object.Integer(count)}, nil
		}

		return nil, ap.TypeError(2, "string/function/table")
	}
}

func gsubFunc(th object.Thread, input, pat string, fn object.Value, n int) (string, int, error) {
	return pattern.ReplaceAllFunc(input, pat, func(caps []pattern.Capture) (string, error) {
		loc := caps[0]

		if len(caps) > 1 {
			caps = caps[1:]
		}
		rargs := make([]object.Value, len(caps))
		for i, cap := range caps {
			rargs[i] = cap.Value(input)
		}

		rets, rerr := th.Call(fn, nil, rargs...)
		if rerr != nil {
			return "", rerr
		}

		repl := loc.Value(input)

		if len(rets) > 0 {
			if val := rets[0]; val != nil && val != object.False {
				var ok bool
				repl, ok = object.ToString(val)
				if !ok {
					return "", fmt.Errorf("invalid replacement value (a %s)", object.ToType(val))
				}
			}
		}

		return repl.String(), nil
	}, n)
}

func gsubTable(th object.Thread, input, pat string, t object.Table, n int) (string, int, error) {
	return pattern.ReplaceAllFunc(input, pat, func(caps []pattern.Capture) (string, error) {
		loc := caps[0]

		repl := loc.Value(input)

		key := repl
		if len(caps) > 1 {
			key = caps[1].Value(input)
		}

		val, err := arith.CallGettable(th, t, key)
		if err != nil {
			return "", err
		}

		if val != nil && val != object.False {
			var ok bool
			repl, ok = object.ToString(val)
			if !ok {
				return "", fmt.Errorf("invalid replacement value (a %s)", object.ToType(val))
			}
		}

		return repl.String(), nil
	}, n)
}

func gsubStr(th object.Thread, input, pat, repl string, n int) (string, int, error) {
	parts, e := gsubParseRepl(repl)
	if e != nil {
		return "", -1, e
	}

	var buf bytes.Buffer

	return pattern.ReplaceAllFunc(input, pat, func(caps []pattern.Capture) (string, error) {
		loc := caps[0]

		if len(caps) == 1 {
			caps = append(caps, loc)
		}

		buf.Reset()

		for _, part := range parts {
			if part[0] == '%' {
				if part[1] == '%' {
					buf.WriteByte('%')
				} else {
					j := int(part[1] - '0')
					if j >= len(caps) {
						return "", fmt.Errorf("invalid capture index %%%d", j)
					}

					cap := caps[j]

					buf.WriteString(cap.Value(input).String())
				}
			} else {
				buf.WriteString(part)
			}
		}

		return buf.String(), nil
	}, n)
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
		if !('0' <= d && d <= '9') && d != '%' {
			return nil, errors.New("invalid use of '%' in replacement string")
		}

		parts = append(parts, repl[i:i+2])

		repl = repl[i+2:]
	}

	return parts, nil
}
