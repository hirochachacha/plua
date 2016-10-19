package table

import (
	"github.com/hirochachacha/plua/internal/arith"
	"github.com/hirochachacha/plua/object"
)

func callLen(th object.Thread, t object.Value) (int, *object.RuntimeError) {
	l, err := arith.CallLen(th, t)
	if err != nil {
		return 0, err
	}
	tlen, ok := object.ToGoInt(l)
	if !ok {
		return 0, object.NewRuntimeError("object length is not an integer")
	}
	return tlen, nil
}
