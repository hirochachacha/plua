package utf8_test

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/hirochachacha/plua/compiler"
	"github.com/hirochachacha/plua/runtime"
	"github.com/hirochachacha/plua/stdlib/base"
	"github.com/hirochachacha/plua/stdlib/utf8"
)

func TestUTF8(t *testing.T) {
	c := compiler.NewCompiler()

	matches, err := filepath.Glob("testdata/*.lua")
	if err != nil {
		t.Fatal(err)
	}

	for _, fname := range matches {
		f, err := os.Open(fname)
		if err != nil {
			t.Fatal(err)
		}

		proto, err := c.Compile(f, "@"+fname)
		if err != nil {
			t.Fatal(err)
		}

		p := runtime.NewProcess()

		p.Require("_G", base.Open)
		p.Require("utf8", utf8.Open)

		_, err = p.Exec(proto)
		if err != nil {
			t.Error(err)
		}
	}
}