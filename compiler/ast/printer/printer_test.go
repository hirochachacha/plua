package printer

import (
	"os"
	"testing"

	"github.com/hirochachacha/plua/compiler/parser"
	"github.com/hirochachacha/plua/compiler/scanner"
)

func TestMain(t *testing.T) {
	f, err := os.Open("hello.lua")
	if err != nil {
		panic(err)
	}

	ast, err := parser.Parse(scanner.NewScanner(f, "hello.lua", scanner.ScanComments), 0)
	if err != nil {
		panic(err)
	}

	Print(ast)
}
