package object

import "unsafe"

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
type GoFunction func(th Thread, args ...Value) (rets []Value, err *RuntimeError)

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

type Type int

func (t Type) String() string {
	return typeNames[t+1]
}

const (
	TNONE Type = iota - 1
	TNIL
	TBOOLEAN
	TLIGHTUSERDATA
	TNUMBER
	TSTRING
	TTABLE
	TFUNCTION
	TUSERDATA
	TTHREAD
	TCHANNEL

	MaxType
)

const (
	TSHRSTR Type = TSTRING
	TLNGSTR Type = TSTRING | (1 << 4)

	TNUMFLT Type = TNUMBER
	TNUMINT Type = TNUMBER | (1 << 4)
)

var typeNames = [...]string{
	TNONE + 1:          "none",
	TNIL + 1:           "nil",
	TBOOLEAN + 1:       "boolean",
	TLIGHTUSERDATA + 1: "userdata",
	TNUMBER + 1:        "number",
	TSTRING + 1:        "string",
	TTABLE + 1:         "table",
	TFUNCTION + 1:      "function",
	TUSERDATA + 1:      "userdata",
	TTHREAD + 1:        "thread",
}
