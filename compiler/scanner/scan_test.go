package scanner_test

import (
	"fmt"
	"os"
	"path/filepath"
	"testing"

	"github.com/hirochachacha/plua/compiler/scanner"
	"github.com/hirochachacha/plua/compiler/token"
)

func TestScan(t *testing.T) {
	matches, err := filepath.Glob("testdata/*.lua")
	if err != nil {
		t.Fatal(err)
	}
	for _, fname := range matches {
		if fname == "testdata/example.lua" {
			continue
		}

		fmt.Println("filename:", fname)

		f, err := os.Open(fname)
		if err != nil {
			t.Fatal(err)
		}
		defer f.Close()

		s := scanner.NewScanner(f, "@"+fname, 0)
		for {
			tok := s.Scan()
			if tok.Type == token.EOF {
				break
			}
			fmt.Printf("line: %d, column: %d, tok: %s, lit: %s\n", tok.Pos.Line, tok.Pos.Column, tok.Type, tok.Lit)
		}
		if err := s.Err(); err != nil {
			t.Fatal(err)
		}

		fmt.Println()
	}
}
