package runtime

import (
	"fmt"

	"github.com/hirochachacha/plua/object"
)

func (th *thread) Traceback(level int) (tb []*object.StackTrace) {
	switch th.status {
	case object.THREAD_ERROR:
		tb = th.err.Traceback
		if level <= 0 && level <= len(tb) {
			return tb[level:]
		}
		return nil
	case object.THREAD_RETURN:
		return nil
	}

	for {
		d := th.GetInfo(level, "Slnt")
		if d == nil {
			break
		}

		tb = append(tb, th.stackTrace(d))

		level++
	}

	return tb
}

func (th *thread) stackTrace(d *object.DebugInfo) *object.StackTrace {
	st := new(object.StackTrace)
	st.Source = d.ShortSource
	st.Line = d.CurrentLine
	st.IsTailCall = d.IsTailCall

	if g := th.getFuncName(d.Func); g != "?" {
		st.Signature = fmt.Sprintf("function '%s'", g)
	} else {
		switch {
		case d.NameWhat != "":
			st.Signature = fmt.Sprintf("%s '%s'", d.NameWhat, d.Name)
		case d.What == "main":
			st.Signature = "main chunk"
		case d.What != "Go":
			st.Signature = fmt.Sprintf("function <%s:%d>", d.ShortSource, d.LineDefined)
		default:
			st.Signature = "?"
		}
	}
	return st
}
