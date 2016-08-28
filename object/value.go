package object

import (
	"math"
	"unsafe"

	"github.com/hirochachacha/blua/errors"
	"github.com/hirochachacha/blua/position"
)

type Value interface {
	value()
}

var (
	True  = Boolean(true)
	False = Boolean(false)

	Infinity = Number(math.Inf(0))
	NaN      = Number(math.NaN())
)

type Error struct {
	Value Value
	Pos   position.Position
}

func (err *Error) Message() Value {
	if msg, ok := ToGoString(err.Value); ok {
		if err.Pos.IsValid() {
			msg = err.Pos.String() + ": " + msg
		}
		return String(msg)
	}
	return err.Value
}

func (err *Error) Error() string {
	if msg, ok := ToGoString(err.Value); ok {
		return msg
	}

	return "(error object is a " + ToType(err.Value).String() + " value)"
}

func (err *Error) NewRuntimeError() error {
	return errors.RuntimeError.WrapWith(err.Pos, err)
}

type Integer int64

type Number float64

type String string

type Boolean bool

// see https://code.google.com/p/go/issues/detail?id=6116
type LightUserdata struct {
	Pointer unsafe.Pointer
}

type GoFunction func(th *Thread, args ...Value) (rets []Value)

// dynamic values

type Table struct {
	Impl iTable
}

type Userdata struct {
	Impl iUserdata
}

type Closure struct {
	Impl iClosure
}

type Thread struct {
	Impl iThread

	hasReflection bool
}

type Channel struct {
	Impl iChannel
}

func (i Integer) value() {}

func (n Number) value() {}

func (s String) value() {}

func (b Boolean) value() {}

func (lud LightUserdata) value() {}

func (fn GoFunction) value() {}

func (t *Table) value() {}

func (ud *Userdata) value() {}

func (cl *Closure) value() {}

func (t *Thread) value() {}

func (ch *Channel) value() {}
