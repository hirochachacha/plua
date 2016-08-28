package runtime_test

import (
	"strings"
	"testing"

	"github.com/hirochachacha/blua/compiler"
	"github.com/hirochachacha/blua/object"
	"github.com/hirochachacha/blua/runtime"
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
					t.Errorf("expected %v, got %v", test.Rets[i], rets[i])
				}
			}
		}
	}
}
