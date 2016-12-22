package printer_test

import (
	"os"

	"github.com/hirochachacha/plua/compiler/ast/printer"
	"github.com/hirochachacha/plua/compiler/parser"
)

func ExampleFprint() {
	ast, err := parser.ParseFile("testdata/example.lua", parser.ParseComments)
	if err != nil {
		panic(err)
	}

	printer.Fprint(os.Stdout, ast)

	// Output:
	// if true then -- comment1
	//   print("x")    -- comment2
	//   print("yyyy") -- comment3
	// elseif true then -- comment4
	//   print("y")        -- comment5
	//   print("xxxxxxxx") -- comment6
	// else -- comment7
	//   -- comment8
	//   -- comment9
	//   print("xxxx") -- comment 10
	//   print("xx")   -- comment 11
	// end
	//
	// -- fake
	// -- art
	// --[[
	//  fake
	//  man
	//  ]]
	//
	// x = 1 + -- bar
	//   4 + 9 -- foo
	//   + 10  -- baz
	//
	// x = x + 5*9 - 10/5 - (-5)
	// x = x + (5-1+9)*6
	//
	// if (x ~= 10) == true then
	//   print(x)
	// end
	//
	// print("hello" .. "world")
	// print("hello".."world", "hello".."world")
	// print(1 + 7)
	// print(1+7, 1+8)
	//
	// y = foo[x+5] + 8
	//
	// t = {[5+9] = y, 11}
	//
	// local foo
	//
	// x = 10; y = 24
	//
	// t = { --xxx
	//   1,        -- foo
	//   81111111, -- bar
	// } -- aaa
	//
	// t =
	//   {
	//     11111,
	//     [5] = 011,
	//   }
	//
	// t = {
	//   x = {
	//     y =
	//       {
	//         z = 9,
	//       },
	//   },
	// }
	//
	// x {
	// }
	//
	// print('\'test')
	//
	// return (x) + 1
}
