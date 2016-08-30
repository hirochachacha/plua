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

	NoErr = none{}
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

type LightUserdata struct {
	Pointer unsafe.Pointer
}

func (lud LightUserdata) Type() Type {
	return TUSERDATA
}

// GoFunction represents functions that can be called by Lua VM.
// Because of nil is a valid value, you cannot use nil as 'no error'.
// If error is none, you must return NoErr instead of nil.
type GoFunction func(th Thread, args ...Value) (rets []Value, err Value)

func (fn GoFunction) Type() Type {
	return TFUNCTION
}

type Userdata struct {
	Value     interface{}
	Metatable Table
}

func (ud *Userdata) Type() Type {
	return TUSERDATA
}

type none struct{}

func (n none) Type() Type {
	return TNONE
}
