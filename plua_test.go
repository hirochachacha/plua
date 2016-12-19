package plua_test

import (
	"fmt"
	"os"
	"testing"

	"github.com/hirochachacha/plua/compiler"
	"github.com/hirochachacha/plua/object"
	"github.com/hirochachacha/plua/runtime"
	"github.com/hirochachacha/plua/stdlib"
)

func TestLuaTest(t *testing.T) {
	wd, err := os.Getwd()
	if err != nil {
		t.Fatal(err)
	}

	err = os.Chdir("testdata/lua-5.3.3-tests")
	if err != nil {
		t.Fatal(err)
	}
	defer func() {
		err = os.Chdir(wd)
		if err != nil {
			t.Fatal(err)
		}
	}()

	c := compiler.NewCompiler()

	proto, err := c.CompileFile("all.lua", 0)
	if err != nil {
		t.Fatal(err)
	}

	p := runtime.NewProcess()

	p.Require("", stdlib.Open)

	p.Globals().Set(object.String("_U"), object.True) // set user test flag

	rets, err := p.Exec(proto)
	if err != nil {
		t.Fatal(err)
	}

	fmt.Println(rets)
}
