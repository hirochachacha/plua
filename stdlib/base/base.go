package base

import (
	"fmt"
	"runtime"
	"strconv"
	"strings"

	"github.com/hirochachacha/plua/compiler"
	"github.com/hirochachacha/plua/internal/compiler_pool"
	"github.com/hirochachacha/plua/internal/errors"
	"github.com/hirochachacha/plua/internal/version"
	"github.com/hirochachacha/plua/object"
	"github.com/hirochachacha/plua/object/fnutil"
)

// assert(v [, message], ...) -> ((v [, message], ...) | panic)
func assert(th object.Thread, args ...object.Value) (rets []object.Value, err *object.RuntimeError) {
	ap := fnutil.NewArgParser(th, args)

	ok, err := ap.ToGoBool(0)
	if err != nil {
		return nil, err
	}

	if !ok {
		val, ok := ap.Get(1)
		if !ok {
			return nil, object.NewRuntimeError("assertion failed!")
		}

		return nil, &object.RuntimeError{Value: val, Level: 1}
	}

	return args, nil
}

// collectgarbage([opt [, arg]])
func collectgarbage(th object.Thread, args ...object.Value) (rets []object.Value, err *object.RuntimeError) {
	ap := fnutil.NewArgParser(th, args)

	opt := "collect"

	if len(args) > 0 {
		opt, err = ap.ToGoString(0)
		if err != nil {
			return nil, err
		}
	}

	switch opt {
	case "collect":
		runtime.GC()
		return []object.Value{object.Integer(0)}, nil
	case "stop":
		return nil, object.NewRuntimeError("not implemented")
	case "restart":
		return nil, object.NewRuntimeError("not implemented")
	case "count":
		m := runtime.MemStats{}
		runtime.ReadMemStats(&m)
		return []object.Value{object.Number(m.Alloc / 1024.0)}, nil
	case "step":
		return nil, object.NewRuntimeError("not implemented")
	case "setpause":
		return nil, object.NewRuntimeError("not implemented")
	case "setstepmul":
		return nil, object.NewRuntimeError("not implemented")
	case "isrunning":
		return []object.Value{object.True}, nil
	}

	return nil, ap.OptionError(0, opt)
}

// dofile([filename]) -> (... | panic)
func dofile(th object.Thread, args ...object.Value) (rets []object.Value, err *object.RuntimeError) {
	ap := fnutil.NewArgParser(th, args)

	fname := ""

	if len(args) > 0 {
		fname, err = ap.ToGoString(0)
		if err != nil {
			return nil, err
		}
	}

	p, err := compiler_pool.CompileFile(fname, 0)
	if err != nil {
		return nil, err
	}

	return th.Call(th.NewClosure(p), nil)
}

// error(message [, level]) -> panic
func _error(th object.Thread, args ...object.Value) (rets []object.Value, err *object.RuntimeError) {
	ap := fnutil.NewArgParser(th, args)

	val, ok := ap.Get(0)
	if !ok {
		return nil, &object.RuntimeError{Value: nil}
	}

	level, err := ap.OptGoInt(1, 1)
	if err != nil {
		return nil, err
	}

	if _, ok := val.(object.String); ok && level > 0 {
		return nil, &object.RuntimeError{Value: val, Level: level}
	}

	return nil, &object.RuntimeError{Value: val}
}

// getmetatable(object) -> Value
func getmetatable(th object.Thread, args ...object.Value) ([]object.Value, *object.RuntimeError) {
	ap := fnutil.NewArgParser(th, args)

	val, err := ap.ToValue(0)
	if err != nil {
		return nil, err
	}

	mt := th.GetMetatable(val)
	if mt != nil {
		if _mt := mt.Get(object.String("__metatable")); _mt != nil {
			return []object.Value{_mt}, nil
		}
	}

	return []object.Value{mt}, nil
}

func inext(th object.Thread, args ...object.Value) ([]object.Value, *object.RuntimeError) {
	ap := fnutil.NewArgParser(th, args)

	t, err := ap.ToTable(0)
	if err != nil {
		return nil, err
	}

	i, err := ap.OptGoInt(1, 0)
	if err != nil {
		return nil, err
	}

	i++

	if i <= t.Len() {
		v := t.Get(object.Integer(i))
		if v == nil {
			return nil, nil
		}
		return []object.Value{object.Integer(i), v}, nil
	}

	return nil, nil
}

func makeINext(tm object.Value) object.GoFunction {
	inext := func(th object.Thread, args ...object.Value) ([]object.Value, *object.RuntimeError) {
		ap := fnutil.NewArgParser(th, args)

		t, err := ap.ToTable(0)
		if err != nil {
			return nil, err
		}

		i, err := ap.OptGoInt(1, 0)
		if err != nil {
			return nil, err
		}

		i++

		rets, err := th.Call(tm, nil, t, object.Integer(i))
		if err != nil {
			return nil, err
		}

		if len(rets) == 0 || rets[0] == nil {
			return nil, nil
		}

		return append([]object.Value{object.Integer(i)}, rets...), nil
	}

	return object.GoFunction(inext)
}

// ipairs(t) -> (inext, t, 0)
func ipairs(th object.Thread, args ...object.Value) ([]object.Value, *object.RuntimeError) {
	ap := fnutil.NewArgParser(th, args)

	t, err := ap.ToValue(0)
	if err != nil {
		return nil, err
	}

	mt := th.GetMetatable(t)

	if mt == nil {
		return []object.Value{object.GoFunction(inext), t, object.Integer(0)}, nil
	}

	if tm := mt.Get(object.TM_IPAIRS); tm != nil {
		return th.Call(tm, nil, args...)
	}

	for i := 0; i < version.MAX_TAG_LOOP; i++ {
		tm := mt.Get(object.TM_INDEX)
		if tm == nil {
			return []object.Value{object.GoFunction(inext), t, object.Integer(0)}, nil
		}

		if object.ToType(tm) == object.TFUNCTION {
			return []object.Value{makeINext(tm), t, object.Integer(0)}, nil
		}

		t = tm
	}

	return nil, object.NewRuntimeError("gettable chain too long; possible loop")
}

// next(t, key) -> (nkey, val)
func next(th object.Thread, args ...object.Value) ([]object.Value, *object.RuntimeError) {
	ap := fnutil.NewArgParser(th, args)

	t, err := ap.ToTable(0)
	if err != nil {
		return nil, err
	}

	var key object.Value
	if len(args) > 1 {
		key = args[1]
	}

	k, v, ok := t.Next(key)
	if !ok {
		return nil, object.NewRuntimeError("invalid key to 'next'")
	}

	if v == nil {
		return nil, nil
	}

	return []object.Value{k, v}, nil
}

// pairs(t) -> (next, t, nil)
func pairs(th object.Thread, args ...object.Value) ([]object.Value, *object.RuntimeError) {
	ap := fnutil.NewArgParser(th, args)

	t, err := ap.ToValue(0)
	if err != nil {
		return nil, err
	}

	if mt := th.GetMetatable(t); mt != nil {
		if tm := mt.Get(object.TM_PAIRS); tm != nil {
			return th.Call(tm, nil, args...)
		}
	}

	return []object.Value{object.GoFunction(next), t, nil}, nil
}

// loadfile(fname [, mode [, env]]]) -> (closure | (nil, errmessage))
func loadfile(th object.Thread, args ...object.Value) ([]object.Value, *object.RuntimeError) {
	ap := fnutil.NewArgParser(th, args)

	fname, err := ap.OptGoString(0, "")
	if err != nil {
		return nil, err
	}

	mode, err := ap.OptGoString(1, "bt")
	if err != nil {
		return nil, err
	}

	var p *object.Proto

	switch mode {
	case "b":
		p, err = compiler_pool.CompileFile(fname, compiler.Binary)
	case "t":
		p, err = compiler_pool.CompileFile(fname, compiler.Text)
	case "bt":
		p, err = compiler_pool.CompileFile(fname, 0)
	default:
		return nil, ap.OptionError(1, mode)
	}

	if err != nil {
		return []object.Value{nil, err.Positioned()}, nil
	}

	cl := th.NewClosure(p)

	if env, ok := ap.Get(2); ok {
		cl.SetUpvalue(0, env)
	}

	return []object.Value{cl}, nil
}

// load(chunk [, chunkname [, mode [, env]]]) -> (closure | (nil, errmessage))
func load(th object.Thread, args ...object.Value) ([]object.Value, *object.RuntimeError) {
	ap := fnutil.NewArgParser(th, args)

	s, err := ap.ToTypes(0, object.TSTRING, object.TFUNCTION)
	if err != nil {
		return nil, err
	}

	var chunk string
	var chunkname string

	switch s := s.(type) {
	case object.String:
		chunk = string(s)

		chunkname, err = ap.OptGoString(1, string(s))
		if err != nil {
			return nil, err
		}
	case object.GoFunction, object.Closure:
		for {
			rets, err := th.Call(s, nil)
			if err != nil {
				return []object.Value{nil, err.Positioned()}, nil
			}
			if len(rets) == 0 || rets[0] == nil {
				break
			}
			s, ok := object.ToGoString(rets[0])
			if !ok {
				return []object.Value{nil, object.String("reader function must return a string")}, nil
			}
			if s == "" {
				break
			}
			chunk += s
		}

		chunkname, err = ap.OptGoString(1, "=(load)")
		if err != nil {
			return nil, err
		}
	}

	mode, err := ap.OptGoString(2, "bt")
	if err != nil {
		return nil, err
	}

	var p *object.Proto
	switch mode {
	case "b":
		p, err = compiler_pool.CompileString(chunk, chunkname, compiler.Binary)
	case "t":
		p, err = compiler_pool.CompileString(chunk, chunkname, compiler.Text)
	case "bt":
		p, err = compiler_pool.CompileString(chunk, chunkname, 0)
	default:
		return nil, ap.OptionError(2, mode)
	}

	if err != nil {
		return []object.Value{nil, err.Positioned()}, nil
	}

	cl := th.NewClosure(p)

	if env, ok := ap.Get(3); ok {
		cl.SetUpvalue(0, env)
	}

	return []object.Value{cl}, nil
}

func _print(th object.Thread, args ...object.Value) ([]object.Value, *object.RuntimeError) {
	if len(args) == 0 {
		fmt.Println("")

		return nil, nil
	}

	tostring := th.Globals().Get(object.String("tostring"))

	for _, arg := range args[:len(args)-1] {
		rets, err := th.Call(tostring, nil, arg)
		if err != nil {
			return nil, err
		}

		if len(rets) == 0 {
			return nil, object.NewRuntimeError("'tostring' must return a string to 'print'")
		}

		s, ok := object.ToGoString(rets[0])
		if !ok {
			return nil, object.NewRuntimeError("'tostring' must return a string to 'print'")
		}

		fmt.Print(s)
		fmt.Print("\t")
	}

	rets, err := th.Call(tostring, nil, args[len(args)-1])
	if err != nil {
		return nil, err
	}

	if len(rets) == 0 {
		return nil, object.NewRuntimeError("'tostring' must return a string to 'print'")
	}

	s, ok := object.ToGoString(rets[0])
	if !ok {
		return nil, object.NewRuntimeError("'tostring' must return a string to 'print'")
	}

	fmt.Println(s)

	return nil, nil
}

func pcall(th object.Thread, args ...object.Value) ([]object.Value, *object.RuntimeError) {
	ap := fnutil.NewArgParser(th, args)

	fn, err := ap.ToTypes(0, object.TNIL, object.TFUNCTION)
	if err != nil {
		return nil, err
	}

	rets, err := th.Call(fn, nil, args[1:]...)
	if err != nil {
		return []object.Value{object.False, err.Positioned()}, nil
	}

	return append([]object.Value{object.True}, rets...), nil
}

func rawequal(th object.Thread, args ...object.Value) ([]object.Value, *object.RuntimeError) {
	ap := fnutil.NewArgParser(th, args)

	x, err := ap.ToValue(0)
	if err != nil {
		return nil, err
	}

	y, err := ap.ToValue(1)
	if err != nil {
		return nil, err
	}

	return []object.Value{object.Boolean(object.Equal(x, y))}, nil
}

func rawlen(th object.Thread, args ...object.Value) ([]object.Value, *object.RuntimeError) {
	ap := fnutil.NewArgParser(th, args)

	x, err := ap.ToTypes(0, object.TSTRING, object.TTABLE)
	if err != nil {
		return nil, err
	}

	switch x := x.(type) {
	case object.String:
		return []object.Value{object.Integer(len(x))}, nil
	case object.Table:
		return []object.Value{object.Integer(x.Len())}, nil
	}

	return nil, nil
}

func rawget(th object.Thread, args ...object.Value) ([]object.Value, *object.RuntimeError) {
	ap := fnutil.NewArgParser(th, args)

	t, err := ap.ToTable(0)
	if err != nil {
		return nil, err
	}

	key, err := ap.ToValue(1)
	if err != nil {
		return nil, err
	}

	return []object.Value{t.Get(key)}, nil
}

func rawset(th object.Thread, args ...object.Value) ([]object.Value, *object.RuntimeError) {
	ap := fnutil.NewArgParser(th, args)

	t, err := ap.ToTable(0)
	if err != nil {
		return nil, err
	}

	key, err := ap.ToValue(1)
	if err != nil {
		return nil, err
	}

	val, err := ap.ToValue(2)
	if err != nil {
		return nil, err
	}

	t.Set(key, val)

	return []object.Value{t}, nil
}

func _select(th object.Thread, args ...object.Value) ([]object.Value, *object.RuntimeError) {
	ap := fnutil.NewArgParser(th, args)

	i, err := ap.ToGoInt(0)
	if err != nil {
		if s, e := ap.ToGoString(0); e == nil {
			if s == "#" {
				return []object.Value{object.Integer(len(args) - 1)}, nil
			}
		}
		return nil, err
	}

	switch {
	case i < 0:
		if len(args) < -i {
			return nil, ap.ArgError(0, "index out of range")
		}

		return args[len(args)+i:], nil
	case i == 0:
		return nil, ap.ArgError(0, "index out of range")
	}

	if len(args) < i {
		return nil, nil
	}

	return args[i:], nil
}

// setmetatable(table, metatable) -> table
func setmetatable(th object.Thread, args ...object.Value) ([]object.Value, *object.RuntimeError) {
	ap := fnutil.NewArgParser(th, args)

	t, err := ap.ToTable(0)
	if err != nil {
		return nil, err
	}

	mt, err := ap.ToTypes(1, object.TNIL, object.TTABLE)
	if err != nil {
		return nil, err
	}

	if old := th.GetMetatable(t); old != nil {
		if old.Get(object.TM_METATABLE) != nil {
			return nil, object.NewRuntimeError("cannot change a protected metatable")
		}
	}

	switch mt := mt.(type) {
	case nil:
		t.SetMetatable(nil)
	case object.Table:
		t.SetMetatable(mt)
	default:
		panic("unreachable")
	}

	return []object.Value{t}, nil
}

func tonumber(th object.Thread, args ...object.Value) ([]object.Value, *object.RuntimeError) {
	ap := fnutil.NewArgParser(th, args)

	switch len(args) {
	case 0, 1:
		val, err := ap.ToValue(0)
		if err != nil {
			return nil, err
		}

		if i, ok := object.ToInteger(val); ok {
			return []object.Value{i}, nil
		}

		if n, ok := object.ToNumber(val); ok {
			if n == 0 { // for ErrRange
				return []object.Value{object.Integer(0)}, nil
			}
			return []object.Value{n}, nil
		}

		return []object.Value{nil}, nil
	}

	base, err := ap.OptGoInt(1, 10)
	if err != nil {
		return nil, err
	}

	if base < 2 || base > 36 {
		return nil, ap.ArgError(1, "base out of range")
	}

	s, err := ap.ToGoString(0)
	if err != nil {
		return nil, err
	}

	if i, err := strconv.ParseInt(strings.TrimSpace(s), base, 64); err == nil {
		return []object.Value{object.Integer(i)}, nil
	}

	return []object.Value{nil}, nil
}

func tostring(th object.Thread, args ...object.Value) ([]object.Value, *object.RuntimeError) {
	ap := fnutil.NewArgParser(th, args)

	val, err := ap.ToValue(0)
	if err != nil {
		return nil, err
	}

	if mt := th.GetMetatable(val); mt != nil {
		if tm := mt.Get(object.TM_TOSTRING); tm != nil {
			return th.Call(tm, nil, val)
		}
	}

	return []object.Value{object.String(object.Repr(val))}, nil
}

func _type(th object.Thread, args ...object.Value) ([]object.Value, *object.RuntimeError) {
	ap := fnutil.NewArgParser(th, args)

	val, err := ap.ToValue(0)
	if err != nil {
		return nil, err
	}

	return []object.Value{object.String(object.ToType(val).String())}, nil
}

func xpcall(th object.Thread, args ...object.Value) ([]object.Value, *object.RuntimeError) {
	ap := fnutil.NewArgParser(th, args)

	f, err := ap.ToFunction(0)
	if err != nil {
		return nil, err
	}

	msgh, err := ap.ToFunction(1)
	if err != nil {
		return nil, err
	}

	rets, err := th.Call(f, nil, args[2:]...)
	if err != nil {
		rets, err = th.Call(msgh, nil, err.Positioned())
		if err != nil {
			return []object.Value{object.False, errors.ErrInErrorHandling.Positioned()}, nil
		}
		return append([]object.Value{object.False}, rets...), nil
	}

	return append([]object.Value{object.True}, rets...), nil
}

func Open(th object.Thread, args ...object.Value) ([]object.Value, *object.RuntimeError) {
	g := th.Globals()

	g.Set(object.String("_G"), g)
	g.Set(object.String("_VERSION"), object.String(version.LUA_NAME))
	g.Set(object.String("assert"), object.GoFunction(assert))
	g.Set(object.String("collectgarbage"), object.GoFunction(collectgarbage))
	g.Set(object.String("dofile"), object.GoFunction(dofile))
	g.Set(object.String("error"), object.GoFunction(_error))
	g.Set(object.String("getmetatable"), object.GoFunction(getmetatable))
	g.Set(object.String("ipairs"), object.GoFunction(ipairs))
	g.Set(object.String("loadfile"), object.GoFunction(loadfile))
	g.Set(object.String("load"), object.GoFunction(load))
	g.Set(object.String("next"), object.GoFunction(next))
	g.Set(object.String("pairs"), object.GoFunction(pairs))
	g.Set(object.String("pcall"), object.GoFunction(pcall))
	g.Set(object.String("print"), object.GoFunction(_print))
	g.Set(object.String("rawequal"), object.GoFunction(rawequal))
	g.Set(object.String("rawlen"), object.GoFunction(rawlen))
	g.Set(object.String("rawget"), object.GoFunction(rawget))
	g.Set(object.String("rawset"), object.GoFunction(rawset))
	g.Set(object.String("select"), object.GoFunction(_select))
	g.Set(object.String("setmetatable"), object.GoFunction(setmetatable))
	g.Set(object.String("tonumber"), object.GoFunction(tonumber))
	g.Set(object.String("tostring"), object.GoFunction(tostring))
	g.Set(object.String("type"), object.GoFunction(_type))
	g.Set(object.String("xpcall"), object.GoFunction(xpcall))

	return []object.Value{g}, nil
}
