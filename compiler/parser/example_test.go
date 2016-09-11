package parser_test

import (
	"os"

	"github.com/hirochachacha/plua/compiler/ast/printer"
	"github.com/hirochachacha/plua/compiler/parser"
	"github.com/hirochachacha/plua/compiler/scanner"
)

func ExampleHello() {
	f, err := os.Open("testdata/example.lua")
	if err != nil {
		panic(err)
	}

	ast, err := parser.Parse(scanner.NewScanner(f, "@hello.lua", 0), 0)
	if err != nil {
		panic(err)
	}

	printer.PrintTree(ast)

	// Output:
	// File { 1:1--
	//   Shebang: "#!/usr/bin/env lua"
	//   Chunk: {
	//     ExprStmt { 3:1--
	//       X: CallExpr { 3:1--
	//         X: Name { 3:1-3:6
	//           NamePos: 3:1
	//           Name: print
	//         }
	//         Colon: -
	//         Name: nil
	//         Lparen: -
	//         Args: {
	//           BasicLit { 3:7-3:21
	//             Token.Type: STRING
	//             Token.Pos: 3:7
	//             Token.Lit: "Hello World!"
	//           }
	//         }
	//         Rparen: -
	//       }
	//     }
	//   }
	//   Comments: {
	//   }
	// }
}
