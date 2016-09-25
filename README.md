plua
====

[![Build Status](https://travis-ci.org/hirochachacha/plua.svg?branch=master)](https://travis-ci.org/hirochachacha/plua)

Description
-----------

Lua 5.3 implementation. (WIP)

```go
package main

import (
	"fmt"
	"strings"

	"github.com/hirochachacha/plua/compiler"
	"github.com/hirochachacha/plua/runtime"
)

func main() {
	c := compiler.NewCompiler()

	proto, err := c.Compile(strings.NewReader(`
function fib(n)
  if n == 0 then
    return 0
  elseif n == 1 then
    return 1
  end
  return fib(n-1) + fib(n-2)
end

return fib(10)
	`), "=fib", compiler.Text)
	if err != nil {
		panic(err)
	}

	p := runtime.NewProcess()

	rets, err := p.Exec(proto)
	if err != nil {
		panic(err)
	}

	fmt.Println(rets)
}
```
