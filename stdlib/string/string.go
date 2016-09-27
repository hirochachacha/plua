package string

import (
	"strings"
	"unicode/utf8"

	"github.com/hirochachacha/plua/internal/compiler_pool"
	"github.com/hirochachacha/plua/internal/limits"
	"github.com/hirochachacha/plua/internal/pattern"
	"github.com/hirochachacha/plua/object"
	"github.com/hirochachacha/plua/object/fnutil"
)

func _byte(th object.Thread, args ...object.Value) ([]object.Value, *object.RuntimeError) {
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
	if i < 1 {
		i = 1
	}

	j, err := ap.OptGoInt(2, i)
	if err != nil {
		return nil, err
	}
	if j < 0 {
		j = len(s) + 1 + j
	}
	if j > len(s) {
		j = len(s)
	}

	if i > j {
		return nil, nil
	}

	rets := make([]object.Value, j-i+1)
	for k := range rets {
		rets[k] = object.Integer(s[k+i-1])
	}

	return rets, nil
}

func char(th object.Thread, args ...object.Value) ([]object.Value, *object.RuntimeError) {
	ap := fnutil.NewArgParser(th, args)

	bs := make([]byte, len(args))

	for i := range args {
		i64, err := ap.ToGoInt64(i)
		if err != nil {
			return nil, err
		}

		if i64 < 0 || i64 > 255 {
			return nil, ap.ArgError(i, "(value out of range)")
		}

		bs[i] = byte(i64)
	}

	return []object.Value{object.String(bs)}, nil
}

func dump(th object.Thread, args ...object.Value) ([]object.Value, *object.RuntimeError) {
	ap := fnutil.NewArgParser(th, args)

	cl, err := ap.ToClosure(0)
	if err != nil {
		return nil, err
	}

	strip := ap.OptGoBool(1, false)

	code, e := compiler_pool.DumpToString(cl.Prototype(), strip)
	if e != nil {
		return nil, object.NewRuntimeError(e.Error())
	}

	return []object.Value{object.String(code)}, nil
}

func find(th object.Thread, args ...object.Value) ([]object.Value, *object.RuntimeError) {
	ap := fnutil.NewArgParser(th, args)

	s, err := ap.ToGoString(0)
	if err != nil {
		return nil, err
	}

	pat, err := ap.ToGoString(1)
	if err != nil {
		return nil, err
	}

	init, err := ap.OptGoInt(2, 1)
	if err != nil {
		return nil, err
	}
	if init < 0 {
		init = len(s) + 1 + init
	}
	if init < 0 {
		init = 1
	}
	if init-1 > len(s) {
		return nil, nil
	}

	s = s[init-1:]

	isPlain := ap.OptGoBool(3, false)

	if isPlain {
		idx := strings.Index(s, pat)
		if idx == -1 {
			return nil, nil
		}

		return []object.Value{object.Integer(idx + init), object.Integer(idx + init + len(pat) - 1)}, nil
	}

	r, e := pattern.Find(pat, s)
	if e != nil {
		return nil, object.NewRuntimeError(e.Error())
	}

	if r == nil {
		return []object.Value{nil}, nil
	}

	rets := make([]object.Value, len(r.Captures)+2)

	rets[0] = object.Integer(r.Item.Begin + init)
	rets[1] = object.Integer(r.Item.End + init - 1)

	for i, cap := range r.Captures {
		rets[i+2] = object.String(s[cap.Begin:cap.End])
	}

	return rets, nil
}

// gmatch(s, patter)
func gmatch(th object.Thread, args ...object.Value) ([]object.Value, *object.RuntimeError) {
	ap := fnutil.NewArgParser(th, args)

	s, err := ap.ToGoString(0)
	if err != nil {
		return nil, err
	}

	pat, err := ap.ToGoString(1)
	if err != nil {
		return nil, err
	}

	ms, e := pattern.MatchAll(pat, s)
	if e != nil {
		return nil, object.NewRuntimeError(e.Error())
	}

	var i int

	fn := func(_ object.Thread, _ ...object.Value) ([]object.Value, *object.RuntimeError) {
		if len(ms) <= i {
			return nil, nil
		}

		m := ms[i]

		i++

		if len(m.Captures) == 0 {
			return []object.Value{object.String(m.Item)}, nil
		}

		rets := make([]object.Value, len(m.Captures))
		for i, cap := range m.Captures {
			rets[i] = object.String(cap)
		}

		return rets, nil
	}

	return []object.Value{object.GoFunction(fn)}, nil
}

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
		var rerr *object.RuntimeError

		replfn := func(ss ...string) string {
			rargs := make([]object.Value, len(ss))
			for i := range ss {
				rargs[i] = object.String(ss[i])
			}

			var rets []object.Value

			rets, rerr = th.Call(repl, nil, rargs...)
			if rerr != nil {
				return ""
			}

			if len(rets) > 0 {
				if s, ok := object.ToGoString(rets[0]); ok {
					return s
				}
			}
			return ""
		}

		ret, count, e := pattern.ReplaceFunc(pat, s, replfn, n)
		if e != nil {
			return nil, object.NewRuntimeError(e.Error())
		}

		if rerr != nil {
			return nil, rerr
		}

		return []object.Value{object.String(ret), object.Integer(count)}, nil
	case object.Table:
		replfn := func(ss ...string) string {
			if len(ss) > 0 {
				v := repl.Get(object.String(ss[0]))
				if r, ok := object.ToGoString(v); ok {
					return r
				}
				return ss[0]
			}
			return ""
		}

		ret, count, e := pattern.ReplaceFunc(pat, s, replfn, n)
		if e != nil {
			return nil, object.NewRuntimeError(e.Error())
		}

		return []object.Value{object.String(ret), object.Integer(count)}, nil
	default:
		if repl, ok := object.ToGoString(repl); ok {
			ret, count, e := pattern.Replace(pat, s, repl, n)
			if e != nil {
				return nil, object.NewRuntimeError(e.Error())
			}

			return []object.Value{object.String(ret), object.Integer(count)}, nil
		}

		return nil, ap.TypeError(2, "string/function/table")
	}
}

func _len(th object.Thread, args ...object.Value) ([]object.Value, *object.RuntimeError) {
	ap := fnutil.NewArgParser(th, args)

	s, err := ap.ToGoString(0)
	if err != nil {
		return nil, err
	}

	return []object.Value{object.Integer(len(s))}, nil
}

func lower(th object.Thread, args ...object.Value) ([]object.Value, *object.RuntimeError) {
	ap := fnutil.NewArgParser(th, args)

	s, err := ap.ToGoString(0)
	if err != nil {
		return nil, err
	}

	s = strings.ToLower(s)

	return []object.Value{object.String(s)}, nil
}

func match(th object.Thread, args ...object.Value) ([]object.Value, *object.RuntimeError) {
	ap := fnutil.NewArgParser(th, args)

	s, err := ap.ToGoString(0)
	if err != nil {
		return nil, err
	}

	pat, err := ap.ToGoString(1)
	if err != nil {
		return nil, err
	}

	init, err := ap.OptGoInt(2, 1)
	if err != nil {
		return nil, err
	}
	if init < 0 {
		init = len(s) + 1 + init
	}
	if init < 0 {
		init = 1
	}
	if init > len(s) {
		return nil, nil
	}

	s = s[init-1:]

	m, e := pattern.Match(pat, s)
	if e != nil {
		return nil, object.NewRuntimeError(e.Error())
	}

	if len(m.Captures) == 0 {
		return []object.Value{object.String(m.Item)}, nil
	}

	rets := make([]object.Value, len(m.Captures))
	for i, cap := range m.Captures {
		rets[i] = object.String(cap)
	}

	return rets, nil
}

// pack(fmt, v1, v2, ...)
func pack(th object.Thread, args ...object.Value) ([]object.Value, *object.RuntimeError) {
	ap := fnutil.NewArgParser(th, args)

	fmt, err := ap.ToGoString(0)
	if err != nil {
		return nil, err
	}

	opts, err := parsekOpts(ap, fmt)
	if err != nil {
		return nil, err
	}

	p := newPacker(ap, opts)

	s, err := p.pack()
	if err != nil {
		return nil, err
	}

	return []object.Value{object.String(s)}, nil
}

// packsize(fmt)
func packsize(th object.Thread, args ...object.Value) ([]object.Value, *object.RuntimeError) {
	ap := fnutil.NewArgParser(th, args)

	fmt, err := ap.ToGoString(0)
	if err != nil {
		return nil, err
	}

	opts, err := parsekOpts(ap, fmt)
	if err != nil {
		return nil, err
	}

	total := int64(0)

	for _, opt := range opts {
		switch opt.typ {
		case kString, kZeroString:
			return nil, ap.ArgError(0, "variable-length format")
		default:
			size := int64(opt.size + opt.padding)

			if total > limits.MaxInt64-size {
				return nil, ap.ArgError(0, "format result too large")
			}

			total += size
		}
	}

	return []object.Value{object.Integer(total)}, nil
}

func rep(th object.Thread, args ...object.Value) ([]object.Value, *object.RuntimeError) {
	ap := fnutil.NewArgParser(th, args)

	s, err := ap.ToGoString(0)
	if err != nil {
		return nil, err
	}

	n, err := ap.ToGoInt(1)
	if err != nil {
		return nil, err
	}

	sep, err := ap.OptGoString(2, "")
	if err != nil {
		return nil, err
	}

	if n <= 0 {
		return []object.Value{object.String("")}, nil
	}

	size := len(s)*n + len(sep)*(n-1)
	if size < 0 {
		return nil, object.NewRuntimeError("result is overflowed")
	}

	buf := make([]byte, size)

	bp := copy(buf, s+sep)
	for bp < len(buf) {
		copy(buf[bp:], buf[:bp])
		bp *= 2
	}

	return []object.Value{object.String(buf)}, nil
}

func reverse(th object.Thread, args ...object.Value) ([]object.Value, *object.RuntimeError) {
	ap := fnutil.NewArgParser(th, args)

	s, err := ap.ToGoString(0)
	if err != nil {
		return nil, err
	}

	bs := make([]byte, len(s))

	var i int
	var j int
	var r rune
	for {
		r, j = utf8.DecodeLastRuneInString(s[:len(s)-i])
		if r == utf8.RuneError {
			break
		}

		utf8.EncodeRune(bs[i:], r)

		i += j
	}

	return []object.Value{object.String(bs)}, nil
}

func sub(th object.Thread, args ...object.Value) ([]object.Value, *object.RuntimeError) {
	ap := fnutil.NewArgParser(th, args)

	s, err := ap.ToGoString(0)
	if err != nil {
		return nil, err
	}

	i, err := ap.ToGoInt(1)
	if err != nil {
		return nil, err
	}
	if i < 0 {
		i = len(s) + 1 + i
	}
	if i < 1 {
		i = 1
	}

	j, err := ap.OptGoInt(2, -1)
	if err != nil {
		return nil, err
	}
	if j < 0 {
		j = len(s) + 1 + j
	}
	if j > len(s) {
		j = len(s)
	}

	if i > j {
		return []object.Value{object.String("")}, nil
	}

	return []object.Value{object.String(s[i-1 : j])}, nil
}

// unpack(fmt, s, [, pos])
func unpack(th object.Thread, args ...object.Value) ([]object.Value, *object.RuntimeError) {
	ap := fnutil.NewArgParser(th, args)

	fmt, err := ap.ToGoString(0)
	if err != nil {
		return nil, err
	}

	s, err := ap.ToGoString(1)
	if err != nil {
		return nil, err
	}

	pos, err := ap.OptGoInt(2, 1)
	if err != nil {
		return nil, err
	}

	if pos < 0 {
		pos = len(s) + 1 + pos
	}

	if pos <= 0 || len(s) < pos-1 {
		return nil, ap.ArgError(2, "initial position out of string")
	}

	opts, err := parsekOpts(ap, fmt)
	if err != nil {
		return nil, err
	}

	u := newUnpacker(ap, s[pos-1:], opts)

	return u.unpack()
}

func upper(th object.Thread, args ...object.Value) ([]object.Value, *object.RuntimeError) {
	ap := fnutil.NewArgParser(th, args)

	s, err := ap.ToGoString(0)
	if err != nil {
		return nil, err
	}

	s = strings.ToUpper(s)

	return []object.Value{object.String(s)}, nil
}

func Open(th object.Thread, args ...object.Value) ([]object.Value, *object.RuntimeError) {
	m := th.NewTableSize(0, 17)

	m.Set(object.String("byte"), object.GoFunction(_byte))
	m.Set(object.String("char"), object.GoFunction(char))
	m.Set(object.String("dump"), object.GoFunction(dump))
	m.Set(object.String("find"), object.GoFunction(find))
	m.Set(object.String("format"), object.GoFunction(format))
	m.Set(object.String("gmatch"), object.GoFunction(gmatch))
	m.Set(object.String("gsub"), object.GoFunction(gsub))
	m.Set(object.String("len"), object.GoFunction(_len))
	m.Set(object.String("lower"), object.GoFunction(lower))
	m.Set(object.String("match"), object.GoFunction(match))
	m.Set(object.String("pack"), object.GoFunction(pack))
	m.Set(object.String("packsize"), object.GoFunction(packsize))
	m.Set(object.String("rep"), object.GoFunction(rep))
	m.Set(object.String("reverse"), object.GoFunction(reverse))
	m.Set(object.String("sub"), object.GoFunction(sub))
	m.Set(object.String("unpack"), object.GoFunction(unpack))
	m.Set(object.String("upper"), object.GoFunction(upper))

	mt := th.NewTableSize(0, 1)

	mt.Set(object.String("__index"), m)

	th.SetMetatable(object.String(""), mt)

	return []object.Value{m}, nil
}
