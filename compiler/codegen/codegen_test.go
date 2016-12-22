package codegen_test

import (
	"testing"

	"github.com/hirochachacha/plua/compiler/codegen"
	"github.com/hirochachacha/plua/compiler/parser"
	"github.com/hirochachacha/plua/object"
)

func TestGenerate(t *testing.T) {
	ast, err := parser.ParseFile("testdata/test.lua", 0)
	if err != nil {
		panic(err)
	}

	proto, err := codegen.Generate(ast)
	if err != nil {
		panic(err)
	}

	object.PrintProto(proto)
}
