package parser_test

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/hirochachacha/plua/compiler/parser"
	"github.com/hirochachacha/plua/compiler/scanner"
)

func TestParse(t *testing.T) {
	matches, err := filepath.Glob("testdata/*.lua")
	if err != nil {
		t.Fatal(err)
	}
	for _, fname := range matches {
		if fname == "testdata/example.lua" {
			continue
		}

		f, err := os.Open(fname)
		if err != nil {
			t.Fatal(err)
		}

		_, err = parser.Parse(scanner.NewScanner(f, "@"+fname, 0), 0)
		if err != nil {
			t.Error(err)
		}
	}
}
