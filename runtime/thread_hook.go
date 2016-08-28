package runtime

import (
	"github.com/hirochachacha/blua/object"
)

type hookType int

func (h hookType) String() string {
	return hookNames[h]
}

const (
	hookCall hookType = iota
	hookReturn
	hookLine
	hookCount
	hookTailCall
)

type maskType uint

const (
	maskCall maskType = 1 << iota
	maskReturn
	maskLine
	maskCount
)

var hookNames = [...]string{
	hookCall:     "call",
	hookReturn:   "return",
	hookLine:     "line",
	hookCount:    "count",
	hookTailCall: "tail call",
}

func (th *thread) onInstruction() bool {
	ctx := th.context

	if ctx.inHook {
		return true
	}
	if ctx.hookFunc == nil {
		return true
	}

	if ctx.hookMask&maskCount != 0 && ctx.hookCount > 0 {
		ctx.instCount++

		if ctx.instCount%ctx.hookCount == 0 {
			if !th.callHook(hookCount, nil) {
				return false
			}
		}
	}

	if ctx.hookMask&maskLine != 0 {
		line := ctx.ci.LineInfo[ctx.ci.pc]
		if line == ctx.lastLine {
			return true
		}
		ctx.lastLine = line

		return th.callHook(hookLine, object.Integer(line))
	}

	return true
}

func (th *thread) onReturn() bool {
	ctx := th.context

	if ctx.inHook {
		return true
	}
	if ctx.hookMask&maskReturn != 0 && ctx.hookFunc != nil {
		return th.callHook(hookReturn, nil)
	}

	return true
}

func (th *thread) onCall() bool {
	ctx := th.context

	if ctx.inHook {
		return true
	}
	if ctx.hookMask&maskCall != 0 && ctx.hookFunc != nil {
		return th.callHook(hookCall, nil)
	}
	return true
}

func (th *thread) onTailCall() bool {
	ctx := th.context

	if ctx.inHook {
		return true
	}
	if ctx.hookMask&maskCall != 0 && ctx.hookFunc != nil {
		return th.callHook(hookTailCall, nil)
	}
	return true
}

func (th *thread) callHook(typ hookType, arg object.Value) bool {
	ctx := th.context

	_, ok := th.docallv(ctx.hookFunc, object.String(typ.String()), arg)

	return ok
}
