package runtime_test

import (
	"strings"
	"testing"

	"github.com/hirochachacha/plua/compiler"
	"github.com/hirochachacha/plua/object"
	"github.com/hirochachacha/plua/runtime"
	"github.com/hirochachacha/plua/stdlib"
)

var testExec = []struct {
	Code string
	Rets []object.Value
}{
	{`return "hello world"`, []object.Value{object.String("hello world")}},
	{`return 1`, []object.Value{object.Integer(1)}},
	{`return 1, 2, 3`, []object.Value{object.Integer(1), object.Integer(2), object.Integer(3)}},
	{`return 3.14`, []object.Value{object.Number(3.14)}},
	{`return false`, []object.Value{object.Boolean(false)}},
	{`return`, nil},
	{`return; return 1`, nil},
	{`function x() return 1 end; return x()`, []object.Value{object.Integer(1)}},
	{`function x() return 1 end; x(); return 2`, []object.Value{object.Integer(2)}},
	{`local a = {1 = 10, 2, 3, 4 = 9}; return a[4]`, []object.Value{object.Integer(9)}},
	{`a = {1 = 10, 2, 3, 4 = 9}; return #a`, []object.Value{object.Integer(2)}},
	{`
	function fib(n)
	  if n == 0 then
	  	return 0
	  elseif n == 1 then
	  	return 1
	  end
	  return fib(n-1) + fib(n-2)
	end
	return fib(10)
	`, []object.Value{object.Integer(55)}},
	{`return pcall(debug.getinfo, print, "X")`, []object.Value{object.False, object.String("bad argument #2 to 'debug.getinfo' (invalid option 'X')")}},
}

func TestExec(t *testing.T) {
	c := compiler.NewCompiler()

	for _, test := range testExec {
		proto, err := c.Compile(strings.NewReader(test.Code), "=testCode")
		if err != nil {
			t.Fatal(err)
		}

		p := runtime.NewProcess()

		p.Require("", stdlib.Open)

		rets, err := p.Exec(proto)
		if err != nil {
			t.Fatal(err)
		}

		if len(rets) != len(test.Rets) {
			t.Errorf("expected %v, got %v", test.Rets, rets)
		} else {
			for i := range rets {
				if !object.Equal(rets[i], test.Rets[i]) {
					t.Errorf("code: %s, expected %v, got %v", test.Code, test.Rets[i], rets[i])
				}
			}
		}
	}
}

var testExecError = []struct {
	Code string

	ErrValue object.Value
}{
	{`error(nil)`, nil},
	{`error("error")`, object.String("error")},
	{`error(1); error(2)`, object.Integer(1)},
	{`function x() error(1) end; return x()`, object.Integer(1)},
	{`function x() error(1) end; pcall(x); error(2)`, object.Integer(2)},
}

func TestExecError(t *testing.T) {
	c := compiler.NewCompiler()

	for _, test := range testExecError {
		proto, err := c.Compile(strings.NewReader(test.Code), "=testCode")
		if err != nil {
			t.Fatal(err)
		}

		p := runtime.NewProcess()

		p.Require("", stdlib.Open)

		_, err = p.Exec(proto)
		if err == nil {
			t.Fatal("expected err, got nil")
		}
		oerr, ok := err.(*object.RuntimeError)
		if !ok {
			t.Fatalf("expected *object.Error, got %T: %v", err, err)
		}

		if !object.Equal(oerr.Value, test.ErrValue) {
			t.Errorf("code: %s: expected %v, got %v", test.Code, test.ErrValue, oerr.Value)
		}
	}
}
