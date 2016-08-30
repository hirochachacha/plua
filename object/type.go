package object

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
