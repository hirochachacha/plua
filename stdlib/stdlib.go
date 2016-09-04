package stdlib

import (
	"github.com/hirochachacha/plua/object"
	"github.com/hirochachacha/plua/stdlib/base"
)

func Open(th object.Thread, args ...object.Value) ([]object.Value, *object.RuntimeError) {
	th.Require("_G", base.Open)

	return nil, nil
}
