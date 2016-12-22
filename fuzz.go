// +build gofuzz

package plua

import (
	"io/ioutil"
	"os"

	"github.com/hirochachacha/plua/compiler"
	"github.com/hirochachacha/plua/compiler/ast/printer"
	"github.com/hirochachacha/plua/compiler/parser"
	"github.com/hirochachacha/plua/object"
	"github.com/hirochachacha/plua/runtime"
	"github.com/hirochachacha/plua/stdlib"
)

func Fuzz(data []byte) int {
	f, err := ioutil.TempFile("", "plua")
	if err != nil {
		return 0
	}
	defer os.Remove(f.Name())
	defer f.Close()

	_, err = f.Write(data)
	if err != nil {
		return 0
	}

	var text bool

	if ast, err := parser.ParseFile(f.Name(), parser.ParseComments); err == nil {
		tmp, err := ioutil.TempFile("", "plua")
		if err != nil {
			return 0
		}
		defer os.Remove(tmp.Name())
		defer tmp.Close()

		err = printer.Fprint(tmp, ast)
		if err != nil {
			panic(err)
		}

		_, err = parser.ParseFile(tmp.Name(), parser.ParseComments)
		if err != nil {
			panic(err)
		}

		text = true
	}

	c := compiler.NewCompiler()

	proto, err := c.CompileFile(f.Name(), compiler.Either)
	if err != nil {
		if text {
			panic(err)
		}
		return 0
	}

	err = object.PrintProto(proto)
	if err != nil {
		panic(err)
	}

	p := runtime.NewProcess()

	p.Require("", stdlib.Open)

	_, err = p.Exec(proto)
	if err != nil {
		err = object.PrintError(err)
		if err != nil {
			panic(err)
		}

		return 0
	}

	return 1
}
