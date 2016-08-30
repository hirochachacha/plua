package runtime

import (
	"github.com/hirochachacha/plua/object"
)

type callInfo struct {
	*closure

	nrets int

	base int
	pc   int
	sp   int

	isTailCall bool

	prev *callInfo

	varargs []object.Value
}

func (ci *callInfo) isBase() bool {
	return ci.prev == nil
}

func (ci *callInfo) isGoFunction() bool {
	return ci.closure == nil
}
