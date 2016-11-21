package string

import (
	"strings"

	"github.com/hirochachacha/plua/internal/compiler_pool"
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

	isPlain := ap.OptGoBool(3, false)

	if isPlain {
		s = s[init-1:]

		idx := strings.Index(s, pat)
		if idx == -1 {
			return nil, nil
		}

		return []object.Value{object.Integer(idx + init), object.Integer(idx + init + len(pat) - 1)}, nil
	}

	captures, e := pattern.FindIndex(s, pat, init-1)
	if e != nil {
		return nil, object.NewRuntimeError(e.Error())
	}

	if captures == nil {
		return nil, nil
	}

	rets := make([]object.Value, len(captures)+1)

	loc := captures[0]

	rets[0] = object.Integer(loc.Begin + 1)
	rets[1] = object.Integer(loc.End)

	for i, cap := range captures[1:] {
		rets[i+2] = cap.Value(s)
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

	p, e := pattern.Compile(pat)
	if e != nil {
		return nil, object.NewRuntimeError(e.Error())
	}

	off := 0

	fn := func(_ object.Thread, _ ...object.Value) ([]object.Value, *object.RuntimeError) {
		captures := p.FindIndex(s, off)
		if captures == nil {
			return nil, nil
		}

		loc := captures[0]

		off = loc.End
		if off == loc.Begin {
			off++
		}

		if len(captures) == 1 {
			return []object.Value{loc.Value(s)}, nil
		}

		rets := make([]object.Value, len(captures)-1)
		for i, cap := range captures[1:] {
			rets[i] = cap.Value(s)
		}

		return rets, nil
	}

	return []object.Value{object.GoFunction(fn)}, nil
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

	captures, e := pattern.FindIndex(s, pat, init-1)
	if e != nil {
		return nil, object.NewRuntimeError(e.Error())
	}

	if captures == nil {
		return nil, nil
	}

	loc := captures[0]

	if len(captures) == 1 {
		return []object.Value{loc.Value(s)}, nil
	}

	rets := make([]object.Value, len(captures)-1)
	for i, cap := range captures[1:] {
		rets[i] = cap.Value(s)
	}

	return rets, nil
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

	b := make([]byte, len(s))
	for i := 0; i < len(s); i++ {
		b[i] = s[len(s)-1-i]
	}

	return []object.Value{object.String(b)}, nil
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
