package runtime

import (
	"github.com/hirochachacha/plua/internal/limits"
	"github.com/hirochachacha/plua/object"
)

type context struct {
	ci      *callInfo
	ciStack ciStack
	stack   []object.Value

	uvcache *uvlist

	inHook    bool
	hookMask  maskType
	hookFunc  object.Value
	instCount int
	hookCount int
	lastLine  int

	status object.ThreadStatus
	data   interface{} // *object.RuntimeError or []object.Value or nil
	errh   object.Value
	prev   *context
}

func (ctx *context) err() *object.RuntimeError {
	if ctx.status == object.THREAD_ERROR {
		return ctx.data.(*object.RuntimeError)
	}
	return nil
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

func (th *thread) pushContext(stackSize int, isHook bool) {
	th.depth++

	// TODO
	ctx := &context{
		ciStack: make([]callInfo, 1, 16),
		stack:   make([]object.Value, stackSize),
	}

	prev := th.context

	ctx.ci = &ctx.ciStack[0]
	ctx.ci.base = 2
	ctx.ci.sp = 2
	ctx.prev = prev
	ctx.stack[0] = th.env.globals // _ENV

	if isHook {
		ctx.inHook = true
		// inherit information for gethook
		ctx.hookMask = prev.hookMask
		ctx.hookFunc = prev.hookFunc
		ctx.hookCount = prev.hookCount
	}

	th.context = ctx
}

func (th *thread) popContext() *context {
	th.depth--

	ctx := th.context

	th.context = th.context.prev

	return ctx
}

func (ctx *context) stackEnsure(size int) bool {
	sp := ctx.ci.sp

	if sp > int(limits.MaxInt)-size {
		return false
	}

	needed := sp + size

	if needed < len(ctx.stack) {
		return true
	}

	var newsize int

	if len(ctx.stack) > int(limits.MaxInt)/2 {
		newsize = int(limits.MaxInt)
	} else {
		newsize *= 2
		if newsize < needed {
			newsize = needed
		}
	}

	newstack := make([]object.Value, newsize)
	copy(newstack, ctx.stack)
	ctx.stack = newstack

	return true
}
