package string

import (
	"strings"
	"unicode/utf8"

	"github.com/hirochachacha/plua/internal/compiler_pool"
	"github.com/hirochachacha/plua/internal/format"
	"github.com/hirochachacha/plua/internal/limits"
	"github.com/hirochachacha/plua/internal/pattern"
	"github.com/hirochachacha/plua/object"
	"github.com/hirochachacha/plua/object/fnutil"
)

var pcache *pattern.Cache

func init() {
	pcache = pattern.NewCache(10)
}

func Byte(th object.Thread, args ...object.Value) ([]object.Value, *object.RuntimeError) {
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
		i = 1
	} else if i < 0 {
		i = len(s) + 1 + i

		if i <= 0 {
			return nil, nil
		}

		if i > len(s) {
			return nil, nil
		}
	} else if i > len(s) {
		return nil, nil
	}

	if j == 0 {
		return nil, nil
	} else if j < 0 {
		j = len(s) + 1 + j

		if j <= 0 {
			return nil, nil
		}

		if j > len(s) {
			return nil, nil
		}
	} else if j > len(s) {
		return nil, nil
	}

	if i > j {
		return nil, nil
	}

	rets := make([]object.Value, j-i+1)
	for k := i - 1; k < j; k++ {
		rets[k] = object.Integer(s[k])
	}

	return rets, nil
}

func Char(th object.Thread, args ...object.Value) ([]object.Value, *object.RuntimeError) {
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

func Dump(th object.Thread, args ...object.Value) ([]object.Value, *object.RuntimeError) {
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

func Find(th object.Thread, args ...object.Value) ([]object.Value, *object.RuntimeError) {
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

	isPlain := ap.OptGoBool(3, false)

	if init == 0 {
		init = 1
	} else if init < 0 {
		init = len(s) + 1 + init

		if init <= 0 {
			return nil, nil
		}

		if init > len(s) {
			return nil, nil
		}
	} else if init > len(s) {
		return nil, nil
	}

	if isPlain {
		idx := strings.Index(s[init-1:], pat)
		if idx == -1 {
			return nil, nil
		}

		return []object.Value{object.Integer(idx + 1), object.Integer(idx + len(pat))}, nil
	}

	pattern, e := pcache.GetOrCompile(pat)
	if err != nil {
		return nil, object.NewRuntimeError(e.Error())
	}

	indices := pattern.FindString(s)

	switch len(indices) {
	case 0:
		return nil, nil
	case 1:
		return []object.Value{object.Integer(indices[0][0] + 1), object.Integer(indices[0][1])}, nil
	}

	rets := make([]object.Value, len(indices)+1)

	first := object.Integer(indices[0][0] + 1)
	second := object.Integer(indices[0][1])

	rets[0] = first
	rets[1] = second
	for i, index := range indices[1:] {
		rets[i+2] = object.String(s[index[0]:index[1]])
	}

	return rets, nil
}

// format(fmt, ...)
func Format(th object.Thread, args ...object.Value) ([]object.Value, *object.RuntimeError) {
	ap := fnutil.NewArgParser(th, args)

	fmt, err := ap.ToGoString(0)
	if err != nil {
		return nil, err
	}

	s, e := format.Format(fmt, args[1:]...)
	if e != nil {
		switch e := e.(type) {
		case *format.ArgError:
			return nil, ap.ArgError(e.ArgNum, e.ExtraMsg)
		case *format.TypeError:
			return nil, ap.TypeError(e.ArgNum+1, e.Expected)
		case *format.OptionError:
			return nil, ap.OptionError(0, e.Opt)
		default:
			return nil, object.NewRuntimeError(e.Error())
		}
	}

	return []object.Value{object.String(s)}, nil
}

// gmatch(s, patter)
func GMatch(th object.Thread, args ...object.Value) ([]object.Value, *object.RuntimeError) {
	ap := fnutil.NewArgParser(th, args)

	s, err := ap.ToGoString(0)
	if err != nil {
		return nil, err
	}

	pat, err := ap.ToGoString(1)
	if err != nil {
		return nil, err
	}

	pattern, e := pcache.GetOrCompile(pat)
	if e != nil {
		return nil, object.NewRuntimeError(e.Error())
	}

	sss := pattern.MatchStringAll(s)

	var i int

	sssLen := len(sss)
	ssLen := 0
	if sssLen > 0 {
		ssLen = len(sss[0])
	}

	fn := func(_ object.Thread, _ ...object.Value) ([]object.Value, *object.RuntimeError) {
		if sssLen <= i {
			return nil, nil
		}

		ss := sss[i]

		i++

		switch ssLen {
		case 0:
			return nil, nil
		case 1:
			return []object.Value{object.String(ss[0])}, nil
		default:
			rets := make([]object.Value, ssLen-1)
			for i, s := range ss[1:] {
				rets[i] = object.String(s)
			}
			return rets, nil
		}
	}

	return []object.Value{object.GoFunction(fn)}, nil
}

// gsub(s, pattern, repl, [, n])
func GSub(th object.Thread, args ...object.Value) ([]object.Value, *object.RuntimeError) {
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

	pattern, e := pcache.GetOrCompile(pat)
	if e != nil {
		return nil, object.NewRuntimeError(e.Error())
	}

	switch repl := repl.(type) {
	case object.GoFunction, object.Closure:
		replfn := func(s string) string {
			rets, err := th.Call(repl, object.String(s))
			if err != nil {
				return ""
			}

			if len(rets) > 0 {
				if s, ok := object.ToGoString(rets[0]); ok {
					return s
				}
			}
			return ""
		}

		ret, err := pattern.ReplaceFuncString(s, replfn, n)
		if err != nil {
			return nil, object.NewRuntimeError(e.Error())
		}

		return []object.Value{object.String(ret)}, nil
	case object.Table:
		replfn := func(s string) string {
			val := repl.Get(object.String(s))
			if s2, ok := object.ToGoString(val); ok {
				return s2
			}
			return ""
		}

		ret, err := pattern.ReplaceFuncString(s, replfn, n)
		if err != nil {
			return nil, object.NewRuntimeError(e.Error())
		}

		return []object.Value{object.String(ret)}, nil
	default:
		if repl, ok := object.ToGoString(repl); ok {
			ret, err := pattern.ReplaceString(s, repl, n)
			if err != nil {
				return nil, object.NewRuntimeError(e.Error())
			}

			return []object.Value{object.String(ret)}, nil
		}

		return nil, ap.TypeError(2, "string/function/table")
	}
}

func Len(th object.Thread, args ...object.Value) ([]object.Value, *object.RuntimeError) {
	ap := fnutil.NewArgParser(th, args)

	s, err := ap.ToGoString(0)
	if err != nil {
		return nil, err
	}

	return []object.Value{object.Integer(len(s))}, nil
}

func Lower(th object.Thread, args ...object.Value) ([]object.Value, *object.RuntimeError) {
	ap := fnutil.NewArgParser(th, args)

	s, err := ap.ToGoString(0)
	if err != nil {
		return nil, err
	}

	s = strings.ToLower(s)

	return []object.Value{object.String(s)}, nil
}

func Match(th object.Thread, args ...object.Value) ([]object.Value, *object.RuntimeError) {
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

	if init == 0 {
		init = 1
	} else if init < 0 {
		init = len(s) + 1 + init

		if init <= 0 {
			return nil, nil
		}

		if init > len(s) {
			return nil, nil
		}
	} else if init > len(s) {
		return nil, nil
	}

	pattern, e := pcache.GetOrCompile(pat)
	if err != nil {
		return nil, object.NewRuntimeError(e.Error())
	}

	ss := pattern.MatchString(s)

	switch len(ss) {
	case 0:
		return nil, nil
	case 1:
		return []object.Value{object.String(ss[0])}, nil
	}

	rets := make([]object.Value, len(ss)-1)

	for i := range rets {
		rets[i] = object.String(ss[i+1])
	}

	return rets, nil
}

// pack(fmt, v1, v2, ...)
func Pack(th object.Thread, args ...object.Value) ([]object.Value, *object.RuntimeError) {
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
func PackSize(th object.Thread, args ...object.Value) ([]object.Value, *object.RuntimeError) {
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

func Repeat(th object.Thread, args ...object.Value) ([]object.Value, *object.RuntimeError) {
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

	buf := make([]byte, len(s)*n+len(sep)*(n-1))

	bp := copy(buf, s+sep)
	for bp < len(buf) {
		copy(buf[bp:], buf[:bp])
		bp *= 2
	}

	return []object.Value{object.String(buf)}, nil
}

func Reverse(th object.Thread, args ...object.Value) ([]object.Value, *object.RuntimeError) {
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

func Sub(th object.Thread, args ...object.Value) ([]object.Value, *object.RuntimeError) {
	ap := fnutil.NewArgParser(th, args)

	s, err := ap.ToGoString(0)
	if err != nil {
		return nil, err
	}

	i, err := ap.ToGoInt(1)
	if err != nil {
		return nil, err
	}

	j, err := ap.OptGoInt(2, -1)
	if err != nil {
		return nil, err
	}

	if i == 0 {
		i = 1
	} else if i < 0 {
		i = len(s) + 1 + i

		if i <= 0 {
			return []object.Value{object.String("")}, nil
		}

		if i > len(s) {
			return []object.Value{object.String("")}, nil
		}
	} else if i > len(s) {
		return []object.Value{object.String("")}, nil
	}

	if j == 0 {
		return []object.Value{object.String("")}, nil
	} else if j < 0 {
		j = len(s) + 1 + j

		if j <= 0 {
			return []object.Value{object.String("")}, nil
		}

		if j > len(s) {
			return []object.Value{object.String("")}, nil
		}
	} else if j > len(s) {
		return []object.Value{object.String("")}, nil
	}

	if i > j {
		return []object.Value{object.String("")}, nil
	}

	return []object.Value{object.String(s[i-1 : j])}, nil
}

// unpack(fmt, s, [, pos])
func Unpack(th object.Thread, args ...object.Value) ([]object.Value, *object.RuntimeError) {
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

func Upper(th object.Thread, args ...object.Value) ([]object.Value, *object.RuntimeError) {
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

	m.Set(object.String("byte"), object.GoFunction(Byte))
	m.Set(object.String("char"), object.GoFunction(Char))
	m.Set(object.String("dump"), object.GoFunction(Dump))
	m.Set(object.String("find"), object.GoFunction(Find))
	m.Set(object.String("format"), object.GoFunction(Format))
	m.Set(object.String("gmatch"), object.GoFunction(GMatch))
	m.Set(object.String("gsub"), object.GoFunction(GSub))
	m.Set(object.String("len"), object.GoFunction(Len))
	m.Set(object.String("lower"), object.GoFunction(Lower))
	m.Set(object.String("match"), object.GoFunction(Match))
	m.Set(object.String("pack"), object.GoFunction(Pack))
	m.Set(object.String("packsize"), object.GoFunction(PackSize))
	m.Set(object.String("rep"), object.GoFunction(Repeat))
	m.Set(object.String("reverse"), object.GoFunction(Reverse))
	m.Set(object.String("sub"), object.GoFunction(Sub))
	m.Set(object.String("unpack"), object.GoFunction(Unpack))
	m.Set(object.String("upper"), object.GoFunction(Upper))

	mt := th.NewTableSize(0, 1)

	mt.Set(object.String("__index"), m)

	th.SetMetatable(object.String(""), mt)

	return []object.Value{m}, nil
}
