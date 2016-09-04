package base

import (
	"fmt"
	"runtime"

	"github.com/hirochachacha/plua/internal/compiler_pool"
	"github.com/hirochachacha/plua/internal/version"
	"github.com/hirochachacha/plua/object"
	"github.com/hirochachacha/plua/object/fnutil"
)

// assert(v [, message], ...) -> ((v [, message], ...) | panic)
func Assert(th object.Thread, args ...object.Value) (rets []object.Value, err *object.RuntimeError) {
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

		return nil, &object.RuntimeError{Value: val}
	}

	return args, nil
}

// collectgarbage([opt [, arg]])
func CollectGarbage(th object.Thread, args ...object.Value) (rets []object.Value, err *object.RuntimeError) {
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
		// do nothing
		return []object.Value{object.Integer(0)}, nil
	case "restart":
		// do nothing
		return []object.Value{object.Integer(0)}, nil
	case "count":
		m := runtime.MemStats{}
		runtime.ReadMemStats(&m)
		return []object.Value{object.Number(m.Alloc / 1024.0)}, nil
	case "step":
		runtime.GC()
		return []object.Value{object.True}, nil
	case "setpause":
		// do nothing
		return []object.Value{object.Integer(-1)}, nil
	case "setstepmul":
		return []object.Value{object.Integer(-1)}, nil
	case "isrunning":
		return []object.Value{object.True}, nil
	case "generational":
		// do nothing
	case "incremental":
		// do nothing
	}

	return nil, ap.OptionError(1, opt)
}

// dofile([filename]) -> (... | panic)
func DoFile(th object.Thread, args ...object.Value) (rets []object.Value, err *object.RuntimeError) {
	ap := fnutil.NewArgParser(th, args)

	fname := ""

	if len(args) > 0 {
		fname, err = ap.ToGoString(0)
		if err != nil {
			return nil, err
		}
	}

	p, e := compiler_pool.CompileFile(fname)
	if e != nil {
		return []object.Value{nil, object.String(e.Error())}, nil
	}

	rets, err = th.Call(th.NewClosure(p), nil)
	if err != nil {
		return nil, err
	}

	return rets, nil
}

// error(message [, level]) -> panic
func Error(th object.Thread, args ...object.Value) (rets []object.Value, err *object.RuntimeError) {
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
func GetMetatable(th object.Thread, args ...object.Value) ([]object.Value, *object.RuntimeError) {
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

	k, v, ok := t.INext(i)
	if !ok {
		return nil, object.NewRuntimeError("invalid key to 'inext'")
	}

	if v == nil {
		return nil, nil
	}

	return []object.Value{object.Integer(k), v}, nil
}

func makeINext(tm object.Value) object.GoFunction {
	inext := func(th object.Thread, args ...object.Value) ([]object.Value, *object.RuntimeError) {
		ap := fnutil.NewArgParser(th, args)

		t, err := ap.ToTable(0)
		if err != nil {
			return nil, err
		}

		key, err := ap.OptGoInt(1, 0)
		if err != nil {
			return nil, err
		}

		rets, err := th.Call(tm, nil, t, object.Integer(key+1))
		if err != nil {
			return nil, err
		}

		return rets, nil
	}

	return object.GoFunction(inext)
}

// ipairs(t) -> (inext, t, 0)
func IPairs(th object.Thread, args ...object.Value) ([]object.Value, *object.RuntimeError) {
	ap := fnutil.NewArgParser(th, args)

	t, err := ap.ToValue(0)
	if err != nil {
		return nil, err
	}

	// TODO expose getTable, setTable?
	for i := 0; i < version.MAX_TAG_LOOP; i++ {
		tm := th.GetMetaField(t, "__index")
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
func Next(th object.Thread, args ...object.Value) ([]object.Value, *object.RuntimeError) {
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
func Pairs(th object.Thread, args ...object.Value) ([]object.Value, *object.RuntimeError) {
	ap := fnutil.NewArgParser(th, args)

	t, err := ap.ToValue(0)
	if err != nil {
		return nil, err
	}

	if fn := th.GetMetaField(t, "__pairs"); fn != nil {
		rets, err := th.Call(fn, nil, args...)
		if err != nil {
			return nil, err
		}

		return rets, nil
	}

	t, err = ap.ToTable(0)
	if err != nil {
		return nil, err
	}

	return []object.Value{object.GoFunction(Next), t, nil}, nil
}

// loadfile(fname [, mode [, env]]]) -> (closure | (nil, errmessage))
func LoadFile(th object.Thread, args ...object.Value) ([]object.Value, *object.RuntimeError) {
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
	var e error

	switch mode {
	case "b":
		p, e = compiler_pool.CompileBinaryFile(fname)
	case "t":
		p, e = compiler_pool.CompileTextFile(fname)
	case "bt":
		p, e = compiler_pool.CompileFile(fname)
	default:
		return nil, ap.OptionError(1, mode)
	}

	if e != nil {
		return []object.Value{nil, object.String(e.Error())}, nil
	}

	cl := th.NewClosure(p)

	if env, ok := ap.Get(2); ok {
		cl.SetUpvalue(0, env)
	}

	return []object.Value{cl}, nil
}

// load(chunk [, chunkname [, mode [, env]]]) -> (closure | (nil, errmessage))
func Load(th object.Thread, args ...object.Value) ([]object.Value, *object.RuntimeError) {
	ap := fnutil.NewArgParser(th, args)

	s, err := ap.ToTypes(0, object.TSTRING, object.TFUNCTION)
	if err != nil {
		return nil, err
	}

	chunk := ""

	switch s := s.(type) {
	case object.String:
		chunk = string(s)
	case object.GoFunction, object.Closure:
		for {
			rets, err := th.Call(s, nil)
			if err != nil {
				return []object.Value{nil, err.Value}, nil
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
	}

	chunkname, err := ap.OptGoString(1, "=(load)")
	if err != nil {
		return nil, err
	}

	mode, err := ap.OptGoString(2, "bt")
	if err != nil {
		return nil, err
	}

	var p *object.Proto
	var e error
	switch mode {
	case "b":
		p, e = compiler_pool.CompileBinaryString(chunk)
	case "t":
		p, e = compiler_pool.CompileTextString(chunk, chunkname)
	case "bt":
		p, e = compiler_pool.CompileString(chunk, chunkname)
	default:
		return nil, ap.OptionError(2, mode)
	}

	if e != nil {
		return []object.Value{nil, object.String(e.Error())}, nil
	}

	cl := th.NewClosure(p)

	if env, ok := ap.Get(3); ok {
		cl.SetUpvalue(0, env)
	}

	return []object.Value{cl}, nil
}

func Print(th object.Thread, args ...object.Value) ([]object.Value, *object.RuntimeError) {
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

func PCall(th object.Thread, args ...object.Value) ([]object.Value, *object.RuntimeError) {
	ap := fnutil.NewArgParser(th, args)

	fn, err := ap.ToFunction(0)
	if err != nil {
		return nil, err
	}

	rets, err := th.Call(fn, nil, args[1:]...)
	if err != nil {
		return []object.Value{object.False, err.Positioned()}, nil
	}

	return append([]object.Value{object.True}, rets...), nil
}

func RawEqual(th object.Thread, args ...object.Value) ([]object.Value, *object.RuntimeError) {
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

func RawLen(th object.Thread, args ...object.Value) ([]object.Value, *object.RuntimeError) {
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

func RawGet(th object.Thread, args ...object.Value) ([]object.Value, *object.RuntimeError) {
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

func RawSet(th object.Thread, args ...object.Value) ([]object.Value, *object.RuntimeError) {
	ap := fnutil.NewArgParser(th, args)

	t, err := ap.ToTable(0)
	if err != nil {
		return nil, err
	}

	key, err := ap.ToValue(1)
	if err != nil {
		return nil, err
	}

	val, err := ap.ToValue(1)
	if err != nil {
		return nil, err
	}

	t.Set(key, val)

	return []object.Value{t}, nil
}

func Select(th object.Thread, args ...object.Value) ([]object.Value, *object.RuntimeError) {
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
func SetMetatable(th object.Thread, args ...object.Value) ([]object.Value, *object.RuntimeError) {
	ap := fnutil.NewArgParser(th, args)

	t, err := ap.ToTable(0)
	if err != nil {
		return nil, err
	}

	mt, err := ap.ToTypes(1, object.TNIL, object.TTABLE)
	if err != nil {
		return nil, err
	}

	if th.GetMetaField(t, "__metatable") != nil {
		return nil, object.NewRuntimeError("cannot change a protected metatable")
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

func ToNumber(th object.Thread, args ...object.Value) ([]object.Value, *object.RuntimeError) {
	ap := fnutil.NewArgParser(th, args)

	val, err := ap.ToValue(0)
	if err != nil {
		return nil, err
	}

	if i, ok := object.ToInteger(val); ok {
		return []object.Value{i}, nil
	}

	if n, ok := object.ToNumber(val); ok {
		return []object.Value{n}, nil
	}

	return nil, nil
}

func ToString(th object.Thread, args ...object.Value) ([]object.Value, *object.RuntimeError) {
	ap := fnutil.NewArgParser(th, args)

	val, err := ap.ToValue(0)
	if err != nil {
		return nil, err
	}

	if rets, done := th.CallMetaField(val, "__tostring"); done {
		if len(rets) == 0 {
			return nil, object.NewRuntimeError("'tostring' must return a string to 'print'")
		}

		return []object.Value{object.String(object.Repr(rets[0]))}, nil
	}

	return []object.Value{object.String(object.Repr(val))}, nil
}

func Type(th object.Thread, args ...object.Value) ([]object.Value, *object.RuntimeError) {
	ap := fnutil.NewArgParser(th, args)

	val, err := ap.ToValue(0)
	if err != nil {
		return nil, err
	}

	return []object.Value{object.String(object.ToType(val).String())}, nil
}

func XPCall(th object.Thread, args ...object.Value) ([]object.Value, *object.RuntimeError) {
	ap := fnutil.NewArgParser(th, args)

	f, err := ap.ToFunction(0)
	if err != nil {
		return nil, err
	}

	msgh, err := ap.ToFunction(1)
	if err != nil {
		return nil, err
	}

	rets, err := th.Call(f, msgh, args[2:]...)
	if err != nil {
		return []object.Value{object.False, err.Positioned()}, nil
	}

	return append([]object.Value{object.True}, rets...), nil
}

func Open(th object.Thread, args ...object.Value) ([]object.Value, *object.RuntimeError) {
	g := th.Globals()

	g.Set(object.String("_G"), g)
	g.Set(object.String("_VERSION"), object.String(version.LUA_NAME))
	g.Set(object.String("assert"), object.GoFunction(Assert))
	g.Set(object.String("collectgarbage"), object.GoFunction(CollectGarbage))
	g.Set(object.String("dofile"), object.GoFunction(DoFile))
	g.Set(object.String("error"), object.GoFunction(Error))
	g.Set(object.String("getmetatable"), object.GoFunction(GetMetatable))
	g.Set(object.String("ipairs"), object.GoFunction(IPairs))
	g.Set(object.String("loadfile"), object.GoFunction(LoadFile))
	g.Set(object.String("load"), object.GoFunction(Load))
	g.Set(object.String("next"), object.GoFunction(Next))
	g.Set(object.String("pairs"), object.GoFunction(Pairs))
	g.Set(object.String("pcall"), object.GoFunction(PCall))
	g.Set(object.String("print"), object.GoFunction(Print))
	g.Set(object.String("rawequal"), object.GoFunction(RawEqual))
	g.Set(object.String("rawlen"), object.GoFunction(RawLen))
	g.Set(object.String("rawget"), object.GoFunction(RawGet))
	g.Set(object.String("rawset"), object.GoFunction(RawSet))
	g.Set(object.String("select"), object.GoFunction(Select))
	g.Set(object.String("setmetatable"), object.GoFunction(SetMetatable))
	g.Set(object.String("tonumber"), object.GoFunction(ToNumber))
	g.Set(object.String("tostring"), object.GoFunction(ToString))
	g.Set(object.String("type"), object.GoFunction(Type))
	g.Set(object.String("xpcall"), object.GoFunction(XPCall))

	return []object.Value{g}, nil
}
