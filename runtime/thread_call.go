package runtime

import "github.com/hirochachacha/plua/object"

// call a callable by stack index.
func (th *thread) call(a, nargs, nrets int) (err *object.RuntimeError) {
	ctx := th.context

	f := ctx.ci.base + a

	fn := ctx.stack[f]

	switch fn := fn.(type) {
	case nil:
		return th.callError(fn)
	case object.GoFunction:
		return th.callGo(fn, f, nargs, nrets, false)
	case object.Closure:
		return th.callLua(fn, f, nargs, nrets)
	}

	tm := th.gettmbyobj(fn, TM_CALL)

	if !isFunction(tm) {
		return th.callError(fn)
	}

	if !ctx.growStack(1) {
		return errStackOverflow
	}

	ctx.ci.top++

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

	if ctx.hookMask != 0 {
		if isTailCall {
			if err := th.onTailCall(); err != nil {
				return err
			}
		} else {
			if err := th.onCall(); err != nil {
				return err
			}
		}
	}

	rets, err := fn(th, ctx.stack[ctx.ci.base:ctx.ci.top]...)
	if err != nil {
		ctx.popFrame()

		return err
	}

	if nrets != -1 && nrets < len(rets) {
		rets = rets[:nrets]
	}

	if !ctx.growStack(len(rets) - 1 - nargs) {
		return errStackOverflow
	}

	copy(ctx.stack[ctx.ci.base-1:], rets)

	retop := ctx.ci.base - 1 + len(rets)

	// clear unused stack
	for r := ctx.ci.top - 1; r >= retop; r-- {
		ctx.stack[r] = nil
	}

	ctx.popFrame()

	// adjust top
	ctx.ci.top = retop

	if ctx.hookMask != 0 {
		return th.onReturn()
	}

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

	if !ctx.growStack(0) {
		return errStackOverflow
	}

	if nvarargs := nargs - cl.NParams; nvarargs > 0 {
		if cl.IsVararg {
			ci.varargs = make([]object.Value, nvarargs)

			copy(ci.varargs, ctx.stack[ci.base+cl.NParams:ci.base+nargs])
		}

		for r := ci.top - 1; r >= ci.base+cl.NParams; r-- {
			ctx.stack[r] = nil
		}
	} else {
		for r := ci.top - 1; r >= ci.base+nargs; r-- {
			ctx.stack[r] = nil
		}
	}

	if ctx.hookMask != 0 {
		return th.onCall()
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
		return th.callError(fn)
	case object.GoFunction:
		return th.callGo(fn, f, nargs, -1, true)
	case object.Closure:
		return th.tailcallLua(fn, f, nargs)
	}

	tm := th.gettmbyobj(fn, TM_CALL)

	if !isFunction(tm) {
		return th.callError(fn)
	}

	if !ctx.growStack(1) {
		return errStackOverflow
	}

	copy(ctx.stack[f+1:], ctx.stack[f:f+1+nargs])

	ctx.ci.top++

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

	if !ctx.growStack(0) {
		return errStackOverflow
	}

	copy(ctx.stack[ci.base-1:ci.base+cl.NParams], ctx.stack[f:f+1+cl.NParams])

	if nvarargs := nargs - cl.NParams; nvarargs > 0 {
		if cl.IsVararg {
			ci.varargs = make([]object.Value, nvarargs)

			copy(ci.varargs, ctx.stack[f+1+cl.NParams:f+1+nargs])
		}

		for r := ci.top - 1; r >= ci.base+cl.NParams; r-- {
			ctx.stack[r] = nil
		}
	} else {
		for r := ci.top - 1; r >= ci.base+nargs; r-- {
			ctx.stack[r] = nil
		}
	}

	if ctx.hookMask != 0 {
		return th.onTailCall()
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
		return th.callError(fn)
	case object.GoFunction:
		copy(ctx.stack[f+3:], ctx.stack[f:f+3])

		return th.callGo(fn, f+3, 2, nrets, false)
	case object.Closure:
		args := ctx.stack[f+1 : f+3]

		rets, err := th.docallLua(fn, nil, args...)
		if err != nil {
			return err
		}

		if nrets != -1 && nrets < len(rets) {
			rets = rets[:nrets]
		}

		if len(rets) == 0 {
			ctx.stack[f+3] = nil
		} else {
			copy(ctx.stack[f+3:], rets)
		}

		return nil
	}

	tm := th.gettmbyobj(fn, TM_CALL)

	if !isFunction(tm) {
		return th.callError(fn)
	}

	ctx.growStack(1)

	copy(ctx.stack[f+1:], ctx.stack[f:f+3])

	ctx.ci.top++

	ctx.stack[f] = tm

	return th.tforcall(a, nrets)
}

func (th *thread) returnLua(a, nrets int) (rets []object.Value, exit bool) {
	ctx := th.context

	rets = ctx.stack[ctx.ci.base+a : ctx.ci.base+a+nrets]

	if ctx.ci.nrets != -1 && ctx.ci.nrets < len(rets) {
		rets = rets[:ctx.ci.nrets]
	}

	th.closeUpvals(ctx.ci.base) // closing upvalues

	if ctx.ci.isBottom() {
		if ctx.hookMask != 0 {
			if err := th.onReturn(); err != nil {
				return nil, true
			}
		}

		ctx.status = object.THREAD_RETURN

		return rets, true
	}

	// copy result to stack
	copy(ctx.stack[ctx.ci.base-1:], rets)

	retop := ctx.ci.base - 1 + len(rets)

	// clear unused stack
	for r := ctx.ci.top - 1; r >= retop; r-- {
		ctx.stack[r] = nil
	}

	ctx.popFrame()

	// adjust top
	ctx.ci.top = retop

	if ctx.hookMask != 0 {
		if err := th.onReturn(); err != nil {
			return nil, true
		}
	}

	return nil, false
}

// call a callable by values, immediately return values.
func (th *thread) docall(fn, errh object.Value, args ...object.Value) (rets []object.Value, err *object.RuntimeError) {
	switch fn := fn.(type) {
	case nil:
		return th.dohandle(errh, th.callError(fn))
	case object.GoFunction:
		rets, err := th.docallGo(fn, args...)
		if err != nil {
			if errh == nil {
				return nil, err
			}

			return th.dohandle(errh, err)
		}

		return rets, nil
	case object.Closure:
		return th.docallLua(fn, errh, args...)
	}

	tm := th.gettmbyobj(fn, TM_CALL)

	return th.docall(tm, errh, append([]object.Value{fn}, args...)...)
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

	if ctx.hookMask != 0 {
		if err := th.onCall(); err != nil {
			return nil, err
		}
	}

	// see getInfo

	old := ctx.stack[1]

	ctx.stack[1] = fn

	rets, err = fn(th, args...)

	ctx.stack[1] = old

	ctx.popFrame()

	if err != nil {
		return nil, err
	}

	if ctx.hookMask != 0 {
		if err := th.onReturn(); err != nil {
			return nil, err
		}
	}

	return rets, nil
}

// call a closure by values, immediately return values.
func (th *thread) docallLua(c object.Closure, errh object.Value, args ...object.Value) (rets []object.Value, err *object.RuntimeError) {
	return th.doExecute(c, errh, args, false)
}

func (th *thread) dohandle(errh object.Value, err *object.RuntimeError) ([]object.Value, *object.RuntimeError) {
	switch errh := errh.(type) {
	case nil:
		return nil, errInErrorHandling
	case object.GoFunction:
		rets, err := th.docallGo(errh, err.Positioned())
		if err != nil {
			return nil, errInErrorHandling
		}

		return rets, nil
	case object.Closure:
		rets, err := th.docallLua(errh, nil, err.Positioned())
		if err != nil {
			return nil, errInErrorHandling
		}

		return rets, nil
	default:
		panic("unexpected")
	}
}
