package fnutil

import (
	"fmt"

	"github.com/hirochachacha/plua/internal/limits"
	"github.com/hirochachacha/plua/object"
)

type ArgParser struct {
	th     object.Thread
	args   []object.Value
	offset int
}

func NewArgParser(th object.Thread, args []object.Value) *ArgParser {
	return &ArgParser{
		th:   th,
		args: args,
	}
}

func (ap *ArgParser) Args() []object.Value {
	return ap.args[ap.offset:]
}

func (ap *ArgParser) Get(n int) (object.Value, bool) {
	n = n + ap.offset

	if len(ap.args) <= n {
		return nil, false
	}

	return ap.args[n], true
}

func (ap *ArgParser) Set(n int, val object.Value) bool {
	n = n + ap.offset

	if len(ap.args) <= n {
		return false
	}

	ap.args[n] = val

	return true
}

func (ap *ArgParser) GetThread() object.Thread {
	if len(ap.args) > ap.offset {
		if th, ok := ap.args[ap.offset].(object.Thread); ok {
			ap.offset++

			return th
		}
	}

	return ap.th
}

func (ap *ArgParser) ToValue(n int) (object.Value, *object.RuntimeError) {
	arg, ok := ap.Get(n)
	if !ok {
		return nil, ap.ArgError(n, "value expected")
	}

	return arg, nil
}

func (ap *ArgParser) ToUserdata(n int) (object.Value, *object.RuntimeError) {
	arg, ok := ap.Get(n)
	if !ok {
		return nil, ap.ArgError(n, "userdata expected, got no value")
	}

	switch ud := arg.(type) {
	case object.LightUserdata:
		return ud, nil
	case *object.Userdata:
		return ud, nil
	}

	return nil, ap.TypeError(n, "userdata")
}

func (ap *ArgParser) ToFunction(n int) (object.Value, *object.RuntimeError) {
	arg, ok := ap.Get(n)
	if !ok {
		return nil, ap.ArgError(n, "function expected, got no value")
	}

	if typ := object.ToType(arg); typ != object.TFUNCTION {
		return nil, ap.TypeError(n, "function")
	}

	return arg, nil
}

func (ap *ArgParser) ToFunctionOrNil(n int) (object.Value, *object.RuntimeError) {
	arg, ok := ap.Get(n)
	if !ok {
		return nil, ap.ArgError(n, "function expected, got no value")
	}

	if typ := object.ToType(arg); typ != object.TFUNCTION && typ != object.TNIL {
		return nil, ap.TypeError(n, "function or nil")
	}

	return arg, nil
}

func (ap *ArgParser) ToTypes(n int, types ...object.Type) (object.Value, *object.RuntimeError) {
	val, ok := ap.Get(n)
	if !ok {
		typess := ""
		for _, typ := range types[:len(types)-1] {
			typess += typ.String() + " or "
		}
		typess += types[len(types)-1].String()

		return nil, ap.ArgError(n, typess+" expected, got no value")
	}

	{
		for _, typ := range types {
			switch val.(type) {
			case nil:
				if typ == object.TNIL {
					goto Found
				}
			case object.Integer:
				if typ == object.TNUMINT || typ == object.TNUMBER {
					goto Found
				}
			case object.Number:
				if typ == object.TNUMBER {
					goto Found
				}
			case object.String:
				if typ == object.TSTRING {
					goto Found
				}
			case object.Boolean:
				if typ == object.TBOOLEAN {
					goto Found
				}
			case object.LightUserdata:
				if typ == object.TUSERDATA {
					goto Found
				}
			case object.GoFunction:
				if typ == object.TFUNCTION {
					goto Found
				}
			case *object.Userdata:
				if typ == object.TUSERDATA {
					goto Found
				}
			case object.Table:
				if typ == object.TTABLE {
					goto Found
				}
			case object.Closure:
				if typ == object.TFUNCTION {
					goto Found
				}
			case object.Thread:
				if typ == object.TTHREAD {
					goto Found
				}
			}
		}

		typess := ""
		for _, typ := range types[:len(types)-1] {
			typess += typ.String() + " or "
		}
		typess += types[len(types)-1].String()

		return nil, ap.TypeError(n, typess)
	}
Found:
	return val, nil
}

func (ap *ArgParser) ToInteger(n int) (object.Integer, *object.RuntimeError) {
	arg, ok := ap.Get(n)
	if !ok {
		return 0, ap.ArgError(n, "integer expected, got no value")
	}

	i, ok := object.ToInteger(arg)
	if !ok {
		return 0, ap.TypeError(n, "integer")
	}

	return i, nil
}

func (ap *ArgParser) ToNumber(n int) (object.Number, *object.RuntimeError) {
	arg, ok := ap.Get(n)
	if !ok {
		return 0, ap.ArgError(n, "number expected, got no value")
	}

	f, ok := object.ToNumber(arg)
	if !ok {
		return 0, ap.TypeError(n, "number")
	}

	return f, nil
}

func (ap *ArgParser) ToString(n int) (object.String, *object.RuntimeError) {
	arg, ok := ap.Get(n)
	if !ok {
		return "", ap.ArgError(n, "string expected, got no value")
	}

	s, ok := object.ToString(arg)
	if !ok {
		return "", ap.TypeError(n, "string")
	}

	return s, nil
}

func (ap *ArgParser) ToBoolean(n int) (object.Boolean, *object.RuntimeError) {
	arg, ok := ap.Get(n)
	if !ok {
		return object.False, ap.ArgError(n, "boolean expected, got no value")
	}

	return object.ToBoolean(arg), nil
}

func (ap *ArgParser) ToLightUserdata(n int) (object.LightUserdata, *object.RuntimeError) {
	arg, ok := ap.Get(n)
	if !ok {
		return object.LightUserdata{}, ap.ArgError(n, "light userdata expected, got no value")
	}

	lud, ok := arg.(object.LightUserdata)
	if !ok {
		return object.LightUserdata{}, ap.TypeError(n, "light userdata")
	}

	return lud, nil
}

func (ap *ArgParser) ToGoFunction(n int) (object.GoFunction, *object.RuntimeError) {
	arg, ok := ap.Get(n)
	if !ok {
		return nil, ap.ArgError(n, "go function expected, got no value")
	}

	fn, ok := arg.(object.GoFunction)
	if !ok {
		return nil, ap.TypeError(n, "go function")
	}

	return fn, nil
}

func (ap *ArgParser) ToTable(n int) (object.Table, *object.RuntimeError) {
	arg, ok := ap.Get(n)
	if !ok {
		return nil, ap.ArgError(n, "table expected, got no value")
	}

	t, ok := arg.(object.Table)
	if !ok {
		return nil, ap.TypeError(n, "table")
	}

	return t, nil
}

func (ap *ArgParser) ToFullUserdata(n int) (*object.Userdata, *object.RuntimeError) {
	arg, ok := ap.Get(n)
	if !ok {
		return nil, ap.ArgError(n, "full userdata expected, got no value")
	}

	ud, ok := arg.(*object.Userdata)
	if !ok {
		return nil, ap.TypeError(n, "full userdata")
	}

	return ud, nil
}

func (ap *ArgParser) ToClosure(n int) (object.Closure, *object.RuntimeError) {
	arg, ok := ap.Get(n)
	if !ok {
		return nil, ap.ArgError(n, "lua function expected, got no value")
	}

	cl, ok := arg.(object.Closure)
	if !ok {
		return nil, ap.TypeError(n, "lua function")
	}

	return cl, nil
}

func (ap *ArgParser) ToThread(n int) (object.Thread, *object.RuntimeError) {
	arg, ok := ap.Get(n)
	if !ok {
		return nil, ap.ArgError(n, "thread expected, got no value")
	}

	th, ok := arg.(object.Thread)
	if !ok {
		return nil, ap.TypeError(n, "thread")
	}

	return th, nil
}

func (ap *ArgParser) ToGoInt(n int) (int, *object.RuntimeError) {
	arg, ok := ap.Get(n)
	if !ok {
		return 0, ap.ArgError(n, "integer expected, got no value")
	}

	i, ok := object.ToGoInt64(arg)
	if !ok {
		if _, ok := object.ToNumber(arg); ok {
			return 0, ap.ArgError(n, "number has no integer representation")
		}
		return 0, ap.TypeError(n, "integer")
	}

	if i < limits.MinInt || i > limits.MaxInt {
		return 0, ap.ArgError(n, "integer overflow")
	}

	return int(i), nil
}

func (ap *ArgParser) ToGoInt64(n int) (int64, *object.RuntimeError) {
	arg, ok := ap.Get(n)
	if !ok {
		return 0, ap.ArgError(n, "integer expected, got no value")
	}

	i, ok := object.ToGoInt64(arg)
	if !ok {
		if _, ok := object.ToNumber(arg); ok {
			return 0, ap.ArgError(n, "number has no integer representation")
		}
		return 0, ap.TypeError(n, "integer")
	}

	return i, nil
}

func (ap *ArgParser) ToGoFloat64(n int) (float64, *object.RuntimeError) {
	arg, ok := ap.Get(n)
	if !ok {
		return 0, ap.ArgError(n, "number expected, got no value")
	}

	f, ok := object.ToGoFloat64(arg)
	if !ok {
		return 0, ap.TypeError(n, "number")
	}

	return f, nil
}

func (ap *ArgParser) ToGoString(n int) (string, *object.RuntimeError) {
	arg, ok := ap.Get(n)
	if !ok {
		return "", ap.ArgError(n, "string expected, got no value")
	}

	s, ok := object.ToGoString(arg)
	if !ok {
		return "", ap.TypeError(n, "string")
	}

	return s, nil
}

func (ap *ArgParser) ToGoBool(n int) (bool, *object.RuntimeError) {
	arg, ok := ap.Get(n)
	if !ok {
		return false, ap.ArgError(n, "boolean expected, got no value")
	}

	return object.ToGoBool(arg), nil
}

func (ap *ArgParser) OptGoInt64(n int, i int64) (int64, *object.RuntimeError) {
	arg, ok := ap.Get(n)
	if !ok || arg == nil {
		return i, nil
	}

	i64, ok := object.ToGoInt64(arg)
	if !ok {
		return 0, ap.TypeError(n, "integer")
	}

	return i64, nil
}

func (ap *ArgParser) OptGoInt(n int, i int) (int, *object.RuntimeError) {
	arg, ok := ap.Get(n)
	if !ok || arg == nil {
		return i, nil
	}

	i64, ok := object.ToGoInt64(arg)
	if !ok {
		return 0, ap.TypeError(n, "integer")
	}

	if i64 < limits.MinInt || i64 > limits.MaxInt {
		return i, nil
	}

	return int(i64), nil
}

func (ap *ArgParser) OptGoFloat64(n int, f float64) (float64, *object.RuntimeError) {
	arg, ok := ap.Get(n)
	if !ok || arg == nil {
		return f, nil
	}

	f64, ok := object.ToGoFloat64(arg)
	if !ok {
		return 0, ap.TypeError(n, "number")
	}

	return f64, nil
}

func (ap *ArgParser) OptGoString(n int, s string) (string, *object.RuntimeError) {
	arg, ok := ap.Get(n)
	if !ok || arg == nil {
		return s, nil
	}

	gs, ok := object.ToGoString(arg)
	if !ok {
		return "", ap.TypeError(n, "string")
	}

	return gs, nil
}

func (ap *ArgParser) OptGoBool(n int, b bool) bool {
	arg, ok := ap.Get(n)
	if !ok || arg == nil {
		return b
	}

	return object.ToGoBool(arg)
}

func (ap *ArgParser) ArgError(n int, extramsg string) *object.RuntimeError {
	n = n + ap.offset

	n++

	d := ap.th.GetInfo(0, "n")
	if d == nil {
		return object.NewRuntimeError(fmt.Sprintf("bad argument #%d (%s)", n, extramsg))
	}

	if d.NameWhat == "method" {
		n--
		if n == 0 {
			return object.NewRuntimeError(fmt.Sprintf("calling '%s' on bad self (%s)", d.Name, extramsg))
		}
	}

	if d.Name == "" {
		d.Name = ap.getFuncName(d.Func)
	}

	return object.NewRuntimeError(fmt.Sprintf("bad argument #%d to '%s' (%s)", n, d.Name, extramsg))
}

func (ap *ArgParser) TypeError(n int, tname string) *object.RuntimeError {
	arg, ok := ap.Get(n)
	if !ok {
		return ap.ArgError(n, fmt.Sprintf("%s expected, got no value", tname))
	}
	return ap.ArgError(n, fmt.Sprintf("%s expected, got %s", tname, typeName(ap.th, arg)))
}

func (ap *ArgParser) OptionError(n int, opt string) *object.RuntimeError {
	return ap.ArgError(n, fmt.Sprintf("invalid option '%s'", opt))
}

func (ap *ArgParser) getFuncName(fn object.Value) string {
	loaded := ap.th.Loaded()

	var key object.Value
	var val object.Value
	for {
		key, val, _ = loaded.Next(key)
		if val == nil {
			break
		}

		if modname, ok := key.(object.String); ok {
			if module, ok := val.(object.Table); ok {
				var mkey object.Value
				var mval object.Value
				for {
					mkey, mval, _ = module.Next(mkey)
					if mval == nil {
						break
					}

					if fname, ok := mkey.(object.String); ok {
						if object.Equal(mval, fn) {
							if modname == "_G" {
								return string(fname)
							}
							return string(modname) + "." + string(fname)
						}
					}
				}
			}
		}
	}

	return "?"
}

func typeName(th object.Thread, arg object.Value) string {
	if mt := th.GetMetatable(arg); mt != nil {
		if name := mt.Get(object.TM_NAME); name != nil {
			if name, ok := name.(object.String); ok {
				return string(name)
			}
			if _, ok := arg.(object.LightUserdata); ok {
				return "light userdata"
			}
			return object.ToType(arg).String()
		}
	}
	if _, ok := arg.(object.LightUserdata); ok {
		return "light userdata"
	}
	return object.ToType(arg).String()
}
