package runtime_test

import (
	"strings"
	"testing"

	"github.com/hirochachacha/plua/compiler"
	"github.com/hirochachacha/plua/object"
	"github.com/hirochachacha/plua/runtime"
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
}

func TestExec(t *testing.T) {
	c := compiler.NewCompiler()

	for _, test := range testExec {
		proto, err := c.Compile(strings.NewReader(test.Code), "=testCode")
		if err != nil {
			t.Fatal(err)
		}

		p := runtime.NewProcess()

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

func Error(th object.Thread, args ...object.Value) (rets []object.Value, err object.Value) {
	return nil, args[0]
}

func PCall(th object.Thread, args ...object.Value) (rets []object.Value, err object.Value) {
	rets, ok := th.PCall(args[0], nil, args[1:]...)

	return append([]object.Value{object.Boolean(ok)}, rets...), nil
}

var testExecError = []struct {
	Code string

	ErrValue object.Value
}{
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

		g := p.Globals()
		g.Set(object.String("error"), object.GoFunction(Error))
		g.Set(object.String("pcall"), object.GoFunction(PCall))

		_, err = p.Exec(proto)
		if err == nil {
			t.Fatal("expected err, got nil")
		}
		rerr, ok := err.(*runtime.Error)
		if !ok {
			t.Fatalf("expected *Error, got %T: %v", err, err)
		}

		if !object.Equal(rerr.Value, test.ErrValue) {
			t.Errorf("code: %s: expected %v, got %v", test.Code, test.ErrValue, rerr.Value)
		}
	}
}
