package object

import "math"

type Value interface {
	Type() Type
	String() string
}

var (
	True  = Boolean(true)
	False = Boolean(false)

	Infinity = Number(math.Inf(0))
	NaN      = Number(math.NaN())
)
