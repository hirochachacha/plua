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

func (ac *ArgParser) Get(n int) (object.Value, bool) {
	n = n + ac.offset

	if len(ac.args) <= n {
		return nil, false
	}

	return ac.args[n], true
}

func (ac *ArgParser) Set(n int, val object.Value) bool {
	n = n + ac.offset

	if len(ac.args) <= n {
		return false
	}

	ac.args[n] = val

	return true
}

func (ac *ArgParser) Args() []object.Value {
	return ac.args[ac.offset:]
}

func (ac *ArgParser) GetThread() object.Thread {
	if len(ac.args) > ac.offset {
		if th, ok := ac.args[ac.offset].(object.Thread); ok {
			ac.offset++

			return th
		}
	}

	return ac.th
}

func (ac *ArgParser) ArgError(n int, extramsg string) string {
	n = n + ac.offset

	n++

	d := ac.th.GetInfo(0, "n")
	if d == nil {
		return fmt.Sprintf("bad argument #%d (%s)", n, extramsg)
	}

	if d.NameWhat == "method" {
		n--
		if n == 0 {
			return fmt.Sprintf("calling '%s' on bad self (%s)", d.Name, extramsg)
		}
	}

	return fmt.Sprintf("bad argument #%d to '%s' (%s)", n, d.Name, extramsg)
}

func (ac *ArgParser) TypeError(n int, tname string) string {
	arg, ok := ac.Get(n)
	if !ok {
		return ac.ArgError(n, fmt.Sprintf("%s expected, got no value", tname))
	}

	var typearg string

	if field := ac.th.GetMetaField(arg, "__name"); field != nil {
		if _name, ok := field.(object.String); ok {
			typearg = string(_name)
		} else {
			if _, ok := arg.(object.LightUserdata); ok {
				typearg = "light userdata"
			} else {
				typearg = object.ToType(arg).String()
			}
		}
	} else {
		if _, ok := arg.(object.LightUserdata); ok {
			typearg = "light userdata"
		} else {
			typearg = object.ToType(arg).String()
		}
	}

	return ac.ArgError(n, fmt.Sprintf("%s expected, got %s", tname, typearg))
}

func (ac *ArgParser) OptionError(n int, opt string) string {
	return ac.ArgError(n, fmt.Sprintf("invalid option '%s'", opt))
}

func (ac *ArgParser) CheckAny(n int) (object.Value, string) {
	arg, ok := ac.Get(n)
	if !ok {
		return nil, ac.ArgError(n, "value expected")
	}

	return arg, ""
}

func (ac *ArgParser) CheckUserdata(n int, tname string) (object.Value, string) {
	arg, ok := ac.Get(n)
	if !ok {
		return nil, ac.ArgError(n, "userdata expected, got no value")
	}

	switch ud := arg.(type) {
	case object.LightUserdata:
		return ud, ""
	case *object.Userdata:
		if tname != "" {
			if ud.Metatable != ac.th.GetMetatableName(tname) {
				return nil, ac.TypeError(n, tname)
			}
		}

		return ud, ""
	}

	return nil, ac.TypeError(n, "userdata")
}

func (ac *ArgParser) CheckFunction(n int) (object.Value, string) {
	arg, ok := ac.Get(n)
	if !ok {
		return nil, ac.ArgError(n, "function expected, got no value")
	}

	if typ := object.ToType(arg); typ != object.TFUNCTION {
		return nil, ac.TypeError(n, "function")
	}

	return arg, ""
}

func (ac *ArgParser) CheckTypes(n int, typs ...object.Type) (object.Value, string) {
	arg, ok := ac.Get(n)
	if !ok {
		typss := ""
		for _, typ := range typs[:len(typs)-1] {
			typss += typ.String() + " or "
		}
		typss += typs[len(typs)-1].String()

		return nil, ac.ArgError(n, typss+" expected, got no value")
	}

	ok = false

	typ1 := object.ToType(arg)

	for _, typ := range typs {
		if typ == typ1 {
			ok = true

			break
		}
	}

	if !ok {
		typss := ""
		for _, typ := range typs[:len(typs)-1] {
			typss += typ.String() + " or "
		}
		typss += typs[len(typs)-1].String()

		return nil, ac.TypeError(n, typss)
	}

	return arg, ""
}

func (ac *ArgParser) ToInteger(n int) (object.Integer, string) {
	arg, ok := ac.Get(n)
	if !ok {
		return 0, ac.ArgError(n, "integer expected, got no value")
	}

	i, ok := object.ToInteger(arg)
	if !ok {
		return 0, ac.TypeError(n, "integer")
	}

	return i, ""
}

func (ac *ArgParser) ToNumber(n int) (object.Number, string) {
	arg, ok := ac.Get(n)
	if !ok {
		return 0, ac.ArgError(n, "number expected, got no value")
	}

	f, ok := object.ToNumber(arg)
	if !ok {
		return 0, ac.TypeError(n, "number")
	}

	return f, ""
}

func (ac *ArgParser) ToString(n int) (object.String, string) {
	arg, ok := ac.Get(n)
	if !ok {
		return "", ac.ArgError(n, "string expected, got no value")
	}

	s, ok := object.ToString(arg)
	if !ok {
		return "", ac.TypeError(n, "string")
	}

	return s, ""
}

func (ac *ArgParser) ToBoolean(n int) (object.Boolean, string) {
	arg, ok := ac.Get(n)
	if !ok {
		return object.False, ac.ArgError(n, "boolean expected, got no value")
	}

	return object.ToBoolean(arg), ""
}

func (ac *ArgParser) ToLightUserdata(n int) (object.LightUserdata, string) {
	arg, ok := ac.Get(n)
	if !ok {
		return object.LightUserdata{}, ac.ArgError(n, "light userdata expected, got no value")
	}

	lud, ok := arg.(object.LightUserdata)
	if !ok {
		return object.LightUserdata{}, ac.TypeError(n, "light userdata")
	}

	return lud, ""
}

func (ac *ArgParser) ToGoFunction(n int) (object.GoFunction, string) {
	arg, ok := ac.Get(n)
	if !ok {
		return nil, ac.ArgError(n, "go function expected, got no value")
	}

	fn, ok := arg.(object.GoFunction)
	if !ok {
		return nil, ac.TypeError(n, "go function")
	}

	return fn, ""
}

func (ac *ArgParser) ToTable(n int) (object.Table, string) {
	arg, ok := ac.Get(n)
	if !ok {
		return nil, ac.ArgError(n, "table expected, got no value")
	}

	t, ok := arg.(object.Table)
	if !ok {
		return nil, ac.TypeError(n, "table")
	}

	return t, ""
}

func (ac *ArgParser) ToFullUserdata(n int) (*object.Userdata, string) {
	arg, ok := ac.Get(n)
	if !ok {
		return nil, ac.ArgError(n, "full userdata expected, got no value")
	}

	ud, ok := arg.(*object.Userdata)
	if !ok {
		return nil, ac.TypeError(n, "full userdata")
	}

	return ud, ""
}

func (ac *ArgParser) ToClosure(n int) (object.Closure, string) {
	arg, ok := ac.Get(n)
	if !ok {
		return nil, ac.ArgError(n, "lua function expected, got no value")
	}

	cl, ok := arg.(object.Closure)
	if !ok {
		return nil, ac.TypeError(n, "lua function")
	}

	return cl, ""
}

func (ac *ArgParser) ToThread(n int) (object.Thread, string) {
	arg, ok := ac.Get(n)
	if !ok {
		return nil, ac.ArgError(n, "thread expected, got no value")
	}

	th, ok := arg.(object.Thread)
	if !ok {
		return nil, ac.TypeError(n, "thread")
	}

	return th, ""
}

func (ac *ArgParser) ToChannel(n int) (object.Channel, string) {
	arg, ok := ac.Get(n)
	if !ok {
		return nil, ac.ArgError(n, "channel expected, got no value")
	}

	ch, ok := arg.(object.Channel)
	if !ok {
		return nil, ac.TypeError(n, "channel")
	}

	return ch, ""
}

func (ac *ArgParser) ToGoInt(n int) (int, string) {
	arg, ok := ac.Get(n)
	if !ok {
		return 0, ac.ArgError(n, "integer expected, got no value")
	}

	i, ok := object.ToGoInt64(arg)
	if !ok {
		return 0, ac.TypeError(n, "integer")
	}

	if i < limits.MinInt || i > limits.MaxInt {
		return 0, ac.ArgError(n, "integer overflow")
	}

	return int(i), ""
}

func (ac *ArgParser) ToGoInt64(n int) (int64, string) {
	arg, ok := ac.Get(n)
	if !ok {
		return 0, ac.ArgError(n, "integer expected, got no value")
	}

	i, ok := object.ToGoInt64(arg)
	if !ok {
		return 0, ac.TypeError(n, "integer")
	}

	return i, ""
}

func (ac *ArgParser) ToGoFloat64(n int) (float64, string) {
	arg, ok := ac.Get(n)
	if !ok {
		return 0, ac.ArgError(n, "number expected, got no value")
	}

	f, ok := object.ToGoFloat64(arg)
	if !ok {
		return 0, ac.TypeError(n, "number")
	}

	return f, ""
}

func (ac *ArgParser) ToGoString(n int) (string, string) {
	arg, ok := ac.Get(n)
	if !ok {
		return "", ac.ArgError(n, "string expected, got no value")
	}

	s, ok := object.ToGoString(arg)
	if !ok {
		return "", ac.TypeError(n, "string")
	}

	return s, ""
}

func (ac *ArgParser) ToGoBool(n int) (bool, string) {
	arg, ok := ac.Get(n)
	if !ok {
		return false, ac.ArgError(n, "boolean expected, got no value")
	}

	return object.ToGoBool(arg), ""
}

func (ac *ArgParser) OptGoInt64(n int, i int64) (int64, string) {
	arg, ok := ac.Get(n)
	if !ok || arg == nil {
		return i, ""
	}

	i64, ok := object.ToGoInt64(arg)
	if !ok {
		return 0, ac.TypeError(n, "integer")
	}

	return i64, ""
}

func (ac *ArgParser) OptGoInt(n int, i int) (int, string) {
	arg, ok := ac.Get(n)
	if !ok || arg == nil {
		return i, ""
	}

	i64, ok := object.ToGoInt64(arg)
	if !ok {
		return 0, ac.TypeError(n, "integer")
	}

	if i64 < limits.MinInt || i64 > limits.MaxInt {
		return i, ""
	}

	return int(i64), ""
}

func (ac *ArgParser) OptGoFloat64(n int, f float64) (float64, string) {
	arg, ok := ac.Get(n)
	if !ok || arg == nil {
		return f, ""
	}

	f64, ok := object.ToGoFloat64(arg)
	if !ok {
		return 0, ac.TypeError(n, "number")
	}

	return f64, ""
}

func (ac *ArgParser) OptGoString(n int, s string) (string, string) {
	arg, ok := ac.Get(n)
	if !ok || arg == nil {
		return s, ""
	}

	gs, ok := object.ToGoString(arg)
	if !ok {
		return "", ac.TypeError(n, "string")
	}

	return gs, ""
}

func (ac *ArgParser) OptGoBool(n int, b bool) bool {
	arg, ok := ac.Get(n)
	if !ok || arg == nil {
		return b
	}

	return object.ToGoBool(arg)
}
