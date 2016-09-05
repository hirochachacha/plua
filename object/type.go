package object

import (
	"fmt"
	"unsafe"

	"github.com/hirochachacha/plua/internal/strconv"
)

type Integer int64

func (i Integer) Type() Type {
	return TNUMBER
}

func (i Integer) String() string {
	s, _ := ToGoString(i)
	return s
}

type Number float64

func (n Number) Type() Type {
	return TNUMBER
}

func (n Number) String() string {
	s, _ := ToGoString(n)
	return s
}

type String string

func (s String) Type() Type {
	return TSTRING
}

func (s String) String() string {
	return strconv.Quote(string(s))
}

type Boolean bool

func (b Boolean) Type() Type {
	return TBOOLEAN
}

func (b Boolean) String() string {
	if b {
		return "true"
	}

	return "false"
}

type LightUserdata struct {
	Pointer unsafe.Pointer
}

func (lud LightUserdata) Type() Type {
	return TUSERDATA
}

func (lud LightUserdata) String() string {
	return fmt.Sprintf("userdata: %p", lud.Pointer)
}

// GoFunction represents functions that can be called by Lua VM.
type GoFunction func(th Thread, args ...Value) (rets []Value, err *RuntimeError)

func (fn GoFunction) Type() Type {
	return TFUNCTION
}

func (fn GoFunction) String() string {
	return fmt.Sprintf("function: %p", fn)
}

type Userdata struct {
	Value     interface{}
	Metatable Table
}

func (ud *Userdata) Type() Type {
	return TUSERDATA
}

func (ud *Userdata) String() string {
	return fmt.Sprintf("userdata: %p", ud)
}

type Type int

func (t Type) String() string {
	return typeNames[t+1]
}

const (
	TNIL Type = iota
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
	TNIL + 1:           "nil",
	TBOOLEAN + 1:       "boolean",
	TLIGHTUSERDATA + 1: "userdata",
	TNUMBER + 1:        "number",
	TSTRING + 1:        "string",
	TTABLE + 1:         "table",
	TFUNCTION + 1:      "function",
	TUSERDATA + 1:      "userdata",
	TTHREAD + 1:        "thread",
	TNUMINT + 1:        "integer",
}
