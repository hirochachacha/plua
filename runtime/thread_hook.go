package runtime

import "github.com/hirochachacha/plua/object"

type hookType uint

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
	if th.context.hookState != noHook {
		return nil
	}

	if th.hookFunc == nil {
		return nil
	}

	if th.hookMask&maskCount != 0 && th.hookCount > 0 {
		th.instCount++

		if th.instCount%th.hookCount == 0 {
			if err := th.callHook(hookCount, nil); err != nil {
				return err
			}
		}
	}

	if th.hookMask&maskLine != 0 {
		ctx := th.context
		line := ctx.ci.LineInfo[ctx.ci.pc]
		if line == th.lastLine {
			return nil
		}
		th.lastLine = line

		return th.callHook(hookLine, object.Integer(line))
	}

	return nil
}

func (th *thread) onReturn() *object.RuntimeError {
	if th.context.hookState != noHook {
		return nil
	}

	if th.hookFunc == nil {
		return nil
	}

	if th.hookMask&maskReturn != 0 {
		return th.callHook(hookReturn, nil)
	}

	return nil
}

func (th *thread) onCall() *object.RuntimeError {
	if th.context.hookState != noHook {
		return nil
	}

	if th.hookFunc == nil {
		return nil
	}

	if th.hookMask&maskCall != 0 {
		return th.callHook(hookCall, nil)
	}

	return nil
}

func (th *thread) onTailCall() *object.RuntimeError {
	if th.context.hookState != noHook {
		return nil
	}

	if th.hookFunc == nil {
		return nil
	}

	if th.hookMask&maskCall != 0 {
		return th.callHook(hookTailCall, nil)
	}

	return nil
}

func (th *thread) callHook(typ hookType, arg object.Value) (err *object.RuntimeError) {
	event := object.String(typ.String())

	_, err = th.doExecute(th.hookFunc, []object.Value{event, arg}, true)

	return err
}
