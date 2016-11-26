package arith

import (
	"github.com/hirochachacha/plua/internal/errors"
	"github.com/hirochachacha/plua/object"
)

func CallLen(th object.Thread, x object.Value) (object.Value, *object.RuntimeError) {
	switch x := x.(type) {
	case object.String:
		return object.Integer(len(x)), nil
	case object.Table:
		tm := gettm(x.Metatable(), object.TM_LEN)
		if tm != nil {
			return calltm(th, tm, x)
		}
		return object.Integer(x.Len()), nil
	default:
		return calluntm(th, x, object.TM_LEN)
	}
}

func CallUnm(th object.Thread, x object.Value) (object.Value, *object.RuntimeError) {
	if unm := Unm(x); unm != nil {
		return unm, nil
	}
	return calluntm(th, x, object.TM_UNM)
}

func CallBnot(th object.Thread, x object.Value) (object.Value, *object.RuntimeError) {
	if unm := Bnot(x); unm != nil {
		return unm, nil
	}
	return calluntm(th, x, object.TM_BNOT)
}

func CallEqual(th object.Thread, not bool, x, y object.Value) (bool, *object.RuntimeError) {
	// fast path for avoiding assertI2I2
	eq := object.Equal(x, y)
	if eq {
		return true != not, nil
	}

	switch x := x.(type) {
	case object.Table:
		if y, ok := y.(object.Table); ok {
			tm := gettm(x.Metatable(), object.TM_EQ)
			if tm == nil {
				tm = gettm(y.Metatable(), object.TM_EQ)
				if tm == nil {
					return false != not, nil
				}
			}

			return callcmptm(th, not, tm, x, y)
		}

		return false != not, nil
	case *object.Userdata:
		if y, ok := y.(*object.Userdata); ok {
			tm := gettm(x.Metatable, object.TM_EQ)
			if tm == nil {
				tm = gettm(y.Metatable, object.TM_EQ)
				if tm == nil {
					return false != not, nil
				}
			}

			return callcmptm(th, not, tm, x, y)
		}

		return false != not, nil
	default:
		return eq != not, nil
	}
}

func CallLessThan(th object.Thread, not bool, x, y object.Value) (bool, *object.RuntimeError) {
	if b := LessThan(x, y); b != nil {
		return b != object.Boolean(not), nil
	}
	return callordertm(th, not, x, y, object.TM_LT)
}

func CallLessThanOrEqualTo(th object.Thread, not bool, x, y object.Value) (bool, *object.RuntimeError) {
	if b := LessThanOrEqualTo(x, y); b != nil {
		return b != object.Boolean(not), nil
	}
	return callordertm(th, not, x, y, object.TM_LE)
}

func CallAdd(th object.Thread, x, y object.Value) (object.Value, *object.RuntimeError) {
	if sum := Add(x, y); sum != nil {
		return sum, nil
	}
	return callbintm(th, x, y, object.TM_ADD)
}

func CallSub(th object.Thread, x, y object.Value) (object.Value, *object.RuntimeError) {
	if sum := Sub(x, y); sum != nil {
		return sum, nil
	}
	return callbintm(th, x, y, object.TM_SUB)
}

func CallMul(th object.Thread, x, y object.Value) (object.Value, *object.RuntimeError) {
	if prod := Mul(x, y); prod != nil {
		return prod, nil
	}
	return callbintm(th, x, y, object.TM_MUL)
}

func CallDiv(th object.Thread, x, y object.Value) (object.Value, *object.RuntimeError) {
	if quo := Div(x, y); quo != nil {
		return quo, nil
	}
	return callbintm(th, x, y, object.TM_DIV)
}

func CallIdiv(th object.Thread, x, y object.Value) (object.Value, *object.RuntimeError) {
	quo, ok := Idiv(x, y)
	if !ok {
		return nil, errors.ErrZeroDivision
	}
	if quo != nil {
		return quo, nil
	}
	return callbintm(th, x, y, object.TM_IDIV)
}

func CallMod(th object.Thread, x, y object.Value) (object.Value, *object.RuntimeError) {
	rem, ok := Mod(x, y)
	if !ok {
		return nil, errors.ErrModuloByZero
	}
	if rem != nil {
		return rem, nil
	}
	return callbintm(th, x, y, object.TM_MOD)
}

func CallPow(th object.Thread, x, y object.Value) (object.Value, *object.RuntimeError) {
	if prod := Pow(x, y); prod != nil {
		return prod, nil
	}
	return callbintm(th, x, y, object.TM_POW)
}

func CallBand(th object.Thread, x, y object.Value) (object.Value, *object.RuntimeError) {
	if band := Band(x, y); band != nil {
		return band, nil
	}
	return callbintm(th, x, y, object.TM_BAND)
}

func CallBor(th object.Thread, x, y object.Value) (object.Value, *object.RuntimeError) {
	if bor := Bor(x, y); bor != nil {
		return bor, nil
	}
	return callbintm(th, x, y, object.TM_BOR)
}

func CallBxor(th object.Thread, x, y object.Value) (object.Value, *object.RuntimeError) {
	if bxor := Bxor(x, y); bxor != nil {
		return bxor, nil
	}
	return callbintm(th, x, y, object.TM_BXOR)
}

func CallShl(th object.Thread, x, y object.Value) (object.Value, *object.RuntimeError) {
	if shl := Shl(x, y); shl != nil {
		return shl, nil
	}
	return callbintm(th, x, y, object.TM_SHL)
}

func CallShr(th object.Thread, x, y object.Value) (object.Value, *object.RuntimeError) {
	if shr := Shr(x, y); shr != nil {
		return shr, nil
	}
	return callbintm(th, x, y, object.TM_SHR)
}

func CallConcat(th object.Thread, x, y object.Value) (object.Value, *object.RuntimeError) {
	if con := Concat(x, y); con != nil {
		return con, nil
	}
	tm := gettmbyobj(th, x, object.TM_CONCAT)
	if tm == nil {
		tm = gettmbyobj(th, y, object.TM_CONCAT)
		if tm == nil {
			return nil, errors.BinaryError(object.TM_CONCAT, x, y)
		}
	}
	return calltm(th, tm, x, y)
}

func gettm(mt object.Table, tag object.Value) object.Value {
	if mt == nil {
		return nil
	}
	return mt.Get(tag)
}

func gettmbyobj(th object.Thread, x object.Value, tag object.Value) object.Value {
	mt := th.GetMetatable(x)
	if mt == nil {
		return nil
	}
	return gettm(mt, tag)
}

func calltm(th object.Thread, tm object.Value, args ...object.Value) (object.Value, *object.RuntimeError) {
	rets, err := th.Call(tm, nil, args...)
	if err != nil {
		return nil, err
	}
	if len(rets) == 0 {
		return nil, nil
	}
	return rets[0], nil
}

func callcmptm(th object.Thread, not bool, tm object.Value, x, y object.Value) (bool, *object.RuntimeError) {
	rets, err := th.Call(tm, nil, x, y)
	if err != nil {
		return false, err
	}

	var ret object.Value

	if len(rets) != 0 {
		ret = rets[0]
	}

	return object.ToGoBool(ret) != not, nil
}

func calluntm(th object.Thread, x object.Value, tag object.Value) (object.Value, *object.RuntimeError) {
	tm := gettmbyobj(th, x, tag)
	if tm == nil {
		return nil, errors.UnaryError(tag, x)
	}
	return calltm(th, tm, x)
}

func callbintm(th object.Thread, x, y object.Value, tag object.Value) (object.Value, *object.RuntimeError) {
	tm := gettmbyobj(th, x, tag)
	if tm == nil {
		tm = gettmbyobj(th, y, tag)
		if tm == nil {
			return nil, errors.BinaryError(tag, x, y)
		}
	}
	return calltm(th, tm, x, y)
}

func callordertm(th object.Thread, not bool, x, y object.Value, tag object.Value) (bool, *object.RuntimeError) {
	tm := gettmbyobj(th, x, tag)
	if tm == nil {
		tm = gettmbyobj(th, y, tag)
		if tm == nil {
			switch tag {
			case object.TM_LT:
				tm = gettmbyobj(th, x, object.TM_LE)
				if tm == nil {
					tm = gettmbyobj(th, y, object.TM_LE)
					if tm == nil {
						return false, errors.CompareError(x, y)
					}
				}

				x, y = y, x

				not = !not
			case object.TM_LE:
				tm = gettmbyobj(th, x, object.TM_LT)
				if tm == nil {
					tm = gettmbyobj(th, y, object.TM_LT)
					if tm == nil {
						return false, errors.CompareError(x, y)
					}
				}

				x, y = y, x

				not = !not
			default:
				panic("unreachable")
			}
		}
	}

	return callcmptm(th, not, tm, x, y)
}
