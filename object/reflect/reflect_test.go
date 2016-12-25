package reflect_test

import (
	"strings"
	"testing"

	"github.com/hirochachacha/plua/compiler"
	"github.com/hirochachacha/plua/object"
	"github.com/hirochachacha/plua/object/reflect"
	"github.com/hirochachacha/plua/runtime"
	"github.com/hirochachacha/plua/stdlib"
)

type myint int

func (i myint) Hello() string { return "hello" }

type myArray [4]int

func (m *myArray) Hello() string { return "hello" }

var testCases = []struct {
	Map  map[string]interface{}
	Code string
}{
	// builtins
	{
		map[string]interface{}{"i": 10},
		`assert(type(i) == "number" and i == 10)`,
	},
	{
		map[string]interface{}{"i": int8(10)},
		`assert(type(i) == "number" and i == 10)`,
	},
	{
		map[string]interface{}{"i": int32(10)},
		`assert(type(i) == "number" and i == 10)`,
	},
	{
		map[string]interface{}{"b": false},
		`assert(type(b) == "boolean" and b == false)`,
	},
	{
		map[string]interface{}{"s": "test"},
		`assert(type(s) == "string" and s == "test")`,
	},

	// uint
	{
		map[string]interface{}{
			"u1": uint(10),
			"u2": uint(10),
			"u3": uint(15),
		},
		`
		assert(type(u1) == "userdata")
		-- assert(u1 == 10)
		assert(u1 == u2)
		assert(u1+5 == u3)
		`,
	},

	// array
	{
		map[string]interface{}{
			"a1": [5]int{1, 2, 3, 4, 5},
			"a2": [5]int{1, 2, 3, 4, 5},
		},
		`
		assert(type(a1) == "userdata")
		assert(a1 == a2)
		assert(a1[1] == 1)
		assert(a1["test"] == nil)
		`,
	},

	// slice
	{
		map[string]interface{}{
			"a": []int{1, 2, 3, 4, 5},
		},
		`
		assert(type(a) == "userdata")
		assert(a[1] == 1)
		assert(a["test"] == nil)
		a[1] = 6
		assert(a[1] == 6)
		`,
	},

	// alias types
	{
		map[string]interface{}{
			"i": myint(10),
		},
		`
		assert(type(i) == "userdata")
		assert(i:Hello() == "hello")
		`,
	},
	{
		map[string]interface{}{
			"a": myArray{1, 2, 3},
		},
		`
		assert(type(a) == "userdata")
		assert(a:Hello() == "hello")
		`,
	},
}

func TestReflect(t *testing.T) {
	c := compiler.NewCompiler()

	for i, test := range testCases {
		proto, err := c.Compile(strings.NewReader(test.Code), "=test_code", 0)
		if err != nil {
			t.Fatalf("%d: %v", i+1, err)
		}

		p := runtime.NewProcess()

		p.Require("", stdlib.Open)

		g := p.Globals()
		for k, v := range test.Map {
			g.Set(object.String(k), reflect.ValueOf(v))
		}

		_, err = p.Exec(proto)
		if err != nil {
			t.Errorf("%d: %v", i+1, err)
		}
	}
}
