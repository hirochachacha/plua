package codegen

import (
	"os"
	"testing"

	"github.com/hirochachacha/plua/compiler/parser"
	"github.com/hirochachacha/plua/compiler/scanner"
	"github.com/hirochachacha/plua/object/printer"
)

func TestMain(t *testing.T) {
	f, err := os.Open("testdata/test.lua")
	if err != nil {
		panic(err)
	}

	ast, err := parser.Parse(scanner.NewScanner(f, "testdata/test.lua", 0), 0)
	if err != nil {
		panic(err)
	}

	proto, err := Generate(ast)
	if err != nil {
		panic(err)
	}

	printer.Print(proto)
}
