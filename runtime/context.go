package runtime

import (
	"github.com/hirochachacha/plua/internal/limits"
	"github.com/hirochachacha/plua/object"
)

type hookState int

const (
	noHook hookState = iota
	isHook
	inHook
)

type context struct {
	ci      *callInfo
	ciStack []callInfo
	stack   []object.Value

	uvcache *uvlist

	hookState hookState

	status object.ThreadStatus
	err    *object.RuntimeError
	errh   object.Value
	prev   *context
}

func (ctx *context) isRoot() bool {
	return ctx.prev == nil
}

func (ctx *context) loadfn(fn object.Value) {
	ctx.stack[ctx.ci.base-1] = fn
}

func (ctx *context) fn(ci *callInfo) object.Value {
	return ctx.stack[ci.base-1]
}

func (th *thread) pushContext(stackSize int, newHook bool) {
	th.depth++

	ctx := &context{
		ciStack: make([]callInfo, 1, 16),
		stack:   make([]object.Value, stackSize),
	}

	prev := th.context

	ctx.ci = &ctx.ciStack[0]
	ctx.ci.base = 2
	ctx.ci.top = 2
	ctx.ci.nrets = -1
	ctx.prev = prev
	ctx.stack[0] = th.env.globals // _ENV
	if newHook {
		ctx.hookState = isHook
	} else {
		if prev != nil && prev.hookState != noHook {
			ctx.hookState = inHook
		}
	}

	th.context = ctx
}

func (th *thread) popContext() *context {
	th.depth--

	ctx := th.context

	th.context = th.context.prev

	return ctx
}

func (ctx *context) pushFrame(ci callInfo) {
	ctx.ciStack = append(ctx.ciStack, ci)
	ctx.ci = &ctx.ciStack[len(ctx.ciStack)-1]
}

func (ctx *context) popFrame() {
	ctx.ciStack = ctx.ciStack[:len(ctx.ciStack)-1]
	ctx.ci = &ctx.ciStack[len(ctx.ciStack)-1]
}

func (ctx *context) growStack(top int) bool {
	if top > int(limits.MaxInt) {
		return false
	}

	if top < len(ctx.stack) {
		return true
	}

	var newsize int

	if len(ctx.stack) > int(limits.MaxInt)/2 {
		newsize = int(limits.MaxInt)
	} else {
		newsize *= 2
		if newsize < top {
			newsize = top
		}
	}

	newstack := make([]object.Value, newsize)
	copy(newstack, ctx.stack)
	ctx.stack = newstack

	return true
}

func dup(stack []object.Value) []object.Value {
	return append([]object.Value(nil), stack...)
}
