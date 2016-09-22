package runtime

import (
	"fmt"

	"github.com/hirochachacha/plua/internal/arith"
	"github.com/hirochachacha/plua/internal/version"
	"github.com/hirochachacha/plua/object"
	"github.com/hirochachacha/plua/opcode"
)

func (th *thread) initExecute(args []object.Value) (rets []object.Value, done bool) {
	ctx := th.context

	switch fn := ctx.stack[ctx.ci.base-1].(type) {
	case nil:
		panic("main function isn't loaded yet")
	case object.GoFunction:
		var err *object.RuntimeError
		rets, err = th.docallGo(fn, args...)

		if err != nil {
			th.error(err)
		} else {
			ctx.status = object.THREAD_RETURN
		}

		done = true
	case object.Closure:
		cl := fn.(*closure)

		ci := &ctx.ciStack[0]
		ci.closure = cl
		ci.top = ci.base + cl.MaxStackSize

		if !ctx.growStack(0) {
			panic(errStackOverflow)
		}

		copy(ctx.stack[ci.base:], args)

		if nvarargs := len(args) - cl.NParams; nvarargs > 0 {
			if cl.IsVararg {
				ci.varargs = make([]object.Value, nvarargs)

				copy(ci.varargs, ctx.stack[ci.base+cl.NParams:ci.base+len(args)])
			}

			for r := ci.top - 1; r >= ci.base+cl.NParams; r-- {
				ctx.stack[r] = nil
			}
		} else {
			for r := ci.top - 1; r >= ci.base+len(args); r-- {
				ctx.stack[r] = nil
			}
		}
	default:
		panic("unreachable")
	}

	return
}

func (th *thread) resumeExecute(args []object.Value) {
	ctx := th.context

	ci := ctx.ci

	if ci.nrets != -1 && ci.nrets < len(args) {
		args = args[:ci.nrets]
	}

	ctx.data = nil

	copy(ctx.stack[ci.base-1:], args)

	top := ctx.ci.base - 1 + len(args)

	for r := ci.top; r >= top; r-- {
		ctx.stack[r] = nil
	}

	// adjust top
	ci.top = top
}

func (th *thread) execute() {
	defer close(th.resume)
	defer close(th.yield)

	args := <-th.resume

	if rets, done := th.initExecute(args); done {
		switch th.status {
		case object.THREAD_RETURN:
			th.yield <- rets
		case object.THREAD_ERROR:
		default:
			panic("unexpected")
		}

		return
	}

	for {
		rets := th.execute0()

		ctx := th.context

		switch ctx.status {
		case object.THREAD_RETURN:
			if ctx.isRoot() {
				th.yield <- rets

				return
			}

			ctx.closeUpvals(0) // close all upvalues on this context

			args = rets

			th.popContext()
		case object.THREAD_ERROR:
			if ctx.isRoot() {
				return
			}

			err := ctx.data.(*object.RuntimeError)

			for ctx.errh == nil {
				ctx.closeUpvals(0) // close all upvalues on this context
				ctx = ctx.prev

				if ctx.isRoot() {
					th.context = ctx

					th.error(err)

					return
				}
			}

			ctx.closeUpvals(0) // close all upvalues on this context

			rets, err = th.dohandle(ctx.errh, err)
			if err != nil {
				rets = []object.Value{err.Positioned()}
			}

			args = rets

			th.popContext()
		default:
			panic("unexpected")
		}

		if th.status != object.THREAD_RUNNING {
			panic("unexpected")
		}

		th.resumeExecute(args)
	}
}

func (th *thread) doExecute(fn, errh object.Value, args []object.Value, isHook bool) (rets []object.Value, err *object.RuntimeError) {
	th.pushContext(basicStackSize, isHook)

	th.errh = errh

	th.loadfn(fn)

	if rets, done := th.initExecute(args); done {
		ctx := th.popContext()

		switch ctx.status {
		case object.THREAD_RETURN:
			return rets, nil
		case object.THREAD_ERROR:
			return nil, ctx.data.(*object.RuntimeError)
		default:
			panic("unexpected")
		}
	}

	rets = th.execute0()

	ctx := th.popContext()

	switch ctx.status {
	case object.THREAD_RETURN:
		ctx.closeUpvals(0) // close all upvalues on this context

		return rets, nil
	case object.THREAD_ERROR:
		err := ctx.data.(*object.RuntimeError)

		ctx.closeUpvals(0) // close all upvalues on this context

		if ctx.errh != nil {
			rets, err = th.dohandle(ctx.errh, err)
			if err != nil {
				return nil, err
			}

			return rets, nil
		}

		return nil, err
	default:
		panic("unreachable")
	}
}

// execute with current context
func (th *thread) execute0() (rets []object.Value) {
	if th.depth >= version.MAX_VM_RECURSION {
		th.error(errStackOverflow)

		return nil
	}

	ctx := th.context

	ctx.status = object.THREAD_RUNNING

	var inst opcode.Instruction

	ci := ctx.ci

	for {
		inst = ci.Code[ci.pc]

		if th.hookMask != 0 {
			if err := th.onInstruction(); err != nil {
				th.error(err)

				return nil
			}
		}

		ci.pc++

		switch inst.OpCode() {
		case opcode.MOVE:
			ctx.setRA(inst, ctx.getRB(inst))
		case opcode.LOADK:
			ctx.setRA(inst, ctx.getKBx(inst))
		case opcode.LOADKX:
			extra := ci.Code[ci.pc]
			if extra.OpCode() != opcode.EXTRAARG {
				th.error(errInvalidByteCode)

				return nil
			}

			ctx.setRA(inst, ctx.getKAx(extra))

			ci.pc++
		case opcode.LOADBOOL:
			ctx.setRA(inst, object.Boolean(inst.B() != 0))
			if inst.C() != 0 {
				ci.pc++
			}
		case opcode.LOADNIL:
			a := inst.A()
			for i := 0; i <= inst.B(); i++ {
				ctx.setR(a+i, nil)
			}
		case opcode.GETUPVAL:
			ctx.setRA(inst, ctx.getUB(inst))
		case opcode.GETTABUP:
			t := ctx.getUB(inst)
			key := ctx.getRKC(inst)

			val, tm, err := th.gettable(t, key)
			if err != nil {
				th.error(err)

				return nil
			}
			if tm != nil {
				if err := th.calltm(inst.A(), tm, t, key); err != nil {
					th.error(err)

					return nil
				}
			} else {
				ctx.setRA(inst, val)
			}
		case opcode.GETTABLE:
			t := ctx.getRB(inst)
			key := ctx.getRKC(inst)

			val, tm, err := th.gettable(t, key)
			if err != nil {
				th.error(err)

				return nil
			}
			if tm != nil {
				if err := th.calltm(inst.A(), tm, t, key); err != nil {
					th.error(err)

					return nil
				}
			} else {
				ctx.setRA(inst, val)
			}
		case opcode.SETTABUP:
			t := ctx.getUA(inst)
			key := ctx.getRKB(inst)
			val := ctx.getRKC(inst)

			tm, err := th.settable(t, key, val)
			if err != nil {
				th.error(err)

				return nil
			}

			if tm != nil {
				if err := th.calltm(inst.A(), tm, t, key, val); err != nil {
					th.error(err)

					return nil
				}
			}
		case opcode.SETUPVAL:
			ctx.setUB(inst, ctx.getRA(inst))
		case opcode.SETTABLE:
			t := ctx.getRA(inst)
			key := ctx.getRKB(inst)
			val := ctx.getRKC(inst)

			tm, err := th.settable(t, key, val)
			if err != nil {
				th.error(err)

				return nil
			}

			if tm != nil {
				if err := th.calltm(inst.A(), tm, t, key, val); err != nil {
					th.error(err)

					return nil
				}
			}
		case opcode.NEWTABLE:
			asize := opcode.LogToInt(inst.B())
			msize := opcode.LogToInt(inst.C())

			t := newTableSize(asize, msize)

			ctx.setRA(inst, t)
		case opcode.SELF:
			a := inst.A()

			t := ctx.getRB(inst)
			ctx.setR(a+1, t)

			key := ctx.getRKC(inst)

			val, tm, err := th.gettable(t, key)
			if err != nil {
				th.error(err)

				return nil
			}
			if tm != nil {
				if err := th.calltm(a, tm, t, key); err != nil {
					th.error(err)

					return nil
				}
			} else {
				ctx.setR(a, val)
			}
		case opcode.ADD:
			rb := ctx.getRKB(inst)
			rc := ctx.getRKC(inst)

			if sum := arith.Add(rb, rc); sum != nil {
				ctx.setRA(inst, sum)
			} else {
				if err := th.callbintm(inst.A(), rb, rc, TM_ADD); err != nil {
					th.error(err)

					return nil
				}
			}
		case opcode.SUB:
			rb := ctx.getRKB(inst)
			rc := ctx.getRKC(inst)

			if diff := arith.Sub(rb, rc); diff != nil {
				ctx.setRA(inst, diff)
			} else {
				if err := th.callbintm(inst.A(), rb, rc, TM_SUB); err != nil {
					th.error(err)

					return nil
				}
			}
		case opcode.MUL:
			rb := ctx.getRKB(inst)
			rc := ctx.getRKC(inst)

			if prod := arith.Mul(rb, rc); prod != nil {
				ctx.setRA(inst, prod)
			} else {
				if err := th.callbintm(inst.A(), rb, rc, TM_MUL); err != nil {
					th.error(err)

					return nil
				}
			}
		case opcode.DIV:
			rb := ctx.getRKB(inst)
			rc := ctx.getRKC(inst)

			if quo := arith.Div(rb, rc); quo != nil {
				ctx.setRA(inst, quo)
			} else {
				if err := th.callbintm(inst.A(), rb, rc, TM_DIV); err != nil {
					th.error(err)

					return nil
				}
			}
		case opcode.IDIV:
			rb := ctx.getRKB(inst)
			rc := ctx.getRKC(inst)

			quo, ok := arith.Idiv(rb, rc)
			if !ok {
				th.error(errZeroDivision)

				return nil
			}

			if quo != nil {
				ctx.setRA(inst, quo)
			} else {
				if err := th.callbintm(inst.A(), rb, rc, TM_IDIV); err != nil {
					th.error(err)

					return nil
				}
			}
		case opcode.BAND:
			rb := ctx.getRKB(inst)
			rc := ctx.getRKC(inst)

			if band := arith.Band(rb, rc); band != nil {
				ctx.setRA(inst, band)
			} else {
				if err := th.callbintm(inst.A(), rb, rc, TM_BAND); err != nil {
					th.error(err)

					return nil
				}
			}
		case opcode.BOR:
			rb := ctx.getRKB(inst)
			rc := ctx.getRKC(inst)

			if bor := arith.Bor(rb, rc); bor != nil {
				ctx.setRA(inst, bor)
			} else {
				if err := th.callbintm(inst.A(), rb, rc, TM_BOR); err != nil {
					th.error(err)

					return nil
				}
			}
		case opcode.BXOR:
			rb := ctx.getRKB(inst)
			rc := ctx.getRKC(inst)

			if bxor := arith.Bxor(rb, rc); bxor != nil {
				ctx.setRA(inst, bxor)
			} else {
				if err := th.callbintm(inst.A(), rb, rc, TM_BXOR); err != nil {
					th.error(err)

					return nil
				}
			}
		case opcode.SHL:
			rb := ctx.getRKB(inst)
			rc := ctx.getRKC(inst)

			if shl := arith.Shl(rb, rc); shl != nil {
				ctx.setRA(inst, shl)
			} else {
				if err := th.callbintm(inst.A(), rb, rc, TM_SHL); err != nil {
					th.error(err)

					return nil
				}
			}
		case opcode.SHR:
			rb := ctx.getRKB(inst)
			rc := ctx.getRKC(inst)

			if shr := arith.Shr(rb, rc); shr != nil {
				ctx.setRA(inst, shr)
			} else {
				if err := th.callbintm(inst.A(), rb, rc, TM_SHR); err != nil {
					th.error(err)

					return nil
				}
			}
		case opcode.MOD:
			rb := ctx.getRKB(inst)
			rc := ctx.getRKC(inst)

			rem, ok := arith.Mod(rb, rc)
			if !ok {
				th.error(errModuloByZero)

				return nil
			}

			if rem != nil {
				ctx.setRA(inst, rem)
			} else {
				if err := th.callbintm(inst.A(), rb, rc, TM_MOD); err != nil {
					th.error(err)

					return nil
				}
			}
		case opcode.POW:
			rb := ctx.getRKB(inst)
			rc := ctx.getRKC(inst)

			if prod := arith.Pow(rb, rc); prod != nil {
				ctx.setRA(inst, prod)
			} else {
				if err := th.callbintm(inst.A(), rb, rc, TM_POW); err != nil {
					th.error(err)

					return nil
				}
			}
		case opcode.UNM:
			rb := ctx.getRB(inst)

			if unm := arith.Unm(rb); unm != nil {
				ctx.setRA(inst, unm)
			} else {
				if err := th.calluntm(inst.A(), rb, TM_UNM); err != nil {
					th.error(err)

					return nil
				}
			}
		case opcode.BNOT:
			rb := ctx.getRB(inst)

			if bnot := arith.Bnot(rb); bnot != nil {
				ctx.setRA(inst, bnot)
			} else {
				if err := th.calluntm(inst.A(), rb, TM_BNOT); err != nil {
					th.error(err)

					return nil
				}
			}
		case opcode.NOT:
			rb := ctx.getRB(inst)

			ctx.setRA(inst, arith.Not(rb))
		case opcode.LEN:
			rb := ctx.getRB(inst)

			var tm object.Value

			switch rb := rb.(type) {
			case object.Table:
				mt := rb.Metatable()
				tm = th.fasttm(mt, TM_LEN)
				if tm != nil {
					if err := th.calltm(inst.A(), tm, rb); err != nil {
						th.error(err)

						return nil
					}
				} else {
					ctx.setRA(inst, object.Integer(rb.Len()))
				}
			case object.String:
				ctx.setRA(inst, object.Integer(len(rb)))
			default:
				if err := th.calluntm(inst.A(), rb, TM_LEN); err != nil {
					th.error(err)

					return nil
				}
			}
		case opcode.CONCAT:
			if err := th.concat(inst.A(), inst.B(), inst.C()); err != nil {
				th.error(err)

				return nil
			}
		case opcode.JMP:
			th.dojmp(inst)
		case opcode.EQ:
			rb := ctx.getRKB(inst)
			rc := ctx.getRKC(inst)

			not := inst.A() != 0

			b, tm := th.eq(rb, rc)
			if tm != nil {
				if err := th.callcmptm(not, tm, rb, rc); err != nil {
					th.error(err)

					return nil
				}
			} else {
				if b != not {
					ci.pc++
				} else {
					jmp := ci.Code[ci.pc]

					if jmp.OpCode() != opcode.JMP {
						th.error(errInvalidByteCode)

						return nil
					}

					ci.pc++

					th.dojmp(jmp)
				}
			}
		case opcode.LT:
			rb := ctx.getRKB(inst)
			rc := ctx.getRKC(inst)

			not := inst.A() != 0

			if b := arith.LessThan(rb, rc); b != nil {
				if bool(b.(object.Boolean)) != not {
					ci.pc++
				} else {
					jmp := ci.Code[ci.pc]

					if jmp.OpCode() != opcode.JMP {
						th.error(errInvalidByteCode)

						return nil
					}

					ci.pc++

					th.dojmp(jmp)
				}
			} else {
				if err := th.callordertm(not, rb, rc, TM_LT); err != nil {
					th.error(err)

					return nil
				}
			}
		case opcode.LE:
			rb := ctx.getRKB(inst)
			rc := ctx.getRKC(inst)

			not := inst.A() != 0

			if b := arith.LessThanOrEqualTo(rb, rc); b != nil {
				if bool(b.(object.Boolean)) != not {
					ci.pc++
				} else {
					jmp := ci.Code[ci.pc]

					if jmp.OpCode() != opcode.JMP {
						th.error(errInvalidByteCode)

						return nil
					}

					ci.pc++

					th.dojmp(jmp)
				}
			} else {
				if err := th.callordertm(not, rb, rc, TM_LE); err != nil {
					th.error(err)

					return nil
				}
			}
		case opcode.TEST:
			ra := ctx.getRA(inst)

			if object.ToGoBool(ra) != (inst.C() != 0) {
				ci.pc++
			} else {
				jmp := ci.Code[ci.pc]

				if jmp.OpCode() != opcode.JMP {
					th.error(errInvalidByteCode)

					return nil
				}

				ci.pc++

				th.dojmp(jmp)
			}
		case opcode.TESTSET:
			rb := ctx.getRB(inst)

			if object.ToGoBool(rb) != (inst.C() != 0) {
				ci.pc++
			} else {
				ctx.setRA(inst, rb)

				jmp := ci.Code[ci.pc]

				if jmp.OpCode() != opcode.JMP {
					th.error(errInvalidByteCode)

					return nil
				}

				ci.pc++

				th.dojmp(jmp)
			}
		case opcode.CALL:
			a := inst.A()

			nargs := inst.B() - 1
			nrets := inst.C() - 1

			if nargs == -1 {
				nargs = ci.top - ci.base - a - 1
			}

			if err := th.call(a, nargs, nrets); err != nil {
				th.error(err)

				return nil
			}
			ci = ctx.ci
		case opcode.TAILCALL:
			a := inst.A()

			nargs := inst.B() - 1

			if nargs == -1 {
				nargs = ci.top - ci.base - a - 1
			}

			if err := th.tailcall(a, nargs); err != nil {
				th.error(err)

				return nil
			}
			ci = ctx.ci
		case opcode.RETURN:
			a := inst.A()

			nrets := inst.B() - 1

			if nrets == -1 {
				nrets = ci.top - ci.base - a
			}

			if rets, exit := th.returnLua(a, nrets); exit {
				return rets
			}

			ci = ctx.ci
		case opcode.FORLOOP:
			a := inst.A()
			ra := ctx.getR(a)
			step := ctx.getR(a + 2)
			limit := ctx.getR(a + 1)

			// forprep already convert val to integer or number.
			// so there are no need to check types.
			if _, ok := ra.(object.Integer); ok {
				ra := ra.(object.Integer)
				step := step.(object.Integer)
				limit := limit.(object.Integer)
				idx := ra + step
				if 0 < step {
					if idx <= limit {
						ci.pc += inst.SBx()
						ctx.setR(a, idx)
						ctx.setR(a+3, idx)
					}
				} else {
					if idx >= limit {
						ci.pc += inst.SBx()
						ctx.setR(a, idx)
						ctx.setR(a+3, idx)
					}
				}
			} else {
				ra := ra.(object.Number)
				step := step.(object.Number)
				limit := limit.(object.Number)
				idx := ra + step
				if 0 < step {
					if idx <= limit {
						ci.pc += inst.SBx()
						ctx.setR(a, idx)
						ctx.setR(a+3, idx)
					}
				} else {
					if idx >= limit {
						ci.pc += inst.SBx()
						ctx.setR(a, idx)
						ctx.setR(a+3, idx)
					}
				}
			}
		case opcode.FORPREP:
			a := inst.A()
			init := ctx.getR(a)
			plimit := ctx.getR(a + 1)
			pstep := ctx.getR(a + 2)

			if init, ok := init.(object.Integer); ok {
				if pstep, ok := pstep.(object.Integer); ok {
					if _, ok := plimit.(object.Integer); ok {
						ctx.setR(a, init-pstep)

						ci.pc += inst.SBx()

						continue

					}
				}
			}

			{
				init, ok := object.ToNumber(init)
				if !ok {
					th.error(th.forLoopError("initial"))

					return nil
				}

				plimit, ok := object.ToNumber(plimit)
				if !ok {
					th.error(th.forLoopError("limit"))

					return nil
				}

				pstep, ok := object.ToNumber(pstep)
				if !ok {
					th.error(th.forLoopError("step"))

					return nil
				}

				ctx.setR(a, init-pstep)
				ctx.setR(a+1, plimit)
				ctx.setR(a+2, pstep)
			}

			ci.pc += inst.SBx()
		case opcode.TFORCALL:
			a := inst.A()
			nrets := inst.C()

			if err := th.tforcall(a, nrets); err != nil {
				th.error(err)

				return nil
			}

			tloop := ci.Code[ci.pc]

			if tloop.OpCode() != opcode.TFORLOOP {
				th.error(errInvalidByteCode)

				return nil
			}
		case opcode.TFORLOOP:
			a := inst.A()
			raplus := ctx.getR(a + 1)

			if raplus != nil {
				ctx.setR(a, raplus)

				ci.pc += inst.SBx()
			}
		case opcode.SETLIST:
			a := inst.A()
			length := inst.B()
			if length == 0 {
				length = ci.top - ci.base - a - 1
			}

			c := inst.C()
			if c == 0 {
				extra := ci.Code[ci.pc]
				if extra.OpCode() != opcode.EXTRAARG {
					th.error(errInvalidByteCode)

					return nil
				}

				ci.pc++

				c = extra.Ax()
			}

			base := (c - 1) * version.LUA_FPF

			t := ctx.getR(a).(object.Table)

			t.SetList(base, ctx.stack[ci.base+a+1:ci.base+a+1+length])
		case opcode.CLOSURE:
			bx := inst.Bx()

			if len(ci.Protos) <= bx {
				th.error(errInvalidByteCode)

				return nil
			}

			cl := th.makeClosure(bx)

			ctx.setRA(inst, cl)
		case opcode.VARARG:
			a := inst.A()
			nrets := inst.B() - 1

			varargs := ci.varargs
			if nrets != -1 {
				varargs = varargs[:nrets]
			}

			if !ctx.growStack(len(varargs)) {
				th.error(errStackOverflow)
			}

			copy(ctx.stack[ci.base+a:], varargs)

			ctx.ci.top = ci.base + a + len(varargs)
		case opcode.EXTRAARG:
			th.error(errInvalidByteCode)

			return nil
		default:
			th.error(errInvalidByteCode)

			return nil
		}
	}
}

func (th *thread) gettable(t, key object.Value) (val object.Value, tm object.Value, err *object.RuntimeError) {
	for i := 0; i < version.MAX_TAG_LOOP; i++ {
		if t, ok := t.(object.Table); ok {
			val := t.Get(key)
			mt := t.Metatable()
			if val != nil || th.fasttm(mt, TM_INDEX) == nil {
				return val, nil, nil
			}
		}

		tm := th.gettmbyobj(t, TM_INDEX)
		if tm == nil {
			return nil, nil, th.indexError(t)
		}

		if isFunction(tm) {
			return nil, tm, nil
		}

		t = tm
	}

	return nil, nil, errGetTable
}

func (th *thread) settable(t, key, val object.Value) (tm object.Value, err *object.RuntimeError) {
	for i := 0; i < version.MAX_TAG_LOOP; i++ {
		if t, ok := t.(object.Table); ok {
			old := t.Get(key)
			mt := t.Metatable()
			if old != nil || th.fasttm(mt, TM_NEWINDEX) == nil {
				if key == nil {
					return nil, errNilIndex
				}

				if key == object.NaN {
					return nil, errNaNIndex
				}

				t.Set(key, val)

				return nil, nil
			}
		}

		tm := th.gettmbyobj(t, TM_NEWINDEX)
		if tm == nil {
			return nil, th.indexError(t)
		}

		if isFunction(tm) {
			return tm, nil
		}

		t = tm
	}

	return nil, errSetTable
}

func (th *thread) dojmp(inst opcode.Instruction) {
	a := inst.A()
	sbx := inst.SBx()
	if a > 0 {
		th.closeUpvals(th.ci.base + a - 1)
	}
	th.ci.pc += sbx
}

func (th *thread) concat(a, b, c int) *object.RuntimeError {
	ctx := th.context
	ci := ctx.ci

	rb := ctx.stack[ci.base+b]
	for r := b + 1; r <= c; r++ {
		rc := ctx.stack[ci.base+r]

		if con := arith.Concat(rb, rc); con != nil {
			rb = con

			continue
		}

		tm := th.gettmbyobj(rb, TM_CONCAT)
		if tm == nil {
			tm = th.gettmbyobj(rc, TM_CONCAT)

			if tm == nil {
				return th.binaryError(TM_CONCAT, rb, rc)
			}
		}

		rets, err := th.docall(tm, nil, rb, rc)
		if err != nil {
			return err
		}

		rb = rets[0]
	}

	ctx.setR(a, rb)

	return nil
}

func (th *thread) eq(rb, rc object.Value) (b bool, tm object.Value) {
	// fast path for avoiding assertI2I2
	eq := object.Equal(rb, rc)
	if eq {
		return true, nil
	}

	switch rb := rb.(type) {
	case object.Table:
		if rc, ok := rc.(object.Table); ok {
			tm := th.fasttm(rb.Metatable(), TM_EQ)
			if tm == nil {
				tm = th.fasttm(rc.Metatable(), TM_EQ)
				if tm == nil {
					return false, nil
				}
			}

			return false, tm
		}

		return false, nil
	case *object.Userdata:
		if rc, ok := rc.(*object.Userdata); ok {
			tm := th.fasttm(rb.Metatable, TM_EQ)
			if tm == nil {
				tm = th.fasttm(rc.Metatable, TM_EQ)
				if tm == nil {
					return false, nil
				}
			}

			return false, tm
		}

		return false, nil
	default:
		return eq, nil
	}
}

func isFunction(val object.Value) bool {
	return object.ToType(val) == object.TFUNCTION
}

func mustFunction(val object.Value) {
	if !isFunction(val) {
		panic(fmt.Sprintf("%v is not a function", val))
	}
}

func mustFunctionOrNil(val object.Value) {
	t := object.ToType(val)
	if t != object.TNIL && t != object.TFUNCTION {
		panic(fmt.Sprintf("%v is not a function", val))
	}
}

func (ctx *context) getR(r int) object.Value {
	return ctx.stack[ctx.ci.base+r]
}

func (ctx *context) setR(r int, val object.Value) {
	ctx.stack[ctx.ci.base+r] = val
}

func (ctx *context) getK(k int) object.Value {
	return ctx.ci.Constants[k]
}

func (ctx *context) getRK(rk int) object.Value {
	if rk&opcode.BitRK != 0 {
		return ctx.getK(rk & ^opcode.BitRK)
	}

	return ctx.getR(rk)
}

func (ctx *context) getU(r int) object.Value {
	return ctx.ci.GetUpvalue(r)
}

func (ctx *context) setU(r int, val object.Value) {
	ctx.ci.SetUpvalue(r, val)
}

func (ctx *context) getRA(inst opcode.Instruction) object.Value {
	return ctx.getR(inst.A())
}

func (ctx *context) getRB(inst opcode.Instruction) object.Value {
	return ctx.getR(inst.B())
}

func (ctx *context) getRC(inst opcode.Instruction) object.Value {
	return ctx.getR(inst.C())
}

func (ctx *context) setRA(inst opcode.Instruction, val object.Value) {
	ctx.setR(inst.A(), val)
}

func (ctx *context) setUB(inst opcode.Instruction, val object.Value) {
	ctx.setU(inst.B(), val)
}

func (ctx *context) getKBx(inst opcode.Instruction) object.Value {
	return ctx.getK(inst.Bx())
}

func (ctx *context) getKAx(inst opcode.Instruction) object.Value {
	return ctx.getK(inst.Ax())
}

func (ctx *context) getRKB(inst opcode.Instruction) object.Value {
	return ctx.getRK(inst.B())
}

func (ctx *context) getRKC(inst opcode.Instruction) object.Value {
	return ctx.getRK(inst.C())
}

func (ctx *context) getUA(inst opcode.Instruction) object.Value {
	return ctx.getU(inst.A())
}

func (ctx *context) getUB(inst opcode.Instruction) object.Value {
	return ctx.getU(inst.B())
}
