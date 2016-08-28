package object

import (
	"github.com/hirochachacha/blua/internal/limits"
)

type ArgChecker struct {
	th     *Thread
	args   []Value
	fail   bool
	offset int
}

func (ac *ArgChecker) Get(n int) Value {
	if len(ac.args) <= n {
		return nil
	}

	return ac.args[n]
}

func (ac *ArgChecker) Set(n int, val Value) {
	if len(ac.args) <= n {
		return
	}

	ac.args[n] = val
}

func (ac *ArgChecker) Args() []Value {
	return ac.args
}

func (ac *ArgChecker) GetThread() *Thread {
	if len(ac.args) > 0 {
		if th, ok := ac.args[0].(*Thread); ok {
			ac.offset = 1

			return th
		}
	}

	return ac.th
}

func (ac *ArgChecker) Error(msg string) {
	if ac.fail {
		return
	}

	ac.th.Error(msg)

	ac.fail = true
}

func (ac *ArgChecker) ArgError(n int, extramsg string) {
	if ac.fail {
		return
	}

	n = n + ac.offset

	ac.th.ArgError(n, extramsg)

	ac.fail = true
}

func (ac *ArgChecker) TypeError(n int, tname string) {
	if ac.fail {
		return
	}

	n = n + ac.offset

	if len(ac.args) <= n {
		ac.th.ArgError(n, tname+" expected, got no value")

		ac.fail = true

		return
	}

	ac.th.TypeError(n, tname, ac.args[n])

	ac.fail = true
}

func (ac *ArgChecker) OptionError(n int, opt string) {
	if ac.fail {
		return
	}

	n = n + ac.offset

	ac.th.OptionError(n, opt)

	ac.fail = true
}

func (ac *ArgChecker) OK() bool {
	return !ac.fail
}

func (ac *ArgChecker) CheckAny(n int) Value {
	if ac.fail {
		return nil
	}

	n = n + ac.offset

	if len(ac.args) <= n {
		ac.th.ArgError(n, "value expected")

		ac.fail = true

		return nil
	}

	return ac.args[n]
}

func (ac *ArgChecker) CheckUserdata(n int, tname string) Value {
	if ac.fail {
		return nil
	}

	n = n + ac.offset

	if len(ac.args) <= n {
		ac.th.ArgError(n, "userdata expected, got no value")

		ac.fail = true

		return nil
	}

	switch ud := ac.args[n].(type) {
	case LightUserdata:
		return ud
	case *Userdata:
		if tname != "" {
			if ud.Metatable() != ac.th.GetMetatableName(tname) {
				ac.th.TypeError(n, tname, ac.args[n])

				ac.fail = true

				return nil
			}
		}

		return ud
	}

	ac.th.TypeError(n, "userdata", ac.args[n])

	ac.fail = true

	return nil
}

func (ac *ArgChecker) CheckFunction(n int) Value {
	if ac.fail {
		return nil
	}

	n = n + ac.offset

	if len(ac.args) <= n {
		ac.th.ArgError(n, "function expected, got no value")

		ac.fail = true

		return nil
	}

	if typ := ToType(ac.args[n]); typ != TFUNCTION {
		ac.th.TypeError(n, "function", ac.args[n])

		ac.fail = true

		return nil
	}

	return ac.args[n]
}

func (ac *ArgChecker) CheckTypes(n int, typs ...Type) Value {
	if ac.fail {
		return nil
	}

	n = n + ac.offset

	if len(ac.args) <= n {
		typss := ""
		for _, typ := range typs[:len(typs)-1] {
			typss += typ.String() + " or "
		}
		typss += typs[len(typs)-1].String()

		ac.th.ArgError(n, typss+" expected, got no value")

		ac.fail = true

		return nil
	}

	var ok bool

	typ1 := ToType(ac.args[n])

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

		ac.th.TypeError(n, typss, ac.args[n])

		ac.fail = true

		return nil
	}

	return ac.args[n]
}

func (ac *ArgChecker) ToInteger(n int) Integer {
	if ac.fail {
		return 0
	}

	n = n + ac.offset

	if len(ac.args) <= n {
		ac.th.ArgError(n, "integer expected, got no value")

		ac.fail = true

		return 0
	}

	i, ok := ToInteger(ac.args[n])
	if !ok {
		ac.th.TypeError(n, "integer", ac.args[n])

		ac.fail = true

		return 0
	}

	return i
}

func (ac *ArgChecker) ToNumber(n int) Number {
	if ac.fail {
		return 0
	}

	n = n + ac.offset

	if len(ac.args) <= n {
		ac.th.ArgError(n, "number expected, got no value")

		ac.fail = true

		return 0
	}

	f, ok := ToNumber(ac.args[n])
	if !ok {
		ac.th.TypeError(n, "number", ac.args[n])

		ac.fail = true

		return 0
	}

	return f
}

func (ac *ArgChecker) ToString(n int) String {
	if ac.fail {
		return ""
	}

	n = n + ac.offset

	if len(ac.args) <= n {
		ac.th.ArgError(n, "string expected, got no value")

		ac.fail = true

		return ""
	}

	s, ok := ToString(ac.args[n])
	if !ok {
		ac.th.TypeError(n, "string", ac.args[n])

		ac.fail = true

		return ""
	}

	return s
}

func (ac *ArgChecker) ToBoolean(n int) Boolean {
	if ac.fail {
		return False
	}

	n = n + ac.offset

	if len(ac.args) <= n {
		ac.th.ArgError(n, "boolean expected, got no value")

		ac.fail = true

		return False
	}

	return ToBoolean(ac.args[n])
}

func (ac *ArgChecker) ToLightUserdata(n int) LightUserdata {
	if ac.fail {
		return LightUserdata{Pointer: nil}
	}

	n = n + ac.offset

	if len(ac.args) <= n {
		ac.th.ArgError(n, "light userdata expected, got no value")

		ac.fail = true

		return LightUserdata{Pointer: nil}
	}

	lud, ok := ac.args[n].(LightUserdata)
	if !ok {
		ac.th.TypeError(n, "light userdata", ac.args[n])

		ac.fail = true

		return LightUserdata{Pointer: nil}
	}

	return lud
}

func (ac *ArgChecker) ToGoFunction(n int) GoFunction {
	if ac.fail {
		return nil
	}

	n = n + ac.offset

	if len(ac.args) <= n {
		ac.th.ArgError(n, "go function expected, got no value")

		ac.fail = true

		return nil
	}

	fn, ok := ac.args[n].(GoFunction)
	if !ok {
		ac.th.TypeError(n, "go function", ac.args[n])

		ac.fail = true

		return nil
	}

	return fn
}

func (ac *ArgChecker) ToTable(n int) *Table {
	if ac.fail {
		return nil
	}

	n = n + ac.offset

	if len(ac.args) <= n {
		ac.th.ArgError(n, "table expected, got no value")

		ac.fail = true

		return nil
	}

	t, ok := ac.args[n].(*Table)
	if !ok {
		ac.th.TypeError(n, "table", ac.args[n])

		ac.fail = true

		return nil
	}

	return t
}

func (ac *ArgChecker) ToFullUserdata(n int) *Userdata {
	if ac.fail {
		return nil
	}

	n = n + ac.offset

	if len(ac.args) <= n {
		ac.th.ArgError(n, "full userdata expected, got no value")

		ac.fail = true

		return nil
	}

	ud, ok := ac.args[n].(*Userdata)
	if !ok {
		ac.th.TypeError(n, "full userdata", ac.args[n])

		ac.fail = true

		return nil
	}

	return ud
}

func (ac *ArgChecker) ToClosure(n int) *Closure {
	if ac.fail {
		return nil
	}

	n = n + ac.offset

	if len(ac.args) <= n {
		ac.th.ArgError(n, "lua function expected, got no value")

		ac.fail = true

		return nil
	}

	cl, ok := ac.args[n].(*Closure)
	if !ok {
		ac.th.TypeError(n, "lua function", ac.args[n])

		ac.fail = true

		return nil
	}

	return cl
}

func (ac *ArgChecker) ToThread(n int) *Thread {
	if ac.fail {
		return nil
	}

	n = n + ac.offset

	if len(ac.args) <= n {
		ac.th.ArgError(n, "thread expected, got no value")

		ac.fail = true

		return nil
	}

	th, ok := ac.args[n].(*Thread)
	if !ok {
		ac.th.TypeError(n, "thread", ac.args[n])

		ac.fail = true

		return nil
	}

	return th
}

func (ac *ArgChecker) ToChannel(n int) *Channel {
	if ac.fail {
		return nil
	}

	if len(ac.args) <= n {
		ac.th.ArgError(n, "channel expected, got no value")

		ac.fail = true

		return nil
	}

	n = n + ac.offset

	ch, ok := ac.args[n].(*Channel)
	if !ok {
		ac.th.TypeError(n, "channel", ac.args[n])

		ac.fail = true

		return nil
	}

	return ch
}

func (ac *ArgChecker) ToGoInt(n int) int {
	if ac.fail {
		return 0
	}

	n = n + ac.offset

	if len(ac.args) <= n {
		ac.th.ArgError(n, "integer expected, got no value")

		ac.fail = true

		return 0
	}

	i, ok := ToGoInt64(ac.args[n])
	if !ok {
		ac.th.TypeError(n, "integer", ac.args[n])

		ac.fail = true

		return 0
	}

	if i < limits.MinInt || i > limits.MaxInt {
		ac.th.ArgError(n, "integer overflow")
	}

	return int(i)
}

func (ac *ArgChecker) ToGoInt64(n int) int64 {
	if ac.fail {
		return 0
	}

	n = n + ac.offset

	if len(ac.args) <= n {
		ac.th.ArgError(n, "integer expected, got no value")

		ac.fail = true

		return 0
	}

	i, ok := ToGoInt64(ac.args[n])
	if !ok {
		ac.th.TypeError(n, "integer", ac.args[n])

		ac.fail = true

		return 0
	}

	return i
}

func (ac *ArgChecker) ToGoFloat64(n int) float64 {
	if ac.fail {
		return 0
	}

	n = n + ac.offset

	if len(ac.args) <= n {
		ac.th.ArgError(n, "number expected, got no value")

		ac.fail = true

		return 0
	}

	f, ok := ToGoFloat64(ac.args[n])
	if !ok {
		ac.th.TypeError(n, "number", ac.args[n])

		ac.fail = true

		return 0
	}

	return f
}

func (ac *ArgChecker) ToGoString(n int) string {
	if ac.fail {
		return ""
	}

	n = n + ac.offset

	if len(ac.args) <= n {
		ac.th.ArgError(n, "string expected, got no value")

		ac.fail = true

		return ""
	}

	s, ok := ToGoString(ac.args[n])
	if !ok {
		ac.th.TypeError(n, "string", ac.args[n])

		ac.fail = true

		return ""
	}

	return s
}

func (ac *ArgChecker) ToGoBool(n int) bool {
	if ac.fail {
		return false
	}

	n = n + ac.offset

	if len(ac.args) <= n {
		ac.th.ArgError(n, "boolean expected, got no value")

		ac.fail = true

		return false
	}

	return ToGoBool(ac.args[n])
}

func (ac *ArgChecker) OptGoInt64(n int, i int64) int64 {
	if ac.fail {
		return i
	}

	n = n + ac.offset

	if len(ac.args) <= n || ac.args[n] == nil {
		return i
	}

	i64, ok := ToGoInt64(ac.args[n])
	if !ok {
		ac.th.TypeError(n, "integer", ac.args[n])

		ac.fail = true

		return i
	}

	return i64
}

func (ac *ArgChecker) OptGoInt(n int, i int) int {
	if ac.fail {
		return i
	}

	n = n + ac.offset

	if len(ac.args) <= n || ac.args[n] == nil {
		return i
	}

	i64, ok := ToGoInt64(ac.args[n])
	if !ok {
		ac.th.TypeError(n, "integer", ac.args[n])

		ac.fail = true

		return i
	}

	if i64 < limits.MinInt || i64 > limits.MaxInt {
		return i
	}

	return int(i64)
}

func (ac *ArgChecker) OptGoFloat64(n int, f float64) float64 {
	if ac.fail {
		return f
	}

	n = n + ac.offset

	if len(ac.args) <= n || ac.args[n] == nil {
		return f
	}

	f64, ok := ToGoFloat64(ac.args[n])
	if !ok {
		ac.th.TypeError(n, "number", ac.args[n])

		ac.fail = true

		return f
	}

	return f64
}

func (ac *ArgChecker) OptGoString(n int, s string) string {
	if ac.fail {
		return s
	}

	n = n + ac.offset

	if len(ac.args) <= n || ac.args[n] == nil {
		return s
	}

	gs, ok := ToGoString(ac.args[n])
	if !ok {
		ac.th.TypeError(n, "string", ac.args[n])

		ac.fail = true

		return s
	}

	return gs
}

func (ac *ArgChecker) OptGoBool(n int, b bool) bool {
	if ac.fail {
		return b
	}

	n = n + ac.offset

	if len(ac.args) <= n || ac.args[n] == nil {
		return b
	}

	return ToGoBool(ac.args[n])
}

func (ac *ArgChecker) IsAny(n int) bool {
	n = n + ac.offset

	if len(ac.args) <= n {
		return false
	}

	return true
}

func (ac *ArgChecker) IsUserdata(n int, tname string) bool {
	n = n + ac.offset

	if len(ac.args) <= n || ac.args[n] == nil {
		return false
	}

	switch ud := ac.args[n].(type) {
	case LightUserdata:
		return true
	case *Userdata:
		if tname != "" {
			if ud.Metatable() != ac.th.GetMetatableName(tname) {
				return false
			}
		}

		return true
	}

	return false
}

func (ac *ArgChecker) IsFunction(n int) bool {
	n = n + ac.offset

	if len(ac.args) <= n || ac.args[n] == nil {
		return false
	}

	if typ := ToType(ac.args[n]); typ != TFUNCTION {
		return false
	}

	return true
}

func (ac *ArgChecker) IsInteger(n int) bool {
	n = n + ac.offset

	if len(ac.args) <= n || ac.args[n] == nil {
		return false
	}

	_, ok := ToInteger(ac.args[n])

	return ok
}

func (ac *ArgChecker) IsNumber(n int) bool {
	n = n + ac.offset

	if len(ac.args) <= n || ac.args[n] == nil {
		return false
	}

	_, ok := ToNumber(ac.args[n])

	return ok
}

func (ac *ArgChecker) IsString(n int) bool {
	n = n + ac.offset

	if len(ac.args) <= n || ac.args[n] == nil {
		return false
	}

	_, ok := ToString(ac.args[n])

	return ok
}

func (ac *ArgChecker) IsBoolean(n int) bool {
	n = n + ac.offset

	if len(ac.args) <= n || ac.args[n] == nil {
		return false
	}

	return true
}

func (ac *ArgChecker) IsLightUserdata(n int) bool {
	n = n + ac.offset

	if len(ac.args) <= n || ac.args[n] == nil {
		return false
	}

	_, ok := ac.args[n].(LightUserdata)

	return ok
}

func (ac *ArgChecker) IsGoFunction(n int) bool {
	n = n + ac.offset

	if len(ac.args) <= n || ac.args[n] == nil {
		return false
	}

	_, ok := ac.args[n].(GoFunction)

	return ok
}

func (ac *ArgChecker) IsTable(n int) bool {
	n = n + ac.offset

	if len(ac.args) <= n || ac.args[n] == nil {
		return false
	}

	_, ok := ac.args[n].(*Table)

	return ok
}

func (ac *ArgChecker) IsFullUserdata(n int) bool {
	n = n + ac.offset

	if len(ac.args) <= n || ac.args[n] == nil {
		return false
	}

	_, ok := ac.args[n].(*Userdata)

	return ok
}

func (ac *ArgChecker) IsClosure(n int) bool {
	n = n + ac.offset

	if len(ac.args) <= n || ac.args[n] == nil {
		return false
	}

	_, ok := ac.args[n].(*Closure)

	return ok
}

func (ac *ArgChecker) IsThread(n int) bool {
	n = n + ac.offset

	if len(ac.args) <= n || ac.args[n] == nil {
		return false
	}

	_, ok := ac.args[n].(*Thread)

	return ok
}

func (ac *ArgChecker) IsChannel(n int) bool {
	n = n + ac.offset

	if len(ac.args) <= n || ac.args[n] == nil {
		return false
	}

	_, ok := ac.args[n].(*Channel)

	return ok
}

func (ac *ArgChecker) IsGoInt(n int) bool {
	n = n + ac.offset

	if len(ac.args) <= n || ac.args[n] == nil {
		return false
	}

	i, ok := ToGoInt64(ac.args[n])
	if !ok {
		return false
	}

	if i < limits.MinInt || i > limits.MaxInt {
		return false
	}

	return true
}
