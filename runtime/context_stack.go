package runtime

import (
	"github.com/hirochachacha/blua/internal/limits"

	"github.com/hirochachacha/blua/object"
)

func (ctx *context) stackPush(val object.Value) {
	ctx.stack[ctx.ci.sp] = val
	ctx.ci.sp++
}

func (ctx *context) stackEnsure(size int) bool {
	if ctx.ci.sp > int(limits.MaxInt)-size-1 {
		return false
	}

	needed := ctx.ci.sp + size + 1

	if needed < len(ctx.stack) {
		return true
	}

	var newsize int

	if len(ctx.stack) > int(limits.MaxInt) {
		newsize = needed
	} else {
		newsize = len(ctx.stack) * 2
		if newsize < needed {
			newsize = needed
		}
	}

	newstack := make([]object.Value, newsize)
	copy(newstack, ctx.stack)
	ctx.stack = newstack

	return true
}

func (ctx *context) stackIndex(i int) int {
	if i < 0 {
		return ctx.ci.sp + i
	}

	return ctx.ci.base + i
}
