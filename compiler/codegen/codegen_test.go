package codegen

import (
	"os"
	"testing"

	"github.com/hirochachacha/blua/compiler/parser"
	"github.com/hirochachacha/blua/compiler/scanner"
	"github.com/hirochachacha/blua/object/printer"
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

	proto := Generate(ast)

	printer.Print(proto)
}
