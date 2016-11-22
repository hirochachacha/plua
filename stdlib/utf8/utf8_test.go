package utf8_test

import (
	"path/filepath"
	"testing"

	"github.com/hirochachacha/plua/compiler"
	"github.com/hirochachacha/plua/runtime"
	"github.com/hirochachacha/plua/stdlib/base"
	"github.com/hirochachacha/plua/stdlib/string"
	"github.com/hirochachacha/plua/stdlib/utf8"
)

func TestUTF8(t *testing.T) {
	c := compiler.NewCompiler()

	matches, err := filepath.Glob("testdata/*.lua")
	if err != nil {
		t.Fatal(err)
	}

	for _, fname := range matches {
		proto, err := c.CompileFile(fname, 0)
		if err != nil {
			t.Fatal(err)
		}

		p := runtime.NewProcess()

		p.Require("_G", base.Open)
		p.Require("utf8", utf8.Open)
		p.Require("string", string.Open)

		_, err = p.Exec(proto)
		if err != nil {
			t.Error(err)
		}
	}
}
