package stdlib

import (
	"github.com/hirochachacha/plua/object"
	"github.com/hirochachacha/plua/stdlib/base"
	"github.com/hirochachacha/plua/stdlib/coroutine"
	"github.com/hirochachacha/plua/stdlib/debug"
	"github.com/hirochachacha/plua/stdlib/goroutine"
	"github.com/hirochachacha/plua/stdlib/io"
	"github.com/hirochachacha/plua/stdlib/load"
	"github.com/hirochachacha/plua/stdlib/math"
	"github.com/hirochachacha/plua/stdlib/os"
	"github.com/hirochachacha/plua/stdlib/string"
	"github.com/hirochachacha/plua/stdlib/table"
	"github.com/hirochachacha/plua/stdlib/utf8"
)

func Open(th object.Thread, args ...object.Value) ([]object.Value, *object.RuntimeError) {
	th.Require("_G", base.Open)
	th.Require("debug", debug.Open)
	th.Require("coroutine", coroutine.Open)
	th.Require("goroutine", goroutine.Open)
	th.Require("io", io.Open)
	th.Require("math", math.Open)
	th.Require("os", os.Open)
	th.Require("package", load.Open)
	th.Require("string", string.Open)
	th.Require("table", table.Open)
	th.Require("utf8", utf8.Open)

	return nil, nil
}
