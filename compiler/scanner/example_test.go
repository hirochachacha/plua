package scanner_test

import (
	"fmt"
	"os"

	"github.com/hirochachacha/plua/compiler/scanner"
	"github.com/hirochachacha/plua/compiler/token"
)

func ExampleScan() {
	f, err := os.Open("testdata/example.lua")
	if err != nil {
		panic(err)
	}

	s := scanner.Scan(f, "@"+"testdata/example.lua", 0)

	for {
		tok := s.Next()
		if tok.Type == token.EOF {
			break
		}
		fmt.Printf("line: %d, column: %d, tok: %s, lit: %s\n", tok.Pos.Line, tok.Pos.Column, tok.Type, tok.Lit)
	}
	if err := s.Err(); err != nil {
		panic(err)
	}

	// Output:
	// line: 3, column: 1, tok: NAME, lit: print
	// line: 3, column: 7, tok: STRING, lit: "Hello World"
}
