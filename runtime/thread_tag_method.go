package runtime

import (
	"github.com/hirochachacha/plua/object"
	"github.com/hirochachacha/plua/opcode"
	"github.com/hirochachacha/plua/runtime/internal/errors"
)

func (th *thread) fasttm(mt object.Table, tag object.TagType) object.Value {
	return th.gettm(mt, tag)
}

func (th *thread) gettm(mt object.Table, tag object.TagType) object.Value {
	if mt == nil {
		return nil
	}

	return mt.Get(object.String(tag.String()))
}

func (th *thread) gettmbyobj(val object.Value, tag object.TagType) object.Value {
	mt := th.env.getMetatable(val)
	if mt == nil {
		return nil
	}

	return th.gettm(mt, tag)
}

func (th *thread) calltm(a int, tm object.Value, args ...object.Value) (err *object.RuntimeError) {
	rets, err := th.docall(tm, nil, args...)
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
	rets, err := th.docall(tm, nil, x, y)
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
			return errors.ErrInvalidByteCode
		}

		ci.pc++

		th.dojmp(jmp)
	}

	return nil
}

func (th *thread) calluntm(a int, x object.Value, tag object.TagType) (err *object.RuntimeError) {
	tm := th.gettmbyobj(x, tag)

	if tm == nil {
		return errors.UnaryError(tag, x)
	}

	return th.calltm(a, tm, x)
}

func (th *thread) callbintm(a int, x, y object.Value, tag object.TagType) (err *object.RuntimeError) {
	tm := th.gettmbyobj(x, tag)
	if tm == nil {
		tm = th.gettmbyobj(y, tag)

		if tm == nil {
			return errors.BinaryError(tag, x, y)
		}
	}

	return th.calltm(a, tm, x, y)
}

func (th *thread) callordertm(not bool, x, y object.Value, tag object.TagType) (err *object.RuntimeError) {
	tm := th.gettmbyobj(x, tag)
	if tm == nil {
		tm = th.gettmbyobj(y, tag)

		if tm == nil {
			switch tag {
			case object.TM_LT:
				tm = th.gettmbyobj(x, object.TM_LE)
				if tm == nil {
					tm = th.gettmbyobj(y, object.TM_LE)

					if tm == nil {
						return errors.CompareError(x, y)
					}
				}

				x, y = y, x

				not = !not
			case object.TM_LE:
				tm = th.gettmbyobj(x, object.TM_LT)
				if tm == nil {
					tm = th.gettmbyobj(y, object.TM_LT)

					if tm == nil {
						return errors.CompareError(x, y)
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
