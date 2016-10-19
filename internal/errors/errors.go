package errors

import (
	"fmt"

	"github.com/hirochachacha/plua/object"
)

var (
	ErrDeadCoroutine    = object.NewRuntimeError("cannot resume dead coroutine")
	ErrGoroutineTwice   = object.NewRuntimeError("cannot resume goroutine twice")
	ErrInvalidByteCode  = object.NewRuntimeError("malformed bytecode detected")
	ErrYieldMainThread  = object.NewRuntimeError("attempt to yield a main thread")
	ErrYieldFromOutside = object.NewRuntimeError("attempt to yield from outside a coroutine")
	ErrYieldGoThread    = object.NewRuntimeError("attempt to yield a goroutine")
	ErrStackOverflow    = object.NewRuntimeError("Go stack overflow")
	ErrGetTable         = object.NewRuntimeError("gettable chain too long; possible loop")
	ErrSetTable         = object.NewRuntimeError("settable chain too long; possible loop")
	ErrNilIndex         = object.NewRuntimeError("table index is nil")
	ErrNaNIndex         = object.NewRuntimeError("table index is nan")
	ErrZeroDivision     = object.NewRuntimeError("attempt to divide by zero")
	ErrModuloByZero     = object.NewRuntimeError("attempt to modulo by zero")
	ErrInErrorHandling  = object.NewRuntimeError("error in error handling")
)

func ForLoopError(elem string) *object.RuntimeError {
	return object.NewRuntimeError(fmt.Sprintf("'for' %s value must be a number", elem))
}

func TypeError(op string, x object.Value) *object.RuntimeError {
	return object.NewRuntimeError(fmt.Sprintf("attempt to %s a %s value%s", op, object.ToType(x), varinfo(x)))
}

func CallError(fn object.Value) *object.RuntimeError {
	return TypeError("call", fn)
}

func IndexError(t object.Value) *object.RuntimeError {
	return TypeError("index", t)
}

func UnaryError(tag object.TagType, x object.Value) *object.RuntimeError {
	switch tag {
	case object.TM_LEN:
		return LengthError(x)
	case object.TM_UNM:
		return UnaryMinusError(x)
	case object.TM_BNOT:
		return BinaryNotError(x)
	default:
		panic("unreachable")
	}
}

func BinaryError(tag object.TagType, x, y object.Value) *object.RuntimeError {
	switch tag {
	case object.TM_ADD, object.TM_SUB, object.TM_MUL, object.TM_MOD, object.TM_POW, object.TM_DIV:
		return ArithError(x, y)
	case object.TM_IDIV, object.TM_BAND, object.TM_BOR, object.TM_BXOR, object.TM_SHL, object.TM_SHR:
		return BitwiseError(x, y)
	case object.TM_CONCAT:
		return ConcatError(x, y)
	default:
		panic("unreachable")
	}
}

func LengthError(x object.Value) *object.RuntimeError {
	return TypeError("get length of", x)
}

func UnaryMinusError(x object.Value) *object.RuntimeError {
	return TypeError("negate", x)
}

func BinaryNotError(x object.Value) *object.RuntimeError {
	return TypeError("bitwise negation on", x)
}

func CompareError(x, y object.Value) *object.RuntimeError {
	t1 := object.ToType(x)
	t2 := object.ToType(y)

	if t1 == t2 {
		return object.NewRuntimeError(fmt.Sprintf("attempt to compare two %s values", t1))
	}

	return object.NewRuntimeError(fmt.Sprintf("attempt to compare %s with %s", t1, t2))
}

func ArithError(x, y object.Value) *object.RuntimeError {
	if _, ok := object.ToNumber(x); ok {
		return TypeError("perform arithmetic on", y)
	}
	return TypeError("perform arithmetic on", x)
}

func BitwiseError(x, y object.Value) *object.RuntimeError {
	if _, ok := object.ToInteger(x); ok {
		if _, ok = y.(object.Number); ok {
			return object.NewRuntimeError(fmt.Sprintf("number%s has no integer representation", varinfo(y)))
		}
		return TypeError("perform bitwise operation on", y)
	}

	if _, ok := x.(object.Number); ok {
		return object.NewRuntimeError(fmt.Sprintf("number%s has no integer representation", varinfo(x)))
	}
	return TypeError("perform bitwise operation on", x)
}

func ConcatError(x, y object.Value) *object.RuntimeError {
	if _, ok := object.ToString(x); ok {
		return TypeError("concatenate", y)
	}
	return TypeError("concatenate", x)
}

func varinfo(x object.Value) string {
	// TODO? INCOMPATIBLE
	// Current implementation uses value instead of pointer everywhere.
	// So there is no way to identify two object is exactly same.
	// That's why I can't implement this.
	return ""
}
