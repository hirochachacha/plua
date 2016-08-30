package runtime

import (
	// "fmt"
	"unsafe"

	"github.com/hirochachacha/plua/object"
)

type upvalue struct {
	ctx   *context
	index int
	val   object.Value
}

func (up *upvalue) get() (val object.Value) {
	if up.index == -1 {
		val = up.val
	} else {
		val = up.ctx.stack[up.index]
	}

	return val
}

func (up *upvalue) set(val object.Value) {
	if up.index == -1 {
		up.val = val
	} else {
		up.ctx.stack[up.index] = val
	}
}

type uvlist struct {
	uv   *upvalue
	next *uvlist
}

type closure struct {
	*object.Proto

	closures []object.Closure // cache for children

	upvals []*upvalue
}

func (cl *closure) Type() object.Type {
	return object.TFUNCTION
}

func (cl *closure) Prototype() *object.Proto {
	return cl.Proto
}

func (cl *closure) GetUpvalue(i int) (val object.Value) {
	if i < 0 || i >= len(cl.upvals) {
		return nil
	}

	return cl.upvals[i].get()
}

func (cl *closure) GetUpvalueName(i int) string {
	if i < 0 || i >= len(cl.upvals) {
		return ""
	}

	return cl.Upvalues[i].Name
}

func (cl *closure) GetUpvalueId(i int) object.LightUserdata {
	if i < 0 || i >= len(cl.upvals) {
		return object.LightUserdata{}
	}

	return object.LightUserdata{Pointer: unsafe.Pointer(cl.upvals[i])}
}

func (cl *closure) SetUpvalue(i int, val object.Value) {
	if i < 0 || i >= len(cl.upvals) {
		return
	}

	cl.upvals[i].set(val)
}

func (cl *closure) NUpvalues() int {
	return len(cl.upvals)
}

func (cl *closure) UpvalueJoin(i int, other object.Closure, j int) {
	cl2 := other.(*closure)

	if 0 <= i && i < len(cl.upvals) && 0 <= j && j < len(cl2.upvals) {
		cl.upvals[i] = cl2.upvals[j]
	}
}

func (ctx *context) findOrCreateUpval(index int) (uv *upvalue) {
	if ctx.uvcache == nil {
		uv = &upvalue{ctx: ctx, index: index}

		ctx.uvcache = &uvlist{uv: uv}

		return
	}

	{
		uvcache := ctx.uvcache

		uv = uvcache.uv

		if uv.index == index {
			return
		}

		if uv.index < index {
			uv = &upvalue{ctx: ctx, index: index}

			ctx.uvcache = &uvlist{uv: uv, next: uvcache}

			return
		}
	}

	prev := ctx.uvcache
	uvcache := ctx.uvcache.next

	for ; uvcache != nil; uvcache = uvcache.next {
		uv = uvcache.uv

		if uv.index == index {
			return
		}

		if uv.index < index {
			break
		}

		prev = uvcache
	}

	uv = &upvalue{ctx: ctx, index: index}

	prev.next = &uvlist{uv: uv, next: uvcache}

	return uv
}

func (ctx *context) closeUpvals(level int) {
	if ctx.uvcache == nil {
		return
	}

	var uvcache *uvlist
	var uv *upvalue

	for {
		uvcache = ctx.uvcache

		if uvcache == nil {
			return
		}

		uv = uvcache.uv

		if uv.index < level {
			return
		}

		uv.val = ctx.stack[uv.index]
		uv.index = -1

		ctx.uvcache = uvcache.next

		uvcache.next = nil
	}
}

func (th *thread) newClosure(p *object.Proto) object.Closure {
	var upvals []*upvalue
	if len(p.Upvalues) > 0 {
		upvals = make([]*upvalue, len(p.Upvalues))
		upvals[0] = &upvalue{index: -1, val: th.env.globals}
		for i := range upvals[1:] {
			upvals[1+i] = &upvalue{index: -1}
		}
	}

	cl := &closure{
		Proto:    p,
		upvals:   upvals,
		closures: make([]object.Closure, len(p.Protos)),
	}

	return cl
}

func (th *thread) makeClosure(bx int) object.Closure {
	ctx := th.context

	ci := th.context.ci

	p := ci.Protos[bx]

	cl := ci.closures[bx]
	if cl == nil {
		goto makeNew
	}

	{
		impl := cl.(*closure)

		for i, uv := range p.Upvalues {
			if uv.Instack {
				if impl.upvals[i].index == -1 {
					goto makeAnother
				}
			} else {
				if impl.upvals[i] != ci.upvals[uv.Index] {
					goto makeAnother
				}
			}
		}

		return cl
	}

makeNew:
	{
		cl := &closure{
			Proto:    p,
			upvals:   make([]*upvalue, len(p.Upvalues)),
			closures: make([]object.Closure, len(p.Protos)),
		}

		for i, uv := range p.Upvalues {
			if uv.Instack {
				cl.upvals[i] = ctx.findOrCreateUpval(ci.base + uv.Index)
			} else {
				cl.upvals[i] = ci.upvals[uv.Index]
			}
		}

		ci.closures[bx] = cl

		return cl
	}

makeAnother:
	{
		cl := &closure{
			Proto:    p,
			upvals:   make([]*upvalue, len(p.Upvalues)),
			closures: make([]object.Closure, len(p.Protos)),
		}

		for i, uv := range p.Upvalues {
			if uv.Instack {
				cl.upvals[i] = ctx.findOrCreateUpval(ci.base + uv.Index)
			} else {
				cl.upvals[i] = ci.upvals[uv.Index]
			}
		}

		return cl
	}
}
