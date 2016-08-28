package runtime

import (
	"fmt"

	"github.com/hirochachacha/blua/errors"
	"github.com/hirochachacha/blua/object"
	"github.com/hirochachacha/blua/position"
)

var protect = new(closure) // just make a stub

var (
	errDeadCoroutine  = errors.RuntimeError.New("cannot resume dead coroutine")
	errGoroutineTwice = errors.RuntimeError.New("cannot resume goroutine twice")
)

func (th *thread) where(msg object.String) object.String {
	ci := th.ci

	if !ci.isGoFunction() {
		line := getCurrentLine(ci)

		if len(ci.Source) == 0 {
			return object.String(fmt.Sprintf("?:%d: %s", line, msg))
		}

		return object.String(fmt.Sprintf("%s:%d: %s", shorten(ci.Source), line, msg))
	}

	return msg
}

func (th *thread) varinfo(x object.Value) string {
	// TODO? INCOMPATIBLE
	// Current implementation uses value instead of pointer everywhere.
	// So there is no way to identify two object is exactly same.
	// That's why I can't implement this.
	return ""
}

func (th *thread) propagate(err *object.Error) {
	th.status = object.THREAD_ERROR
	th.data = err
}

func (th *thread) throwInternalError(msg object.Value, where bool) {
	ctx := th.context

	if ctx.status != object.THREAD_ERROR {
		pos := position.Position{}

		if where {
			d := th.getInfo(0, "Sl")
			if d != nil {
				pos.Filename = "@" + d.ShortSource
				pos.Line = d.CurrentLine
			}
		}

		err := &object.Error{
			Value: msg,
			Pos:   pos,
		}

		th.propagate(err)
	}
}

func (th *thread) throwByteCodeError() {
	th.throwInternalError(object.String("malformed bytecode detected"), false)
}

func (th *thread) throwRuntimeError(msg string) {
	th.throwInternalError(object.String(msg), true)
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
