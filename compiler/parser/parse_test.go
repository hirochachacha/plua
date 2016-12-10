package parser_test

import (
	"path/filepath"
	"testing"

	"github.com/hirochachacha/plua/compiler/parser"
)

func TestParseFile(t *testing.T) {
	matches, err := filepath.Glob("testdata/*.lua")
	if err != nil {
		t.Fatal(err)
	}
	for _, fname := range matches {
		if fname == "testdata/example.lua" {
			continue
		}

		ast, err := parser.ParseFile(fname, 0)
		if err != nil {
			t.Error(err)
		}

		_ = ast

		// printer.PrintTree(ast)
	}
}
