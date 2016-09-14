package debug

import (
	"bufio"
	"fmt"
	"os"

	"github.com/hirochachacha/plua/internal/compiler_pool"
	"github.com/hirochachacha/plua/object"
	"github.com/hirochachacha/plua/object/fnutil"
)

// debug()
func debug(th object.Thread, args ...object.Value) ([]object.Value, *object.RuntimeError) {
	stdin := bufio.NewScanner(os.Stdin)

	for {
		_, err := os.Stderr.WriteString("lua_debug> ")
		if err != nil {
			return nil, object.NewRuntimeError(err.Error())
		}

		if !stdin.Scan() {
			if err := stdin.Err(); err != nil {
				return nil, object.NewRuntimeError(err.Error())
			}

			return nil, nil
		}

		line := stdin.Text()
		if line == "cont" {
			return nil, nil
		}

		p, err := compiler_pool.CompileString(line, "=(debug command)")
		if err != nil {
			return nil, object.NewRuntimeError(err.Error())
		}

		rets, e := th.Call(th.NewClosure(p), nil)
		if e != nil {
			return nil, e
		}

		if len(rets) != 0 {
			s := object.Repr(rets[0])

			_, err := fmt.Fprintf(os.Stderr, "\n%s\n", s)
			if err != nil {
				return nil, object.NewRuntimeError(err.Error())
			}
		}
	}
}

// gethook([thread]) -> (fn, mask, count)
func gethook(th object.Thread, args ...object.Value) ([]object.Value, *object.RuntimeError) {
	ap := fnutil.NewArgParser(th, args)

	th1 := ap.GetThread()

	hook, mask, count := th1.GetHook()

	return []object.Value{hook, object.String(mask), object.Integer(count)}, nil
}

// getinfo([thread,] f [, what]) -> debug_info
func getinfo(th object.Thread, args ...object.Value) ([]object.Value, *object.RuntimeError) {
	ap := fnutil.NewArgParser(th, args)

	th1 := ap.GetThread()

	f, err := ap.ToTypes(0, object.TNUMINT, object.TFUNCTION)
	if err != nil {
		return nil, err
	}

	what, err := ap.OptGoString(1, "flnStu")
	if err != nil {
		return nil, err
	}

	opts := map[rune]bool{
		'S': false,
		'l': false,
		'u': false,
		't': false,
		'n': false,
		'L': false,
		'f': false,
	}

	// check what and remove dups
	{
		var w string
		for _, r := range what {
			found, ok := opts[r]
			if !ok {
				return nil, ap.OptionError(1, string(r))
			}
			if !found {
				w += string(r)

				opts[r] = true
			}
		}
		what = w
	}

	var d *object.DebugInfo

	switch f := f.(type) {
	case object.Integer:
		level, ok := object.ToGoInt(f)
		if !ok {
			return nil, nil
		}

		d = th1.GetInfo(level, what)
	case object.GoFunction:
		d = th1.GetInfoByFunc(f, what)
	case object.Closure:
		d = th1.GetInfoByFunc(f, what)
	default:
		panic("unreachable")
	}

	if d == nil {
		return nil, nil
	}

	t := th.NewTableSize(0, 13)

	if opts['S'] {
		t.Set(object.String("source"), object.String(d.Source))
		t.Set(object.String("short_src"), object.String(d.ShortSource))
		t.Set(object.String("linedefined"), object.Integer(d.LineDefined))
		t.Set(object.String("lastlinedefined"), object.Integer(d.LastLineDefined))
		t.Set(object.String("what"), object.String(d.What))
	}

	if opts['l'] {
		t.Set(object.String("currentline"), object.Integer(d.CurrentLine))
	}

	if opts['u'] {
		t.Set(object.String("nups"), object.Integer(d.NUpvalues))
		t.Set(object.String("nparams"), object.Integer(d.NParams))
		if d.IsVararg {
			t.Set(object.String("isvararg"), object.True)
		} else {
			t.Set(object.String("isvararg"), object.False)
		}
	}

	if opts['n'] {
		t.Set(object.String("name"), object.String(d.Name))
		t.Set(object.String("namewhat"), object.String(d.NameWhat))
	}

	if opts['t'] {
		if d.IsTailCall {
			t.Set(object.String("istailcall"), object.True)
		} else {
			t.Set(object.String("istailcall"), object.False)
		}
	}

	if opts['L'] {
		t.Set(object.String("activelines"), d.Lines)
	}

	if opts['f'] {
		t.Set(object.String("func"), d.Func)
	}

	return []object.Value{t}, nil
}

// getlocal([thread,] f, local)
func getlocal(th object.Thread, args ...object.Value) ([]object.Value, *object.RuntimeError) {
	ap := fnutil.NewArgParser(th, args)

	th1 := ap.GetThread()

	f, err := ap.ToTypes(0, object.TNUMINT, object.TFUNCTION)
	if err != nil {
		return nil, err
	}

	local, err := ap.ToGoInt(1)
	if err != nil {
		return nil, err
	}

	var d *object.DebugInfo

	switch f := f.(type) {
	case object.Integer:
		level, ok := object.ToGoInt(f)
		if !ok {
			return nil, nil
		}

		d = th1.GetInfo(level, "")

		if d == nil {
			return nil, ap.ArgError(0, "level out of range")
		}
	case object.GoFunction:
		d = th1.GetInfoByFunc(f, "")

		if d == nil {
			return nil, nil
		}
	case object.Closure:
		d = th1.GetInfoByFunc(f, "")

		if d == nil {
			return nil, nil
		}
	default:
		panic("unreachable")
	}

	name, val := th1.GetLocal(d, local)
	if len(name) == 0 {
		return nil, nil
	}

	return []object.Value{object.String(name), val}, nil
}

func getmetatable(th object.Thread, args ...object.Value) ([]object.Value, *object.RuntimeError) {
	ap := fnutil.NewArgParser(th, args)

	val, err := ap.ToValue(0)
	if err != nil {
		return nil, err
	}

	return []object.Value{th.GetMetatable(val)}, nil
}

func getregistry(th object.Thread, args ...object.Value) ([]object.Value, *object.RuntimeError) {
	return nil, nil
}

// getupvalue(f, up)
func getupvalue(th object.Thread, args ...object.Value) ([]object.Value, *object.RuntimeError) {
	ap := fnutil.NewArgParser(th, args)

	f, err := ap.ToFunction(0)
	if err != nil {
		return nil, err
	}

	up, err := ap.ToGoInt(1)
	if err != nil {
		return nil, err
	}

	switch f := f.(type) {
	case object.GoFunction:
		return nil, nil
	case object.Closure:
		if f.NUpvalues() < up {
			return nil, nil
		}

		name := f.GetUpvalueName(int(up - 1))
		if name == "" {
			name = "(*no name)"
		}

		val := f.GetUpvalue(int(up - 1))

		return []object.Value{object.String(name), val}, nil
	default:
		panic("unreachable")
	}
}

// getuservalue(u)
func getuservalue(th object.Thread, args ...object.Value) ([]object.Value, *object.RuntimeError) {
	if len(args) == 0 {
		return nil, nil
	}
	ud, ok := args[0].(*object.Userdata)
	if !ok {
		return nil, nil
	}

	val, ok := ud.Value.(object.Value)
	if !ok {
		return nil, nil
	}

	return []object.Value{val}, nil
}

// sethook([thread,] hook, mask [, count])
func sethook(th object.Thread, args ...object.Value) ([]object.Value, *object.RuntimeError) {
	ap := fnutil.NewArgParser(th, args)

	th1 := ap.GetThread()

	hook, err := ap.ToFunction(0)
	if err != nil {
		return nil, err
	}

	mask, err := ap.OptGoString(1, "")
	if err != nil {
		return nil, err
	}

	count, err := ap.OptGoInt(2, 0)
	if err != nil {
		return nil, err
	}

	th1.SetHook(hook, mask, count)

	return nil, nil
}

// setlocal([thread,] level, local, value)
func setlocal(th object.Thread, args ...object.Value) ([]object.Value, *object.RuntimeError) {
	ap := fnutil.NewArgParser(th, args)

	th1 := ap.GetThread()

	level, err := ap.ToGoInt(0)
	if err != nil {
		return nil, err
	}

	local, err := ap.ToGoInt(1)
	if err != nil {
		return nil, err
	}

	val, err := ap.ToValue(2)
	if err != nil {
		return nil, err
	}

	d := th1.GetInfo(level, "")
	if d == nil {
		return nil, ap.ArgError(1, "level out of range")
	}

	name := th1.SetLocal(d, local, val)
	if name == "" {
		return nil, nil
	}

	return []object.Value{object.String(name)}, nil
}

// setmetatable(value, table)
func setmetatable(th object.Thread, args ...object.Value) ([]object.Value, *object.RuntimeError) {
	ap := fnutil.NewArgParser(th, args)

	val, err := ap.ToValue(0)
	if err != nil {
		return nil, err
	}

	mt, err := ap.ToTypes(1, object.TNIL, object.TTABLE)
	if err != nil {
		return nil, err
	}

	switch mt := mt.(type) {
	case nil:
		th.SetMetatable(val, nil)
	case object.Table:
		th.SetMetatable(val, mt)
	default:
		panic("unreachable")
	}

	return nil, nil
}

// setupvalue(f, up, value)
func setupvalue(th object.Thread, args ...object.Value) ([]object.Value, *object.RuntimeError) {
	ap := fnutil.NewArgParser(th, args)

	f, err := ap.ToFunction(0)
	if err != nil {
		return nil, err
	}

	up, err := ap.ToGoInt(1)
	if err != nil {
		return nil, err
	}

	val, err := ap.ToValue(2)
	if err != nil {
		return nil, err
	}

	switch f := f.(type) {
	case object.GoFunction:
		return nil, nil
	case object.Closure:
		if f.NUpvalues() < up {
			return nil, nil
		}

		name := f.GetUpvalueName(int(up - 1))
		if name == "" {
			name = "(*no name)"
		}

		f.SetUpvalue(int(up-1), val)

		return []object.Value{object.String(name)}, nil
	default:
		panic("unreachable")
	}
}

func setuservalue(th object.Thread, args ...object.Value) ([]object.Value, *object.RuntimeError) {
	ap := fnutil.NewArgParser(th, args)

	ud, err := ap.ToFullUserdata(0)
	if err != nil {
		return nil, err
	}

	val, err := ap.ToValue(1)
	if err != nil {
		return nil, err
	}

	ud.Value = val

	return nil, nil
}

// TODO
// traceback([thread,] [message [, level]])
func traceback(th object.Thread, args ...object.Value) ([]object.Value, *object.RuntimeError) {
	return nil, object.NewRuntimeError("not implemented")
}

// upvalueid(f, n)
func upvalueid(th object.Thread, args ...object.Value) ([]object.Value, *object.RuntimeError) {
	ap := fnutil.NewArgParser(th, args)

	f, err := ap.ToFunction(0)
	if err != nil {
		return nil, err
	}

	up, err := ap.ToGoInt(1)
	if err != nil {
		return nil, err
	}

	switch f := f.(type) {
	case object.GoFunction:
		return nil, nil
	case object.Closure:
		if up < 1 || up > f.NUpvalues() {
			return nil, ap.ArgError(1, "invalid upvalue index")
		}

		id := f.GetUpvalueId(int(up - 1))

		return []object.Value{id}, nil
	default:
		panic("unreachable")
	}
}

// upvaluejoin(f1, n1, f2, n2)
func upvaluejoin(th object.Thread, args ...object.Value) ([]object.Value, *object.RuntimeError) {
	ap := fnutil.NewArgParser(th, args)

	f1, err := ap.ToClosure(0)
	if err != nil {
		return nil, err
	}

	n1, err := ap.ToGoInt(1)
	if err != nil {
		return nil, err
	}

	f2, err := ap.ToClosure(2)
	if err != nil {
		return nil, err
	}

	n2, err := ap.ToGoInt(3)
	if err != nil {
		return nil, err
	}

	if n1 < 1 || n1 > f1.NUpvalues() {
		return nil, ap.ArgError(1, "invalid upvalue index")
	}

	if n2 < 1 || n2 > f2.NUpvalues() {
		return nil, ap.ArgError(3, "invalid upvalue index")
	}

	f1.UpvalueJoin(n1-1, f2, n2-1)

	return nil, nil
}

func Open(th object.Thread, args ...object.Value) ([]object.Value, *object.RuntimeError) {
	m := th.NewTableSize(0, 16)

	m.Set(object.String("debug"), object.GoFunction(debug))
	m.Set(object.String("gethook"), object.GoFunction(gethook))
	m.Set(object.String("getinfo"), object.GoFunction(getinfo))
	m.Set(object.String("getlocal"), object.GoFunction(getlocal))
	m.Set(object.String("getmetatable"), object.GoFunction(getmetatable))
	m.Set(object.String("getregistry"), object.GoFunction(getregistry))
	m.Set(object.String("getupvalue"), object.GoFunction(getupvalue))
	m.Set(object.String("getuservalue"), object.GoFunction(getuservalue))
	m.Set(object.String("sethook"), object.GoFunction(sethook))
	m.Set(object.String("setlocal"), object.GoFunction(setlocal))
	m.Set(object.String("setmetatable"), object.GoFunction(setmetatable))
	m.Set(object.String("setupvalue"), object.GoFunction(setupvalue))
	m.Set(object.String("setuservalue"), object.GoFunction(setuservalue))
	m.Set(object.String("traceback"), object.GoFunction(traceback))
	m.Set(object.String("upvalueid"), object.GoFunction(upvalueid))
	m.Set(object.String("upvaluejoin"), object.GoFunction(upvaluejoin))

	return []object.Value{m}, nil
}
