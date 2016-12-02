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
	ErrModuloByZero     = object.NewRuntimeError("attempt to perform 'n%0'")
	ErrInErrorHandling  = object.NewRuntimeError("error in error handling")
)

func ForLoopError(elem string) *object.RuntimeError {
	return object.NewRuntimeError(fmt.Sprintf("'for' %s value must be a number", elem))
}

func typeName(th object.Thread, arg object.Value) string {
	if mt := th.GetMetatable(arg); mt != nil {
		if name := mt.Get(object.TM_NAME); name != nil {
			if name, ok := name.(object.String); ok {
				return string(name)
			}
			if _, ok := arg.(object.LightUserdata); ok {
				return "light userdata"
			}
			return object.ToType(arg).String()
		}
	}
	if _, ok := arg.(object.LightUserdata); ok {
		return "light userdata"
	}
	return object.ToType(arg).String()
}

func TypeError(th object.Thread, op string, x object.Value) *object.RuntimeError {
	return object.NewRuntimeError(fmt.Sprintf("attempt to %s a %s value%s", op, typeName(th, x), varinfo(th, x)))
}

func CallError(th object.Thread, fn object.Value) *object.RuntimeError {
	return TypeError(th, "call", fn)
}

func IndexError(th object.Thread, t object.Value) *object.RuntimeError {
	return TypeError(th, "index", t)
}

func UnaryError(th object.Thread, tag object.Value, x object.Value) *object.RuntimeError {
	switch tag {
	case object.TM_LEN:
		return LengthError(th, x)
	case object.TM_UNM:
		return UnaryMinusError(th, x)
	case object.TM_BNOT:
		return BinaryNotError(th, x)
	default:
		panic("unreachable")
	}
}

func BinaryError(th object.Thread, tag object.Value, x, y object.Value) *object.RuntimeError {
	switch tag {
	case object.TM_ADD, object.TM_SUB, object.TM_MUL, object.TM_MOD, object.TM_POW, object.TM_DIV:
		return ArithError(th, x, y)
	case object.TM_IDIV, object.TM_BAND, object.TM_BOR, object.TM_BXOR, object.TM_SHL, object.TM_SHR:
		return BitwiseError(th, x, y)
	case object.TM_CONCAT:
		return ConcatError(th, x, y)
	default:
		panic("unreachable")
	}
}

func LengthError(th object.Thread, x object.Value) *object.RuntimeError {
	return TypeError(th, "get length of", x)
}

func UnaryMinusError(th object.Thread, x object.Value) *object.RuntimeError {
	return TypeError(th, "negate", x)
}

func BinaryNotError(th object.Thread, x object.Value) *object.RuntimeError {
	if _, ok := object.ToNumber(x); ok {
		return object.NewRuntimeError(fmt.Sprintf("number%s has no integer representation", varinfo(th, x)))
	}
	return TypeError(th, "bitwise negation on", x)
}

func CompareError(th object.Thread, x, y object.Value) *object.RuntimeError {
	t1 := typeName(th, x)
	t2 := typeName(th, y)

	if t1 == t2 {
		return object.NewRuntimeError(fmt.Sprintf("attempt to compare two %s values", t1))
	}

	return object.NewRuntimeError(fmt.Sprintf("attempt to compare %s with %s", t1, t2))
}

func ArithError(th object.Thread, x, y object.Value) *object.RuntimeError {
	if _, ok := object.ToNumber(x); ok {
		return TypeError(th, "perform arithmetic on", y)
	}
	return TypeError(th, "perform arithmetic on", x)
}

func BitwiseError(th object.Thread, x, y object.Value) *object.RuntimeError {
	if _, ok := object.ToInteger(x); ok {
		if _, ok := object.ToNumber(y); ok {
			return object.NewRuntimeError(fmt.Sprintf("number%s has no integer representation", varinfo(th, y)))
		}
		return TypeError(th, "perform bitwise operation on", y)
	}
	if _, ok := object.ToNumber(x); ok {
		return object.NewRuntimeError(fmt.Sprintf("number%s has no integer representation", varinfo(th, x)))
	}
	return TypeError(th, "perform bitwise operation on", x)
}

func ConcatError(th object.Thread, x, y object.Value) *object.RuntimeError {
	if _, ok := object.ToString(x); ok {
		return TypeError(th, "concatenate", y)
	}
	return TypeError(th, "concatenate", x)
}

func varinfo(th object.Thread, x object.Value) string {
	// TODO? INCOMPATIBLE
	// Current implementation uses value instead of pointer everywhere.
	// So there is no way to identify two object is exactly same.
	// That's why I can't implement this.
	return ""
}
