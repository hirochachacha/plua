package object

import (
	"bytes"
	"fmt"
	"strings"

	"github.com/hirochachacha/blua/position"
)

const (
	LEVELS1 = 12
	LEVELS2 = 10
)

func (th *Thread) ErrorLevel(msg interface{}, level int) {
	if err, ok := msg.(error); ok {
		msg = err.Error()
	}

	pos := position.Position{}

	d := th.GetInfo(level, "Sl")
	if d != nil {
		pos.Filename = "@" + d.ShortSource
		pos.Line = d.CurrentLine
	}

	err := &Error{
		Value: th.ValueOf(msg),
		Pos:   pos,
	}

	th.Impl.Propagate(err)
}

func (th *Thread) Error(msg interface{}) {
	th.ErrorLevel(msg, 1)
}

func (th *Thread) Requiref(openf GoFunction, modname string) bool {
	loaded := th.Loaded()

	if mod := loaded.Get(String(modname)); ToGoBool(mod) {
		return false
	}

	rets := openf(th, String(modname))
	if len(rets) == 0 {
		rets = []Value{nil}
	}

	globals := th.Globals()

	loaded.Set(String(modname), rets[0])
	globals.Set(String(modname), rets[0])

	return true
}

func (th *Thread) Repr(val Value) string {
	if rets, done := th.CallMetaField(val, "__tostring"); done {
		if len(rets) == 0 {
			return ""
		}

		return Repr(rets[0])
	}

	return Repr(val)
}

func (th *Thread) NewMetatableNameSize(tname string, alen, mlen int) *Table {
	reg := th.Registry()

	if mt := reg.Get(String(tname)); mt != nil {
		return nil
	}

	mt := th.NewTableSize(alen, mlen)
	mt.Set(String("__name"), String(tname))
	reg.Set(String(tname), mt)

	return mt
}

func (th *Thread) GetMetatableName(tname string) *Table {
	reg := th.Registry()

	mt, ok := reg.Get(String(tname)).(*Table)
	if !ok {
		return nil
	}

	return mt
}

func (th *Thread) GetMetaField(val Value, field string) Value {
	mt := th.GetMetatable(val)
	if mt == nil {
		return nil
	}

	return mt.Get(String(field))
}

func (th *Thread) CallMetaField(val Value, field string) (rets []Value, done bool) {
	if fn := th.GetMetaField(val, field); fn != nil {
		rets, _ := th.Call(fn, val)

		return rets, true
	}

	return nil, false
}

func (th *Thread) ArgError(n int, extramsg string) {
	n++

	if n <= 0 {
		panic("invalid n")
	}

	d := th.GetInfo(0, "n")
	if d == nil {
		th.Error(fmt.Sprintf("bad argument #%d (%s)", n, extramsg))

		return
	}

	if d.NameWhat == "method" {
		n--
		if n == 0 {
			th.Error(fmt.Sprintf("calling '%s' on bad self (%s)", d.Name, extramsg))

			return
		}
	}

	th.Error(fmt.Sprintf("bad argument #%d to '%s' (%s)", n, d.Name, extramsg))
}

func (th *Thread) TypeError(n int, expected string, actual Value) {
	var typearg string

	if field := th.GetMetaField(actual, "__name"); field != nil {
		if _name, ok := field.(String); ok {
			typearg = string(_name)
		} else {
			if _, ok := actual.(LightUserdata); ok {
				typearg = "light userdata"
			} else {
				typearg = ToType(actual).String()
			}
		}
	} else {
		if _, ok := actual.(LightUserdata); ok {
			typearg = "light userdata"
		} else {
			typearg = ToType(actual).String()
		}
	}

	msg := fmt.Sprintf("%s expected, got %s", expected, typearg)

	th.ArgError(n, msg)
}

func (th *Thread) OptionError(n int, opt string) {
	th.ArgError(n, fmt.Sprintf("invalid option '%s'", opt))
}

func (th *Thread) NewArgChecker(args []Value) *ArgChecker {
	return &ArgChecker{
		th:   th,
		args: args,
	}
}

func (th *Thread) Traceback(msg string, level int) string {
	nlevels := countLevel(th)
	mark := 0
	if nlevels > LEVELS1+LEVELS2 {
		mark = LEVELS1
	}

	buf := new(bytes.Buffer)

	if len(msg) > 0 {
		buf.WriteString(msg)
		buf.WriteRune('\n')
	}

	buf.WriteString("stack traceback:")

	var d *DebugInfo
	for {
		d = th.GetInfo(level, "Slnt")
		if d == nil {
			break
		}

		if mark == level {
			buf.WriteString("\n\t...")
			level = nlevels - LEVELS2
		} else {
			buf.WriteString("\n\t")
			buf.WriteString(d.ShortSource)
			buf.WriteRune(':')
			if d.CurrentLine > 0 {
				fmt.Fprintf(buf, "%d:", d.CurrentLine)
			}
			buf.WriteString(" in ")
			buf.WriteString(getFuncName(th, d))
			if d.IsTailCall {
				buf.WriteString("\n\t(...tail calls...)")
			}
		}

		level++
	}

	return buf.String()
}

func countLevel(th *Thread) int {
	li := 1
	le := 1
	for th.GetInfo(le, "") != nil {
		li = le
		le *= 2
	}
	for li < le {
		m := (li + le) / 2
		if th.GetInfo(m, "") != nil {
			li = m + 1
		} else {
			le = m
		}
	}
	return le - 1
}

func getFuncName(th *Thread, d *DebugInfo) string {
	if s, ok := getGlobalFuncName(th, d); ok {
		return fmt.Sprintf("function '%s'", s)
	}

	switch {
	case len(d.NameWhat) != 0:
		return fmt.Sprintf("%s '%s'", d.NameWhat, d.Name)
	case d.What[0] == 'm':
		return "main chunk"
	case d.What[0] != 'C':
		return fmt.Sprintf("function <%s:%d>", d.ShortSource, d.LineDefined)
	}

	return "?"
}

func findField(t *Table, val1 Value, level int) (string, bool) {
	var key, val Value
	if level > 0 {
		for {
			key, val, _ = t.Next(key)
			if key == nil {
				break
			}
			if key, ok := key.(String); ok {
				if Equal(val, val1) {
					return string(key), true
				}
				if t, ok := val.(*Table); ok {
					if suf, ok := findField(t, val1, level-1); ok {
						return string(key) + "." + suf, true
					}
				}
			}
		}
	}

	return "", false
}

func getGlobalFuncName(th *Thread, d *DebugInfo) (string, bool) {
	loaded := th.Registry().Get(String("_LOADED")).(*Table)

	name, ok := findField(loaded, d.Func, 2)
	if ok {
		if strings.HasPrefix(name, "_G.") {
			return name[3:], true
		}
		return name, true
	}

	return "", false
}
