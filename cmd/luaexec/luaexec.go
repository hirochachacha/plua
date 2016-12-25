package main

import (
	"fmt"
	"os"

	"github.com/hirochachacha/plua/compiler"
	"github.com/hirochachacha/plua/object"
	"github.com/hirochachacha/plua/runtime"
	"github.com/hirochachacha/plua/stdlib"
)

func main() {
	if len(os.Args) < 2 {
		fmt.Fprintln(os.Stderr, "usage: luaexec file")
		os.Exit(2)
	}

	c := compiler.NewCompiler()

	proto, err := c.CompileFile(os.Args[1], compiler.Either)
	if err != nil {
		object.PrintError(err)
		os.Exit(1)
	}

	p := runtime.NewProcess()

	a := p.NewTableSize(len(os.Args)-2, 2)
	for i, arg := range os.Args {
		a.Set(object.Integer(i-1), object.String(arg))
	}

	p.Globals().Set(object.String("arg"), a)

	p.Require("", stdlib.Open)

	_, err = p.Exec(proto)
	if err != nil {
		object.PrintError(err)
		os.Exit(1)
	}
}
