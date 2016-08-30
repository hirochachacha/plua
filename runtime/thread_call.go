package runtime

import (
	"github.com/hirochachacha/plua/object"
)

// call a closure by stack index.
func (th *thread) callLua(c object.Closure, f, nargs, nrets int) bool {
	ctx := th.context

	cl := c.(*closure)

	nvarargs := nargs - cl.NParams

	ci := &callInfo{
		closure: cl,
		nrets:   nrets,
		base:    f + 1,
		sp:      f + 1 + cl.NParams,
		prev:    ctx.ci,
	}

	ctx.ci = ci

	ctx.stackEnsure(cl.MaxStackSize)

	switch {
	case nvarargs == 0:
		// do nothing
	case nvarargs < 0:
		for r := ci.sp - 1; r >= ci.base+nargs; r-- {
			ctx.stack[r] = nil
		}
	case nvarargs > 0:
		if cl.IsVararg {
			ci.varargs = make([]object.Value, nvarargs)

			copy(ci.varargs, ctx.stack[ci.base+cl.NParams:ci.base+nargs])
		}
	}

	if ctx.hookMask != 0 {
		return th.onCall()
	}

	return true
}

// call a closure by values, immediately return values.
func (th *thread) docallvLua(c object.Closure, args ...object.Value) (rets []object.Value, ok bool) {
	return th.doExecute(c, nil, args)
}

// call a closure by values with error handler, immediately return values.
func (th *thread) dopcallvLua(c object.Closure, errh object.Value, args ...object.Value) (rets []object.Value, ok bool) {
	if errh == nil {
		return th.doExecute(c, protect, args)
	}
	return th.doExecute(c, errh, args)
}

// tail call a closure by stack index.
func (th *thread) tailcallLua(c object.Closure, f, nargs int) bool {
	ctx := th.context

	cl := c.(*closure)

	ci := ctx.ci

	maxsp := ci.base + ctx.ci.MaxStackSize

	ci.pc = 0
	ci.sp = ci.base + cl.NParams

	ci.closure = cl

	ci.isTailCall = true

	ctx.stackEnsure(cl.MaxStackSize)

	nvarargs := nargs - cl.NParams

	switch {
	case nvarargs == 0:
		copy(ctx.stack[ci.base-1:ci.base+nargs], ctx.stack[f:f+1+nargs])

		for r := maxsp - 1; r >= ci.base+nargs; r-- {
			ctx.stack[r] = nil
		}
	case nvarargs < 0:
		copy(ctx.stack[ci.base-1:ci.base+nargs], ctx.stack[f:f+1+nargs])

		for r := maxsp - 1; r >= ci.base+nargs; r-- {
			ctx.stack[r] = nil
		}
	case nvarargs > 0:
		copy(ctx.stack[ci.base-1:ci.base+cl.NParams], ctx.stack[f:f+1+cl.NParams])

		if cl.IsVararg {
			ci.varargs = make([]object.Value, nvarargs)

			copy(ci.varargs, ctx.stack[f+1+cl.NParams:f+1+nargs])
		}

		for r := maxsp - 1; r >= ci.base+cl.NParams; r-- {
			ctx.stack[r] = nil
		}
	}

	if ctx.hookMask != 0 {
		return th.onTailCall()
	}

	return true
}

// call a go function by stack index, immediately store values.
func (th *thread) callGo(fn object.GoFunction, f, nargs, nrets int, isTailCall bool) bool {
	ctx := th.context

	sp := f + 1 + nargs

	args := ctx.stack[f+1 : sp]

	ci := &callInfo{
		nrets:      nrets,
		prev:       ctx.ci,
		isTailCall: isTailCall,
		base:       f + 1,
		sp:         sp,

		// fake infos
		pc: -1,
	}

	ctx.ci = ci

	if ctx.hookMask != 0 {
		if isTailCall {
			if !th.onTailCall() {
				return false
			}
		} else {
			if !th.onCall() {
				return false
			}
		}
	}

	rets, err := fn(th, args...)
	if err != object.NoErr {
		th.error(err)
	}

	ctx.stackEnsure(len(rets))

	ctx.ci = ctx.ci.prev

	if th.status == object.THREAD_ERROR {
		return false
	}

	if nrets != -1 && nrets < len(rets) {
		rets = rets[:nrets]
	}

	copy(ctx.stack[f:], rets)

	// clear unused stack
	for r := sp; r >= f+len(rets); r-- {
		ctx.stack[r] = nil
	}

	// adjust sp
	ctx.ci.sp = f + len(rets)

	if ctx.hookMask != 0 {
		return th.onReturn()
	}

	return true
}

// call a go function by values, immediately return values.
func (th *thread) callvGo(fn object.GoFunction, args ...object.Value) (rets []object.Value, ok bool) {
	ctx := th.context

	ci := &callInfo{
		nrets:      -1,
		prev:       ctx.ci,
		isTailCall: false,

		// fake infos
		base: 2,
		sp:   -1,
		pc:   -1,
	}

	ctx.ci = ci

	if ctx.hookMask != 0 {
		if !th.onCall() {
			return nil, false
		}
	}

	// see getInfo

	old := ctx.stack[1]

	ctx.stack[1] = fn

	rets, err := fn(th, args...)
	if err != object.NoErr {
		th.error(err)
	}

	ctx.stack[1] = old

	ctx.ci = ctx.ci.prev

	if th.status == object.THREAD_ERROR {
		return nil, false
	}

	if ctx.hookMask != 0 {
		if !th.onReturn() {
			return nil, false
		}
	}

	return rets, true
}

// call a go function by values with error handler, immediately return values.
func (th *thread) pcallvGo(fn object.GoFunction, errh object.Value, args ...object.Value) (rets []object.Value, ok bool) {
	rets, ok = th.callvGo(fn, args...)

	ctx := th.context

	if ctx.status == object.THREAD_ERROR {
		err := ctx.data.(*Error)

		ctx.status = object.THREAD_RUNNING
		ctx.data = nil

		val := err.RetValue()

		if errh == nil {
			return []object.Value{val}, false
		}

		return th.dohandle(errh, val), false
	}

	return
}

// call a callable by stack index.
func (th *thread) call(a, nargs, nrets int) (ok bool) {
	ctx := th.context

	f := ctx.ci.base + a

	fn := ctx.stack[f]

	switch fn := fn.(type) {
	case nil:
		th.throwCallError(fn)

		return false
	case object.GoFunction:
		return th.callGo(fn, f, nargs, nrets, false)
	case object.Closure:
		return th.callLua(fn, f, nargs, nrets)
	}

	tm := th.gettmbyobj(fn, TM_CALL)

	if !isfunction(tm) {
		th.throwCallError(fn)

		return false
	}

	ctx.stackEnsure(1)

	copy(ctx.stack[f+1:], ctx.stack[f:f+1+nargs])

	ctx.stack[f] = tm

	return th.call(a, nargs+1, nrets)
}

// call a callable by values, immediately return values.
func (th *thread) docallv(fn object.Value, args ...object.Value) (rets []object.Value, ok bool) {
	switch fn := fn.(type) {
	case nil:
		th.throwCallError(fn)

		return nil, false
	case object.GoFunction:
		return th.callvGo(fn, args...)
	case object.Closure:
		return th.docallvLua(fn, args...)
	}

	tm := th.gettmbyobj(fn, TM_CALL)

	return th.docallv(tm, append([]object.Value{fn}, args...)...)
}

// call a callable by values with error handler, immediately return values.
func (th *thread) dopcallv(fn object.Value, errh object.Value, args ...object.Value) (rets []object.Value, ok bool) {
	switch fn := fn.(type) {
	case nil:
		return th.dohandle(errh, object.String("attempt to call a nil value")), false
	case object.GoFunction:
		return th.pcallvGo(fn, errh, args...)
	case object.Closure:
		return th.dopcallvLua(fn, errh, args...)
	}

	tm := th.gettmbyobj(fn, TM_CALL)

	return th.dopcallv(tm, errh, append([]object.Value{fn}, args...)...)
}

// call a error handler.
func (th *thread) dohandle(errh object.Value, arg object.Value) (rets []object.Value) {
	switch errh := errh.(type) {
	case nil:
		return []object.Value{object.String("error in error handling")}
	case object.GoFunction:
		rets, ok := th.callvGo(errh, arg)
		if !ok {
			th.status = object.THREAD_RUNNING
			th.data = nil

			return []object.Value{object.String("error in error handling")}
		}

		return rets
	case object.Closure:
		rets, ok := th.dopcallvLua(errh, nil, arg)
		if !ok {
			return []object.Value{object.String("error in error handling")}
		}

		return rets
	default:
		panic("unexpected")
	}
}

// tail call a callable by stack index.
func (th *thread) tailcall(a, nargs int) (ok bool) {
	ctx := th.context

	th.closeUpvals(ctx.ci.base) // closing upvalues

	f := ctx.ci.base + a

	fn := ctx.stack[f]

	switch fn := ctx.stack[f].(type) {
	case nil:
		th.throwCallError(fn)

		return false
	case object.GoFunction:
		return th.callGo(fn, f, nargs, -1, true)
	case object.Closure:
		return th.tailcallLua(fn, f, nargs)
	}

	tm := th.gettmbyobj(fn, TM_CALL)

	if !isfunction(tm) {
		th.throwCallError(fn)

		return false
	}

	ctx.stackEnsure(1)

	copy(ctx.stack[f+1:], ctx.stack[f:f+1+nargs])

	ctx.ci.sp++

	ctx.stack[f] = tm

	return th.tailcall(a, nargs+1)
}

// tforcall a callable by stack index.
func (th *thread) tforcall(a, nrets int) (ok bool) {
	ctx := th.context

	f := ctx.ci.base + a

	fn := ctx.stack[f]

	switch fn := fn.(type) {
	case nil:
		th.throwCallError(fn)

		return false
	case object.GoFunction:
		copy(ctx.stack[f+3:], ctx.stack[f:f+3])

		return th.callGo(fn, f+3, 2, nrets, false)
	case object.Closure:
		args := ctx.stack[f+1 : f+3]

		rets, ok := th.docallvLua(fn, args...)

		if !ok {
			return false
		}

		if nrets != -1 && nrets < len(rets) {
			rets = rets[:nrets]
		}

		if len(rets) == 0 {
			ctx.stack[f+3] = nil
		} else {
			copy(ctx.stack[f+3:], rets)
		}

		return true
	}

	tm := th.gettmbyobj(fn, TM_CALL)

	if !isfunction(tm) {
		th.throwCallError(fn)

		return false
	}

	ctx.stackEnsure(1)

	copy(ctx.stack[f+1:], ctx.stack[f:f+3])

	ctx.ci.sp++

	ctx.stack[f] = tm

	return th.tforcall(a, nrets)
}

// post process XXXcall.
func (th *thread) returnLua(a, nrets int) (rets []object.Value, exit bool) {
	ctx := th.context

	ci := ctx.ci

	rets = ctx.stack[ci.base+a : ci.base+a+nrets]

	if ci.nrets != -1 && ci.nrets < len(rets) {
		rets = rets[:ci.nrets]
	}

	th.closeUpvals(ci.base) // closing upvalues

	prev := ci.prev

	if prev.isBase() {
		if ctx.hookMask != 0 {
			if !th.onReturn() {
				return nil, true
			}
		}

		ctx.status = object.THREAD_RETURN

		return rets, true
	}

	retbase := ci.base - 1

	maxsp := ci.base + ci.MaxStackSize

	// copy result to stack
	copy(ctx.stack[retbase:], rets)

	// clear unused stack
	for r := maxsp - 1; r >= retbase+len(rets); r-- {
		ctx.stack[r] = nil
	}

	ctx.ci = prev

	// adjust sp
	ctx.ci.sp = retbase + len(rets)

	if ctx.hookMask != 0 {
		if !th.onReturn() {
			return nil, true
		}
	}

	return nil, false
}
