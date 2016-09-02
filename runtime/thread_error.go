package runtime

import (
	"fmt"

	"errors"

	"github.com/hirochachacha/plua/object"
)

var (
	errDeadCoroutine  = errors.New("runtime: cannot resume dead coroutine")
	errGoroutineTwice = errors.New("runtime: cannot resume goroutine twice")
)

var protect = new(closure)

func (th *thread) varinfo(x object.Value) string {
	// TODO? INCOMPATIBLE
	// Current implementation uses value instead of pointer everywhere.
	// So there is no way to identify two object is exactly same.
	// That's why I can't implement this.
	return ""
}

func (th *thread) propagate(e *object.RuntimeError) {
	th.status = object.THREAD_ERROR
	th.data = e
}

func (th *thread) errorLevel(err *object.RuntimeError, level int) {
	if th.status != object.THREAD_ERROR {
		if err.Level > 0 {
			d := th.getInfo(level, "Sl")
			if d != nil {
				err.Pos.Filename = "@" + d.ShortSource
				err.Pos.Line = d.CurrentLine
			}
		}

		th.propagate(err)
	}
}

func (th *thread) error(err *object.RuntimeError) {
	th.errorLevel(err, 1)
}

func (th *thread) throwInternalError(msg string) {
	th.errorLevel(&object.RuntimeError{Value: object.String(msg)}, 0)
}

func (th *thread) throwByteCodeError() {
	th.throwInternalError("malformed bytecode detected")
}

func (th *thread) throwRuntimeError(msg string) {
	th.throwInternalError(msg)
}

func (th *thread) throwYieldMainThreadError() {
	th.throwRuntimeError("attempt to yield a main thread")
}

func (th *thread) throwYieldFromOutsideError() {
	th.throwRuntimeError("attempt to yield from outside a coroutine")
}

func (th *thread) throwYieldGoThreadError() {
	th.throwRuntimeError("attempt to yield a goroutine")
}

func (th *thread) throwForError(elem string) {
	th.throwRuntimeError(fmt.Sprintf("'for' %s value must be a number", elem))
}

func (th *thread) throwStackOverflowError() {
	th.throwRuntimeError("Go stack overflow")
}

func (th *thread) throwGetTableError() {
	th.throwRuntimeError("gettable chain too long; possible loop")
}

func (th *thread) throwSetTableError() {
	th.throwRuntimeError("settable chain too long; possible loop")
}

func (th *thread) throwTypeError(op string, x object.Value) {
	th.throwRuntimeError(fmt.Sprintf("attempt to %s a %s value%s", op, object.ToType(x), th.varinfo(x)))
}

func (th *thread) throwCallError(fn object.Value) {
	th.throwTypeError("call", fn)
}

func (th *thread) throwIndexError(t object.Value) {
	th.throwTypeError("index", t)
}

func (th *thread) throwNilIndexError() {
	th.throwRuntimeError("table index is nil")
}

func (th *thread) throwNaNIndexError() {
	th.throwRuntimeError("table index is nan")
}

func (th *thread) throwZeroDivisionError() {
	th.throwRuntimeError("attempt to divide by zero")
}

func (th *thread) throwModuloByZeroError() {
	th.throwRuntimeError("attempt to modulo by zero")
}

func (th *thread) throwUnaryError(tag tagType, x object.Value) {
	switch tag {
	case TM_LEN:
		th.throwLengthError(x)
	case TM_UNM:
		th.throwUnaryMinusError(x)
	case TM_BNOT:
		th.throwBinaryNotError(x)
	default:
		panic("unreachable")
	}
}

func (th *thread) throwBinaryError(tag tagType, x, y object.Value) {
	switch tag {
	case TM_ADD, TM_SUB, TM_MUL, TM_MOD, TM_POW, TM_DIV:
		th.throwArithError(x, y)
	case TM_IDIV, TM_BAND, TM_BOR, TM_BXOR, TM_SHL, TM_SHR:
		th.throwBitwiseError(x, y)
	case TM_CONCAT:
		th.throwConcatError(x, y)
	default:
		panic("unreachable")
	}
}

func (th *thread) throwLengthError(x object.Value) {
	th.throwTypeError("get length of", x)
}

func (th *thread) throwUnaryMinusError(x object.Value) {
	th.throwTypeError("negate", x)
}

func (th *thread) throwBinaryNotError(x object.Value) {
	th.throwTypeError("bitwise negation on", x)
}

func (th *thread) throwCompareError(x, y object.Value) {
	t1 := object.ToType(x)
	t2 := object.ToType(y)

	if t1 == t2 {
		th.throwRuntimeError(fmt.Sprintf("attempt to compare two %s values", t1))
	} else {
		th.throwRuntimeError(fmt.Sprintf("attempt to compare %s with %s", t1, t2))
	}
}

func (th *thread) throwArithError(x, y object.Value) {
	if _, ok := object.ToNumber(x); ok {
		th.throwTypeError("perform arithmetic on", y)
	} else {
		th.throwTypeError("perform arithmetic on", x)
	}
}

func (th *thread) throwBitwiseError(x, y object.Value) {
	if _, ok := object.ToInteger(x); ok {
		if _, ok = y.(object.Number); ok {
			th.throwRuntimeError(fmt.Sprintf("number%s has no integer representation", th.varinfo(y)))
		} else {
			th.throwTypeError("perform bitwise operation on", y)
		}
	} else {
		if _, ok = x.(object.Number); ok {
			th.throwRuntimeError(fmt.Sprintf("number%s has no integer representation", th.varinfo(x)))
		} else {
			th.throwTypeError("perform bitwise operation on", x)
		}
	}
}

func (th *thread) throwConcatError(x, y object.Value) {
	if _, ok := object.ToString(x); ok {
		th.throwTypeError("concatenate", y)
	} else {
		th.throwTypeError("concatenate", x)
	}
}
