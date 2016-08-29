package object

import (
	"math"
	"unsafe"
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

type GoFunction func(th Thread, args ...Value) (rets []Value, err Value)

func (fn GoFunction) Type() Type {
	return TFUNCTION
}
