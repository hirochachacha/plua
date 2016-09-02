/******************************************************************************
* Original: src/lvm.c

* Copyright (C) 1994-2016 Lua.org, PUC-Rio.
* Portions Copyright 2016 Hiroshi Ioka. All rights reserved.
*
* Permission is hereby granted, free of charge, to any person obtaining
* a copy of this software and associated documentation files (the
* "Software"), to deal in the Software without restriction, including
* without limitation the rights to use, copy, modify, merge, publish,
* distribute, sublicense, and/or sell copies of the Software, and to
* permit persons to whom the Software is furnished to do so, subject to
* the following conditions:
*
* The above copyright notice and this permission notice shall be
* included in all copies or substantial portions of the Software.
*
* THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND,
* EXPRESS OR IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF
* MERCHANTABILITY, FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT.
* IN NO EVENT SHALL THE AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY
* CLAIM, DAMAGES OR OTHER LIABILITY, WHETHER IN AN ACTION OF CONTRACT,
* TORT OR OTHERWISE, ARISING FROM, OUT OF OR IN CONNECTION WITH THE
* SOFTWARE OR THE USE OR OTHER DEALINGS IN THE SOFTWARE.
******************************************************************************/

package runtime

import (
	// "fmt"

	"github.com/hirochachacha/plua/internal/arith"
	"github.com/hirochachacha/plua/internal/version"
	"github.com/hirochachacha/plua/object"
	"github.com/hirochachacha/plua/opcode"
)

func (th *thread) initExecute(args []object.Value) (rets []object.Value, exit bool) {
	ctx := th.context

	switch fn := ctx.stack[1].(type) {
	case nil:
		panic("main function isn't loaded yet")
	case object.GoFunction:
		rets, _ = th.callvGo(fn, args...)
		exit = true
	case object.Closure:
		copy(ctx.stack[2:], args)

		th.callLua(fn, 1, len(args), -1)
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

	for r := ci.sp; r >= ci.base+len(args); r-- {
		ctx.stack[r] = nil
	}

	// adjust sp
	ctx.ci.sp = ci.base + len(args)
}

func (th *thread) execute() {
	defer close(th.resume)
	defer close(th.yield)

	args := <-th.resume

	if rets, exit := th.initExecute(args); exit {
		if th.status == object.THREAD_RUNNING {
			th.yield <- rets
		}

		if th.status != object.THREAD_ERROR {
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

					th.propagate(err)

					return
				}
			}

			ctx.closeUpvals(0) // close all upvalues on this context

			val := err.Positioned()

			if ctx.errh == protect {
				rets = []object.Value{val}
			} else {
				rets = th.dohandle(ctx.errh, val)
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

func (th *thread) doExecute(fn, errh object.Value, args []object.Value) (rets []object.Value, ok bool) {
	th.pushContext(basicStackSize)

	th.errh = errh

	th.loadfn(fn)

	if rets, exit := th.initExecute(args); exit {
		if th.status == object.THREAD_RUNNING {
			return rets, true
		}

		if th.status != object.THREAD_ERROR {
			panic("unexpected")
		}

		return nil, false
	}

	rets = th.execute0()

	ctx := th.popContext()

	switch ctx.status {
	case object.THREAD_RETURN:
		ctx.closeUpvals(0) // close all upvalues on this context

		return rets, true
	case object.THREAD_ERROR:
		err := ctx.data.(*object.RuntimeError)

		ctx.closeUpvals(0) // close all upvalues on this context

		if ctx.errh != nil {
			val := err.Positioned()

			if ctx.errh == protect {
				rets = []object.Value{val}
			} else {
				rets = th.dohandle(ctx.errh, val)
			}

			return rets, false
		}

		th.propagate(err)

		return nil, false
	default:
		panic("unreachable")
	}
}

// execute with current context
func (th *thread) execute0() (rets []object.Value) {
	// fmt.Println("execute0")

	// defer func() { fmt.Println("exit execute0", rets) }()

	if th.depth >= version.MAX_VM_RECURSION {
		th.throwStackOverflowError()

		return nil
	}

	ctx := th.context

	ctx.status = object.THREAD_RUNNING

	var inst opcode.Instruction

	ci := ctx.ci

	for {
		inst = ci.Code[ci.pc]

		if ctx.hookMask != 0 {
			if !th.onInstruction() {
				return
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
				th.throwByteCodeError()

				return
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

			val, tm, ok := th.gettable(t, key)
			if !ok {
				return
			}
			if tm != nil {
				if !th.calltm(inst.A(), tm, t, key) {
					return
				}
			} else {
				ctx.setRA(inst, val)
			}
		case opcode.GETTABLE:
			t := ctx.getRB(inst)
			key := ctx.getRKC(inst)

			val, tm, ok := th.gettable(t, key)
			if !ok {
				return
			}
			if tm != nil {
				if !th.calltm(inst.A(), tm, t, key) {
					return
				}
			} else {
				ctx.setRA(inst, val)
			}
		case opcode.SETTABUP:
			t := ctx.getUA(inst)
			key := ctx.getRKB(inst)
			val := ctx.getRKC(inst)

			tm, ok := th.settable(t, key, val)
			if !ok {
				return
			}

			if tm != nil {
				if !th.calltm(inst.A(), tm, t, key, val) {
					return
				}
			}
		case opcode.SETUPVAL:
			ctx.setUB(inst, ctx.getRA(inst))
		case opcode.SETTABLE:
			t := ctx.getRA(inst)
			key := ctx.getRKB(inst)
			val := ctx.getRKC(inst)

			tm, ok := th.settable(t, key, val)
			if !ok {
				return
			}

			if tm != nil {
				if !th.calltm(inst.A(), tm, t, key, val) {
					return
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

			val, tm, ok := th.gettable(t, key)
			if !ok {
				return
			}
			if tm != nil {
				if !th.calltm(a, tm, t, key) {
					return
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
				if !th.callbintm(inst.A(), rb, rc, TM_ADD) {
					return
				}
			}
		case opcode.SUB:
			rb := ctx.getRKB(inst)
			rc := ctx.getRKC(inst)

			if diff := arith.Sub(rb, rc); diff != nil {
				ctx.setRA(inst, diff)
			} else {
				if !th.callbintm(inst.A(), rb, rc, TM_SUB) {
					return
				}
			}
		case opcode.MUL:
			rb := ctx.getRKB(inst)
			rc := ctx.getRKC(inst)

			if prod := arith.Mul(rb, rc); prod != nil {
				ctx.setRA(inst, prod)
			} else {
				if !th.callbintm(inst.A(), rb, rc, TM_MUL) {
					return
				}
			}
		case opcode.DIV:
			rb := ctx.getRKB(inst)
			rc := ctx.getRKC(inst)

			if quo := arith.Div(rb, rc); quo != nil {
				ctx.setRA(inst, quo)
			} else {
				if !th.callbintm(inst.A(), rb, rc, TM_DIV) {
					return
				}
			}
		case opcode.IDIV:
			rb := ctx.getRKB(inst)
			rc := ctx.getRKC(inst)

			quo, ok := arith.Idiv(rb, rc)
			if !ok {
				th.throwZeroDivisionError()

				return
			}

			if quo != nil {
				ctx.setRA(inst, quo)
			} else {
				if !th.callbintm(inst.A(), rb, rc, TM_IDIV) {
					return
				}
			}
		case opcode.BAND:
			rb := ctx.getRKB(inst)
			rc := ctx.getRKC(inst)

			if band := arith.Band(rb, rc); band != nil {
				ctx.setRA(inst, band)
			} else {
				if !th.callbintm(inst.A(), rb, rc, TM_BAND) {
					return
				}
			}
		case opcode.BOR:
			rb := ctx.getRKB(inst)
			rc := ctx.getRKC(inst)

			if bor := arith.Bor(rb, rc); bor != nil {
				ctx.setRA(inst, bor)
			} else {
				if !th.callbintm(inst.A(), rb, rc, TM_BOR) {
					return
				}
			}
		case opcode.BXOR:
			rb := ctx.getRKB(inst)
			rc := ctx.getRKC(inst)

			if bxor := arith.Bxor(rb, rc); bxor != nil {
				ctx.setRA(inst, bxor)
			} else {
				if !th.callbintm(inst.A(), rb, rc, TM_BXOR) {
					return
				}
			}
		case opcode.SHL:
			rb := ctx.getRKB(inst)
			rc := ctx.getRKC(inst)

			if shl := arith.Shl(rb, rc); shl != nil {
				ctx.setRA(inst, shl)
			} else {
				if !th.callbintm(inst.A(), rb, rc, TM_SHL) {
					return
				}
			}
		case opcode.SHR:
			rb := ctx.getRKB(inst)
			rc := ctx.getRKC(inst)

			if shr := arith.Shr(rb, rc); shr != nil {
				ctx.setRA(inst, shr)
			} else {
				if !th.callbintm(inst.A(), rb, rc, TM_SHR) {
					return
				}
			}
		case opcode.MOD:
			rb := ctx.getRKB(inst)
			rc := ctx.getRKC(inst)

			rem, ok := arith.Mod(rb, rc)
			if !ok {
				th.throwModuloByZeroError()

				return
			}

			if rem != nil {
				ctx.setRA(inst, rem)
			} else {
				if !th.callbintm(inst.A(), rb, rc, TM_MOD) {
					return
				}
			}
		case opcode.POW:
			rb := ctx.getRKB(inst)
			rc := ctx.getRKC(inst)

			if prod := arith.Pow(rb, rc); prod != nil {
				ctx.setRA(inst, prod)
			} else {
				if !th.callbintm(inst.A(), rb, rc, TM_POW) {
					return
				}
			}
		case opcode.UNM:
			rb := ctx.getRB(inst)

			if unm := arith.Unm(rb); unm != nil {
				ctx.setRA(inst, unm)
			} else {
				if !th.calluntm(inst.A(), rb, TM_UNM) {
					return
				}
			}
		case opcode.BNOT:
			rb := ctx.getRB(inst)

			if bnot := arith.Bnot(rb); bnot != nil {
				ctx.setRA(inst, bnot)
			} else {
				if !th.calluntm(inst.A(), rb, TM_BNOT) {
					return
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
					if !th.calltm(inst.A(), tm, rb) {
						return
					}
				} else {
					ctx.setRA(inst, object.Integer(rb.ALen()))
				}
			case object.String:
				ctx.setRA(inst, object.Integer(len(rb)))
			default:
				if !th.calluntm(inst.A(), rb, TM_LEN) {
					return
				}
			}
		case opcode.CONCAT:
			if !th.concat(inst.A(), inst.B(), inst.C()) {
				return
			}
		case opcode.JMP:
			th.dojmp(inst)
		case opcode.EQ:
			rb := ctx.getRKB(inst)
			rc := ctx.getRKC(inst)

			not := inst.A() != 0

			b, tm := th.eq(rb, rc)
			if tm != nil {
				if !th.callcmptm(not, tm, rb, rc) {
					return
				}
			} else {
				if b != not {
					ci.pc++
				} else {
					jmp := ci.Code[ci.pc]

					if jmp.OpCode() != opcode.JMP {
						th.throwByteCodeError()

						return
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
						th.throwByteCodeError()

						return
					}

					ci.pc++

					th.dojmp(jmp)
				}
			} else {
				if !th.callordertm(not, rb, rc, TM_LT) {
					return
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
						th.throwByteCodeError()

						return
					}

					ci.pc++

					th.dojmp(jmp)
				}
			} else {
				if !th.callordertm(not, rb, rc, TM_LE) {
					return
				}
			}
		case opcode.TEST:
			ra := ctx.getRA(inst)

			if object.ToGoBool(ra) != (inst.C() != 0) {
				ci.pc++
			} else {
				jmp := ci.Code[ci.pc]

				if jmp.OpCode() != opcode.JMP {
					th.throwByteCodeError()

					return
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
					th.throwByteCodeError()

					return
				}

				ci.pc++

				th.dojmp(jmp)
			}
		case opcode.CALL:
			a := inst.A()

			nargs := inst.B() - 1
			nrets := inst.C() - 1

			if nargs == -1 {
				nargs = ci.sp - ci.base - a - 1
			}

			if !th.call(a, nargs, nrets) {
				return
			}
			ci = ctx.ci
		case opcode.TAILCALL:
			a := inst.A()

			nargs := inst.B() - 1

			if nargs == -1 {
				nargs = ci.sp - ci.base - a - 1
			}

			if !th.tailcall(a, nargs) {
				return
			}
			ci = ctx.ci
		case opcode.RETURN:
			a := inst.A()

			nrets := inst.B() - 1

			if nrets == -1 {
				nrets = ci.sp - ci.base - a
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
					th.throwForError("initial")

					return
				}

				plimit, ok := object.ToNumber(plimit)
				if !ok {
					th.throwForError("limit")

					return
				}

				pstep, ok := object.ToNumber(pstep)
				if !ok {
					th.throwForError("step")

					return
				}

				ctx.setR(a, init-pstep)
				ctx.setR(a+1, plimit)
				ctx.setR(a+2, pstep)
			}

			ci.pc += inst.SBx()
		case opcode.TFORCALL:
			a := inst.A()
			nrets := inst.C()

			if !th.tforcall(a, nrets) {
				return
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
				length = ci.sp - ci.base - a - 1
			}

			c := inst.C()
			if c == 0 {
				extra := ci.Code[ci.pc]
				if extra.OpCode() != opcode.EXTRAARG {
					th.throwByteCodeError()

					return
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
				th.throwByteCodeError()

				return
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

			ctx.stackEnsure(len(varargs))

			copy(ctx.stack[ci.base+a:], varargs)

			ctx.ci.sp = ci.base + a + len(varargs)
		case opcode.EXTRAARG:
			th.throwByteCodeError()

			return
		default:
			th.throwByteCodeError()

			return
		}
	}
}

func (th *thread) gettable(t, key object.Value) (val object.Value, tm object.Value, ok bool) {
	for i := 0; i < version.MAX_TAG_LOOP; i++ {
		if t, ok := t.(object.Table); ok {
			val := t.Get(key)
			mt := t.Metatable()
			if val != nil || th.fasttm(mt, TM_INDEX) == nil {
				return val, nil, true
			}
		}

		tm := th.gettmbyobj(t, TM_INDEX)
		if tm == nil {
			th.throwIndexError(t)

			return nil, nil, false
		}

		if isfunction(tm) {
			return nil, tm, true
		}

		t = tm
	}

	th.throwGetTableError()

	return nil, nil, false
}

func (th *thread) settable(t, key, val object.Value) (tm object.Value, ok bool) {
	for i := 0; i < version.MAX_TAG_LOOP; i++ {
		if t, ok := t.(object.Table); ok {
			old := t.Get(key)
			mt := t.Metatable()
			if old != nil || th.fasttm(mt, TM_NEWINDEX) == nil {
				if key == nil {
					th.throwNilIndexError()

					return nil, false
				}

				if key == object.NaN {
					th.throwNaNIndexError()

					return nil, false
				}

				t.Set(key, val)

				return nil, true
			}
		}

		tm := th.gettmbyobj(t, TM_NEWINDEX)
		if tm == nil {
			th.throwIndexError(t)

			return nil, false
		}

		if isfunction(tm) {
			return tm, true
		}

		t = tm
	}

	th.throwSetTableError()

	return nil, false
}

func (th *thread) dojmp(inst opcode.Instruction) {
	a := inst.A()
	sbx := inst.SBx()
	if a > 0 {
		th.closeUpvals(th.ci.base + a - 1)
	}
	th.ci.pc += sbx
}

func (th *thread) concat(a, b, c int) bool {
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
				th.throwBinaryError(TM_CONCAT, rb, rc)

				return false
			}
		}

		rets, ok := th.docallv(tm, rb, rc)
		if !ok {
			return false
		}

		rb = rets[0]
	}

	ctx.setR(a, rb)

	return true
}

func (th *thread) eq(rb, rc object.Value) (b bool, tm object.Value) {
	switch rb := rb.(type) {
	case object.Table:
		if rc, ok := rc.(object.Table); ok {
			if rb == rc {
				return true, nil
			}

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
			if rb == rc {
				return true, nil
			}

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
	}

	return object.Equal(rb, rc), nil
}

func isfunction(val object.Value) bool {
	return object.ToType(val) == object.TFUNCTION
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
