package runtime

import (
	"fmt"

	"github.com/hirochachacha/blua/object"
)

type context struct {
	ci    *callInfo
	stack []object.Value

	uvcache *uvlist

	inHook    bool
	hookMask  maskType
	hookFunc  object.Value
	instCount int
	hookCount int
	lastLine  int

	status object.ThreadStatus
	data   interface{} // *object.Error or []object.Value or nil
	errh   object.Value
	prev   *context
}

func (ctx *context) err() error {
	if ctx.status == object.THREAD_ERROR {
		err := ctx.data.(*object.Error)

		return fmt.Errorf("runtime: %v", err)
	}
	return nil
}

func (ctx *context) isRoot() bool {
	return ctx.prev == nil
}

func (ctx *context) loadfn(fn object.Value) {
	ctx.stack[ctx.ci.base-1] = fn
}

func (ctx *context) fn() object.Value {
	return ctx.stack[ctx.ci.base-1]
}

func (th *thread) pushContext(stackSize int) {
	th.depth++

	th.pushContextWith(make([]object.Value, stackSize))
}

func (th *thread) pushContextWith(stack []object.Value) {
	ctx := &context{
		stack: stack,
		ci: &callInfo{
			base: 2,
			sp:   2,
		},
	}

	ctx.prev = th.context
	ctx.stack[0] = th.env.globals // _ENV

	th.context = ctx
}

func (th *thread) popContext() *context {
	th.depth--

	ctx := th.context

	th.context = th.context.prev

	return ctx
}
