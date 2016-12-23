package runtime

import (
	"github.com/hirochachacha/plua/internal/errors"
	"github.com/hirochachacha/plua/object"
)

// call a callable by stack index.
func (th *thread) call(a, nargs, nrets int) (err *object.RuntimeError) {
	ctx := th.context

	f := ctx.ci.base + a

	fn := ctx.stack[f]

	switch fn := fn.(type) {
	case nil:
		return errors.CallError(th, fn)
	case object.GoFunction:
		return th.callGo(fn, f, nargs, nrets, false)
	case object.Closure:
		return th.callLua(fn, f, nargs, nrets)
	}

	tm := th.gettmbyobj(fn, object.TM_CALL)

	if !isFunction(tm) {
		return errors.CallError(th, fn)
	}

	ctx.ci.top = f + 2 + nargs

	if !ctx.growStack(ctx.ci.top) {
		return errors.ErrStackOverflow
	}

	copy(ctx.stack[f+1:], ctx.stack[f:f+1+nargs])

	ctx.stack[f] = tm

	return th.call(a, nargs+1, nrets)
}

// call a go function by stack index, immediately store values.
func (th *thread) callGo(fn object.GoFunction, f, nargs, nrets int, isTailCall bool) (err *object.RuntimeError) {
	ctx := th.context

	ctx.pushFrame(callInfo{
		nrets:      nrets,
		isTailCall: isTailCall,
		base:       f + 1,
		top:        f + 1 + nargs,

		// fake infos
		pc: -1,
	})

	if isTailCall {
		if err := th.onTailCall(); err != nil {
			return err
		}
	} else {
		if err := th.onCall(); err != nil {
			return err
		}
	}

	rets, err := fn(th, ctx.stack[ctx.ci.base:ctx.ci.top]...)
	if err != nil {
		return err
	}

	if err := th.onReturn(); err != nil {
		return err
	}

	if nrets != -1 && nrets < len(rets) {
		rets = rets[:nrets]
	}

	top := ctx.ci.base - 1 + len(rets)

	if !ctx.growStack(top) {
		return errors.ErrStackOverflow
	}

	copy(ctx.stack[ctx.ci.base-1:], rets)

	// clear unused stack
	for r := ctx.ci.base - 1 + nrets; r >= top; r-- {
		ctx.stack[r] = nil
	}

	ctx.popFrame()

	// adjust top
	ctx.ci.top = top

	return nil
}

// call a closure by stack index.
func (th *thread) callLua(c object.Closure, f, nargs, nrets int) (err *object.RuntimeError) {
	ctx := th.context

	cl := c.(*closure)

	ctx.pushFrame(callInfo{
		closure: cl,
		nrets:   nrets,
		base:    f + 1,
		top:     f + 1 + cl.MaxStackSize,
	})

	ci := ctx.ci

	if !ctx.growStack(ci.top) {
		return errors.ErrStackOverflow
	}

	ci.varargs = nil

	if nargs > cl.NParams {
		if cl.IsVararg {
			ci.varargs = dup(ctx.stack[ci.base+cl.NParams : ci.base+nargs])
		}
		for r := ci.base + nargs - 1; r >= ci.base+cl.NParams; r-- {
			ctx.stack[r] = nil
		}
	} else {
		for r := ci.base + cl.NParams - 1; r >= ci.base+nargs; r-- {
			ctx.stack[r] = nil
		}
	}

	if err := th.onCall(); err != nil {
		return err
	}

	return nil
}

// tail call a callable by stack index.
func (th *thread) tailcall(a, nargs int) (err *object.RuntimeError) {
	ctx := th.context

	th.closeUpvals(ctx.ci.base) // closing upvalues

	f := ctx.ci.base + a

	fn := ctx.stack[f]

	switch fn := ctx.stack[f].(type) {
	case nil:
		return errors.CallError(th, fn)
	case object.GoFunction:
		return th.callGo(fn, f, nargs, -1, true)
	case object.Closure:
		return th.tailcallLua(fn, f, nargs)
	}

	tm := th.gettmbyobj(fn, object.TM_CALL)

	if !isFunction(tm) {
		return errors.CallError(th, fn)
	}

	ctx.ci.top = f + 2 + nargs

	if !ctx.growStack(ctx.ci.top) {
		return errors.ErrStackOverflow
	}

	copy(ctx.stack[f+1:], ctx.stack[f:f+1+nargs])

	ctx.stack[f] = tm

	return th.tailcall(a, nargs+1)
}

// tail call a closure by stack index.
func (th *thread) tailcallLua(c object.Closure, f, nargs int) (err *object.RuntimeError) {
	ctx := th.context

	cl := c.(*closure)

	ci := ctx.ci

	ci.pc = 0
	ci.top = ci.base + cl.MaxStackSize
	ci.closure = cl
	ci.isTailCall = true

	if !ctx.growStack(ci.top) {
		return errors.ErrStackOverflow
	}

	copy(ctx.stack[ci.base-1:], ctx.stack[f:f+1+nargs])

	for r := f + nargs; r >= ci.base+nargs; r-- {
		ctx.stack[r] = nil
	}

	ci.varargs = nil

	if nargs > cl.NParams {
		if cl.IsVararg {
			ci.varargs = dup(ctx.stack[ci.base+cl.NParams : ci.base+nargs])
		}
		for r := ci.base + nargs - 1; r >= ci.base+cl.NParams; r-- {
			ctx.stack[r] = nil
		}
	} else {
		for r := ci.base + cl.NParams - 1; r >= ci.base+nargs; r-- {
			ctx.stack[r] = nil
		}
	}

	if err := th.onTailCall(); err != nil {
		return err
	}

	return nil
}

// tforcall a callable by stack index.
func (th *thread) tforcall(a, nrets int) (err *object.RuntimeError) {
	ctx := th.context

	f := ctx.ci.base + a

	fn := ctx.stack[f]

	switch fn := fn.(type) {
	case nil:
		return errors.CallError(th, fn)
	case object.GoFunction:
		copy(ctx.stack[f+3:], ctx.stack[f:f+3])

		return th.callGo(fn, f+3, 2, nrets, false)
	case object.Closure:
		args := ctx.stack[f+1 : f+3]

		rets, err := th.docallLua(fn, args...)
		if err != nil {
			return err
		}

		if nrets != -1 && nrets < len(rets) {
			rets = rets[:nrets]
		}

		ctx.ci.top = f + 3 + nrets

		if !ctx.growStack(ctx.ci.top) {
			return errors.ErrStackOverflow
		}

		copy(ctx.stack[f+3:], rets)

		// clear unused stack
		for r := f + 3 + nrets; r >= f+3+len(rets); r-- {
			ctx.stack[r] = nil
		}

		return nil
	}

	tm := th.gettmbyobj(fn, object.TM_CALL)

	if !isFunction(tm) {
		return errors.CallError(th, fn)
	}

	ctx.ci.top = f + 4

	if !ctx.growStack(ctx.ci.top) {
		return errors.ErrStackOverflow
	}

	copy(ctx.stack[f+1:], ctx.stack[f:f+3])

	ctx.stack[f] = tm

	return th.tforcall(a, nrets)
}

func (th *thread) returnLua(a, nrets int) (rets []object.Value, exit bool) {
	if err := th.onReturn(); err != nil {
		return nil, true
	}

	ctx := th.context

	rets = ctx.stack[ctx.ci.base+a : ctx.ci.base+a+nrets]

	if ctx.ci.nrets != -1 && ctx.ci.nrets < len(rets) {
		rets = rets[:ctx.ci.nrets]
	}

	th.closeUpvals(ctx.ci.base) // closing upvalues

	if ctx.ci.isBottom() {
		ctx.status = object.THREAD_RETURN

		return rets, true
	}

	top := ctx.ci.base - 1 + len(rets)

	// copy result to stack
	copy(ctx.stack[ctx.ci.base-1:], rets)

	// clear unused stack
	for r := ctx.ci.base - 1 + ctx.ci.nrets; r >= top; r-- {
		ctx.stack[r] = nil
	}

	ctx.popFrame()

	// adjust top
	ctx.ci.top = top

	return nil, false
}

// call a callable by values, immediately return values.
func (th *thread) docall(fn object.Value, args ...object.Value) (rets []object.Value, err *object.RuntimeError) {
	switch fn := fn.(type) {
	case nil:
		err := errors.CallError(th, fn)

		th.trackError(err)

		return nil, err
	case object.GoFunction:
		old := th.stack[1]

		rets, err := th.docallGo(fn, args...)

		if err != nil {
			th.trackErrorOnce(err)

			th.stack[1] = old

			th.popFrame()

			th.trackError(err)

			return nil, err
		}

		th.stack[1] = old

		return rets, nil
	case object.Closure:
		return th.docallLua(fn, args...)
	}

	tm := th.gettmbyobj(fn, object.TM_CALL)

	return th.docall(tm, append([]object.Value{fn}, args...)...)
}

// call a go function by values, immediately return values.
func (th *thread) docallGo(fn object.GoFunction, args ...object.Value) (rets []object.Value, err *object.RuntimeError) {
	ctx := th.context

	ctx.pushFrame(callInfo{
		nrets:      -1,
		isTailCall: false,

		// fake infos
		base: 2,
		top:  -1,
		pc:   -1,
	})

	ctx.stack[1] = fn

	if err := th.onCall(); err != nil {
		return nil, err
	}

	rets, err = fn(th, args...)
	if err != nil {
		return nil, err
	}

	if err := th.onReturn(); err != nil {
		return nil, err
	}

	ctx.popFrame()

	return rets, nil
}

// call a closure by values, immediately return values.
func (th *thread) docallLua(c object.Closure, args ...object.Value) (rets []object.Value, err *object.RuntimeError) {
	return th.doExecute(c, args, false)
}

func (th *thread) gettmbyobj(val object.Value, tag object.Value) object.Value {
	mt := th.GetMetatable(val)
	if mt == nil {
		return nil
	}

	return mt.Get(tag)
}
