plua
====

[![GoDoc](https://godoc.org/github.com/hirochachacha/plua.svg?status.svg)](http://godoc.org/github.com/hirochachacha/plua)
[![Build Status](https://travis-ci.org/hirochachacha/plua.svg?branch=master)](https://travis-ci.org/hirochachacha/plua)
[![Code Climate](https://codeclimate.com/github/hirochachacha/plua/badges/gpa.svg)](https://codeclimate.com/github/hirochachacha/plua)
[![Test Coverage](https://codeclimate.com/github/hirochachacha/plua/badges/coverage.svg)](https://codeclimate.com/github/hirochachacha/plua/coverage)

Description
-----------

Lua 5.3 implementation.

```go
package main

import (
	"fmt"
	"strings"

	"github.com/hirochachacha/plua/compiler"
	"github.com/hirochachacha/plua/runtime"
	"github.com/hirochachacha/plua/stdlib"
)

var input = `
-- example code is taken from https://tour.golang.org/concurrency/5

function fibonacci(ch, quit)
  local x, y = 0, 1
  while true do
    local chosen, recv, recvOK = goroutine.select(
      goroutine.case("send", ch, x),
      goroutine.case("recv", quit)
    )

    if chosen == 1 then
      x, y = y, x+y
    elseif chosen == 2 then
	  print("quit")
      return
    end
  end
end

ch = goroutine.newchannel()
quit = goroutine.newchannel()

goroutine.wrap(function()
  for i = 1, 10 do
    print(ch:recv())
  end
  quit:send(nil)
end)()

fibonacci(ch, quit)

return "ok"
`

func main() {
	c := compiler.NewCompiler()

	proto, err := c.Compile(strings.NewReader(input), "=input.lua", compiler.Text)
	if err != nil {
		panic(err)
	}

	p := runtime.NewProcess()

	p.Require("", stdlib.Open)

	rets, err := p.Exec(proto)
	if err != nil {
		// object.PrintError(err) // print traceback from error
		panic(err)
	}

	fmt.Println(rets[0])
}
```
Output:
```
0	true
1	true
1	true
2	true
3	true
5	true
8	true
13	true
21	true
34	true
quit
ok
```
