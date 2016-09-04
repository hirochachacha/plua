package runtime

import (
	"fmt"

	"github.com/hirochachacha/plua/object"
)

var (
	errDeadCoroutine    = object.NewRuntimeError("cannot resume dead coroutine")
	errGoroutineTwice   = object.NewRuntimeError("cannot resume goroutine twice")
	errInvalidByteCode  = object.NewRuntimeError("malformed bytecode detected")
	errYieldMainThread  = object.NewRuntimeError("attempt to yield a main thread")
	errYieldFromOutside = object.NewRuntimeError("attempt to yield from outside a coroutine")
	errYieldGoThread    = object.NewRuntimeError("attempt to yield a goroutine")
	errStackOverflow    = object.NewRuntimeError("Go stack overflow")
	errGetTable         = object.NewRuntimeError("gettable chain too long; possible loop")
	errSetTable         = object.NewRuntimeError("settable chain too long; possible loop")
	errNilIndex         = object.NewRuntimeError("table index is nil")
	errNaNIndex         = object.NewRuntimeError("table index is nan")
	errZeroDivision     = object.NewRuntimeError("attempt to divide by zero")
	errModuloByZero     = object.NewRuntimeError("attempt to modulo by zero")
	errInErrorHandling  = object.NewRuntimeError("error in error handling")
)

func (th *thread) forLoopError(elem string) *object.RuntimeError {
	return object.NewRuntimeError(fmt.Sprintf("'for' %s value must be a number", elem))
}

func (th *thread) typeError(op string, x object.Value) *object.RuntimeError {
	return object.NewRuntimeError(fmt.Sprintf("attempt to %s a %s value%s", op, object.ToType(x), th.varinfo(x)))
}

func (th *thread) callError(fn object.Value) *object.RuntimeError {
	return th.typeError("call", fn)
}

func (th *thread) indexError(t object.Value) *object.RuntimeError {
	return th.typeError("index", t)
}

func (th *thread) unaryError(tag tagType, x object.Value) *object.RuntimeError {
	switch tag {
	case TM_LEN:
		return th.lengthError(x)
	case TM_UNM:
		return th.unaryMinusError(x)
	case TM_BNOT:
		return th.binaryNotError(x)
	default:
		panic("unreachable")
	}
}

func (th *thread) binaryError(tag tagType, x, y object.Value) *object.RuntimeError {
	switch tag {
	case TM_ADD, TM_SUB, TM_MUL, TM_MOD, TM_POW, TM_DIV:
		return th.arithError(x, y)
	case TM_IDIV, TM_BAND, TM_BOR, TM_BXOR, TM_SHL, TM_SHR:
		return th.bitwiseError(x, y)
	case TM_CONCAT:
		return th.concatError(x, y)
	default:
		panic("unreachable")
	}
}

func (th *thread) lengthError(x object.Value) *object.RuntimeError {
	return th.typeError("get length of", x)
}

func (th *thread) unaryMinusError(x object.Value) *object.RuntimeError {
	return th.typeError("negate", x)
}

func (th *thread) binaryNotError(x object.Value) *object.RuntimeError {
	return th.typeError("bitwise negation on", x)
}

func (th *thread) compareError(x, y object.Value) *object.RuntimeError {
	t1 := object.ToType(x)
	t2 := object.ToType(y)

	if t1 == t2 {
		return object.NewRuntimeError(fmt.Sprintf("attempt to compare two %s values", t1))
	}

	return object.NewRuntimeError(fmt.Sprintf("attempt to compare %s with %s", t1, t2))
}

func (th *thread) arithError(x, y object.Value) *object.RuntimeError {
	if _, ok := object.ToNumber(x); ok {
		return th.typeError("perform arithmetic on", y)
	}
	return th.typeError("perform arithmetic on", x)
}

func (th *thread) bitwiseError(x, y object.Value) *object.RuntimeError {
	if _, ok := object.ToInteger(x); ok {
		if _, ok = y.(object.Number); ok {
			return object.NewRuntimeError(fmt.Sprintf("number%s has no integer representation", th.varinfo(y)))
		}
		return th.typeError("perform bitwise operation on", y)
	}

	if _, ok := x.(object.Number); ok {
		return object.NewRuntimeError(fmt.Sprintf("number%s has no integer representation", th.varinfo(x)))
	}
	return th.typeError("perform bitwise operation on", x)
}

func (th *thread) concatError(x, y object.Value) *object.RuntimeError {
	if _, ok := object.ToString(x); ok {
		return th.typeError("concatenate", y)
	}
	return th.typeError("concatenate", x)
}

func (th *thread) varinfo(x object.Value) string {
	// TODO? INCOMPATIBLE
	// Current implementation uses value instead of pointer everywhere.
	// So there is no way to identify two object is exactly same.
	// That's why I can't implement this.
	return ""
}

func (th *thread) error(err *object.RuntimeError) {
	if th.status != object.THREAD_ERROR {
		th.status = object.THREAD_ERROR
		th.data = err
	}
}
