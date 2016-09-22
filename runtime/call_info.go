package runtime

import (
	"github.com/hirochachacha/plua/object"
)

type callInfo struct {
	*closure

	nrets int

	base int
	pc   int
	top  int

	isTailCall bool

	varargs []object.Value
}

func (ci *callInfo) isGoFunction() bool {
	return ci.closure == nil
}

func (ci *callInfo) isBottom() bool {
	return ci.base == 2
}
