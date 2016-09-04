package runtime

import (
	// "fmt"
	"github.com/hirochachacha/plua/object"
	"github.com/hirochachacha/plua/opcode"
)

const MaxTagType = TM_CALL + 1

type tagType int

func (t tagType) String() string {
	return tagNames[t]
}

const (
	TM_INDEX tagType = iota
	TM_NEWINDEX
	TM_GC
	TM_MODE
	TM_LEN
	TM_EQ
	TM_ADD
	TM_SUB
	TM_MUL
	TM_MOD
	TM_POW
	TM_DIV
	TM_IDIV
	TM_BAND
	TM_BOR
	TM_BXOR
	TM_SHL
	TM_SHR
	TM_UNM
	TM_BNOT
	TM_LT
	TM_LE
	TM_CONCAT
	TM_CALL
)

var tagNames = [...]string{
	TM_INDEX:    "__index",
	TM_NEWINDEX: "__newindex",
	TM_GC:       "__gc",
	TM_MODE:     "__mode",
	TM_LEN:      "__len",
	TM_EQ:       "__eq",
	TM_ADD:      "__add",
	TM_SUB:      "__sub",
	TM_MUL:      "__mul",
	TM_MOD:      "__mod",
	TM_POW:      "__pow",
	TM_DIV:      "__div",
	TM_IDIV:     "__idiv",
	TM_BAND:     "__band",
	TM_BOR:      "__bor",
	TM_BXOR:     "__bxor",
	TM_SHL:      "__shl",
	TM_SHR:      "__shr",
	TM_UNM:      "__unm",
	TM_BNOT:     "__bnot",
	TM_LT:       "__lt",
	TM_LE:       "__le",
	TM_CONCAT:   "__concat",
	TM_CALL:     "__call",
}

func (th *thread) fasttm(mt object.Table, tag tagType) object.Value {
	return th.gettm(mt, tag)
}

func (th *thread) gettm(mt object.Table, tag tagType) object.Value {
	if mt == nil {
		return nil
	}

	return mt.Get(object.String(tag.String()))
}

func (th *thread) gettmbyobj(val object.Value, tag tagType) object.Value {
	mt := th.env.getMetatable(val)
	if mt == nil {
		return nil
	}

	return th.gettm(mt, tag)
}

func (th *thread) calltm(a int, tm object.Value, args ...object.Value) (err *object.RuntimeError) {
	rets, err := th.docallv(tm, args...)
	if err != nil {
		return err
	}

	if len(rets) == 0 {
		th.stack[th.ci.base+a] = nil
	} else {
		th.stack[th.ci.base+a] = rets[0]
	}

	return nil
}

func (th *thread) callcmptm(not bool, tm object.Value, x, y object.Value) (err *object.RuntimeError) {
	rets, err := th.docallv(tm, x, y)
	if err != nil {
		return err
	}

	var ret object.Value

	if len(rets) != 0 {
		ret = rets[0]
	}

	ci := th.ci

	if object.ToGoBool(ret) != not {
		ci.pc++
	} else {
		jmp := ci.Code[ci.pc]

		if jmp.OpCode() != opcode.JMP {
			return errInvalidByteCode
		}

		ci.pc++

		th.dojmp(jmp)
	}

	return nil
}

func (th *thread) calluntm(a int, x object.Value, tag tagType) (err *object.RuntimeError) {
	tm := th.gettmbyobj(x, tag)

	if tm == nil {
		return th.unaryError(tag, x)
	}

	return th.calltm(a, tm, x)
}

func (th *thread) callbintm(a int, x, y object.Value, tag tagType) (err *object.RuntimeError) {
	tm := th.gettmbyobj(x, tag)
	if tm == nil {
		tm = th.gettmbyobj(y, tag)

		if tm == nil {
			return th.binaryError(tag, x, y)
		}
	}

	return th.calltm(a, tm, x, y)
}

func (th *thread) callordertm(not bool, x, y object.Value, tag tagType) (err *object.RuntimeError) {
	tm := th.gettmbyobj(x, tag)
	if tm == nil {
		tm = th.gettmbyobj(y, tag)

		if tm == nil {
			switch tag {
			case TM_LT:
				tm = th.gettmbyobj(x, TM_LE)
				if tm == nil {
					tm = th.gettmbyobj(y, TM_LE)

					if tm == nil {
						return th.compareError(x, y)
					}
				}

				x, y = y, x

				not = !not
			case TM_LE:
				tm = th.gettmbyobj(x, TM_LT)
				if tm == nil {
					tm = th.gettmbyobj(y, TM_LT)

					if tm == nil {
						return th.compareError(x, y)
					}
				}

				x, y = y, x

				not = !not
			default:
				panic("unreachable")
			}
		}
	}

	return th.callcmptm(not, tm, x, y)
}
