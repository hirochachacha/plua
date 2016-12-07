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
	{`function x() return 1 end; return x()`, []object.Value{object.Integer(1)}},
	{`function x() return 1 end; x(); return 2`, []object.Value{object.Integer(2)}},
	{`local a = {1 = 10, 2, 3, 4 = 9}; return a[4]`, []object.Value{object.Integer(9)}},
	{`a = {1 = 10, 2, 3, 4 = 9}; return #a`, []object.Value{object.Integer(2)}},
	{`return #"test"`, []object.Value{object.Integer(4)}},
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
	{
		`
		local x = {}
		for k, v in pairs({2, 2, 3, 4}) do
			x[k] = v
		end
		return x[1], x[2], x[3], x[4]
		`,
		[]object.Value{object.Integer(2), object.Integer(2), object.Integer(3), object.Integer(4)},
	},
	{
		`
		local x = {}
		for k, v in ipairs({1, 2, 3, 4}) do
			x[k] = v
		end
		return x[1], x[2], x[3], x[4]
		`,
		[]object.Value{object.Integer(1), object.Integer(2), object.Integer(3), object.Integer(4)},
	},
	{
		`
		local x = {}
		(function()
			for k, v in ipairs({1, 2, 3, 4}) do
				x[k] = v
			end
		end)()
		return x[1], x[2], x[3], x[4]
		`,
		[]object.Value{object.Integer(1), object.Integer(2), object.Integer(3), object.Integer(4)},
	},
	{
		`
		local x = {}
		i = 1
		assert(math.type(i) == "integer")
		x[i] = 5
		i = i + 0.0
		assert(math.type(i) == "float")
		return x[1]
		`,
		[]object.Value{object.Integer(5)},
	},
	{
		`
		return #{nil, nil}
		`,
		[]object.Value{object.Integer(0)},
	},
	{
		`
		return #{1, nil, 3, nil}
		`,
		[]object.Value{object.Integer(3)},
	},
	{
		`
		t = {1, nil, 3, nil}
		t[3] = nil
		return #t
		`,
		[]object.Value{object.Integer(1)},
	},
	{
		`
		function f()
		  return 2, 3
		end
		t = {1, f()}
		return #t
		`,
		[]object.Value{object.Integer(3)},
	},
	{
		`
		function f()
		  return 2, 3
		end
		t = {1, f(), nil}
		return #t
		`,
		[]object.Value{object.Integer(2)},
	},
}

func TestExec(t *testing.T) {
	c := compiler.NewCompiler()

	for i, test := range testExec {
		proto, err := c.Compile(strings.NewReader(test.Code), "=test_code", 0)
		if err != nil {
			t.Fatalf("%d: %v", i+1, err)
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

	for i, test := range testExecError {
		proto, err := c.Compile(strings.NewReader(test.Code), "=test_code", 0)
		if err != nil {
			t.Fatalf("%d: %v", i+1, err)
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
