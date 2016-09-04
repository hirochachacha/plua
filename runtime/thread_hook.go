package runtime

import (
	"github.com/hirochachacha/plua/object"
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

func (th *thread) onInstruction() *object.RuntimeError {
	ctx := th.context

	if ctx.inHook {
		return nil
	}
	if ctx.hookFunc == nil {
		return nil
	}

	if ctx.hookMask&maskCount != 0 && ctx.hookCount > 0 {
		ctx.instCount++

		if ctx.instCount%ctx.hookCount == 0 {
			if err := th.callHook(hookCount, nil); err != nil {
				return err
			}
		}
	}

	if ctx.hookMask&maskLine != 0 {
		line := ctx.ci.LineInfo[ctx.ci.pc]
		if line == ctx.lastLine {
			return nil
		}
		ctx.lastLine = line

		return th.callHook(hookLine, object.Integer(line))
	}

	return nil
}

func (th *thread) onReturn() *object.RuntimeError {
	ctx := th.context

	if ctx.inHook {
		return nil
	}
	if ctx.hookMask&maskReturn != 0 && ctx.hookFunc != nil {
		return th.callHook(hookReturn, nil)
	}

	return nil
}

func (th *thread) onCall() *object.RuntimeError {
	ctx := th.context

	if ctx.inHook {
		return nil
	}
	if ctx.hookMask&maskCall != 0 && ctx.hookFunc != nil {
		return th.callHook(hookCall, nil)
	}
	return nil
}

func (th *thread) onTailCall() *object.RuntimeError {
	ctx := th.context

	if ctx.inHook {
		return nil
	}
	if ctx.hookMask&maskCall != 0 && ctx.hookFunc != nil {
		return th.callHook(hookTailCall, nil)
	}
	return nil
}

func (th *thread) callHook(typ hookType, arg object.Value) *object.RuntimeError {
	ctx := th.context

	_, err := th.docallv(ctx.hookFunc, object.String(typ.String()), arg)

	return err
}
