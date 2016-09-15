package runtime

import (
	"github.com/hirochachacha/plua/object"
)

type callInfo struct {
	*closure

	nrets int

	base int
	pc   int
	sp   int

	isTailCall bool

	varargs []object.Value
}

func (ci *callInfo) isGoFunction() bool {
	return ci.closure == nil
}

type ciStack []callInfo

func (stack ciStack) top() *callInfo {
	return &stack[len(stack)-1]
}

func (stack ciStack) push(ci callInfo) ciStack {
	return append(stack, ci)
}

func (stack ciStack) pop() ciStack {
	return stack[:len(stack)-1]
}

func (stack ciStack) isLuaBottom() bool {
	return len(stack) == 2
}
