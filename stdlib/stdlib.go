package stdlib

import (
	"github.com/hirochachacha/plua/object"
	"github.com/hirochachacha/plua/stdlib/base"
	"github.com/hirochachacha/plua/stdlib/debug"
)

func Open(th object.Thread, args ...object.Value) ([]object.Value, *object.RuntimeError) {
	th.Require("_G", base.Open)
	th.Require("debug", debug.Open)

	return nil, nil
}
