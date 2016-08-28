package object

import (
	"math"
	"unsafe"

	"github.com/hirochachacha/blua/errors"
	"github.com/hirochachacha/blua/position"
)

type Value interface {
	Type() Type
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

func (i Integer) Type() Type {
	return TNUMBER
}

type Number float64

func (n Number) Type() Type {
	return TNUMBER
}

type String string

func (s String) Type() Type {
	return TSTRING
}

type Boolean bool

func (b Boolean) Type() Type {
	return TBOOLEAN
}

// see https://code.google.com/p/go/issues/detail?id=6116
type LightUserdata struct {
	Pointer unsafe.Pointer
}

func (lud LightUserdata) Type() Type {
	return TUSERDATA
}

type GoFunction func(th Thread, args ...Value) (rets []Value)

func (fn GoFunction) Type() Type {
	return TFUNCTION
}
