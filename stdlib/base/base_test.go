package base_test

import (
	"reflect"
	"strings"
	"testing"

	"github.com/hirochachacha/plua/compiler"
	"github.com/hirochachacha/plua/object"
	"github.com/hirochachacha/plua/runtime"
	"github.com/hirochachacha/plua/stdlib/base"
)

type execCase struct {
	Code      string
	Rets      []object.Value
	ErrString string
}

var testAsserts = []execCase{
	{
		`return assert("hello", "world")`,
		[]object.Value{object.String("hello"), object.String("world")},
		"",
	},
	{
		`assert(nil)`,
		nil,
		"assertion failed!",
	},
	{
		`assert(false)`,
		nil,
		"assertion failed!",
	},
	{
		`assert(false, "hello")`,
		nil,
		"hello",
	},
	{
		`assert()`,
		nil,
		"got no value",
	},
	{
		`function x() assert() end; x()`,
		nil,
		"got no value",
	},
	{
		`t = {x = assert}; t.x()`,
		nil,
		"got no value",
	},

	{
		`assert()`,
		nil,
		"bad argument #1 to 'assert'",
	},
	{
		`function x() assert() end; x()`,
		nil,
		"bad argument #1 to 'assert'",
	},
	{
		`x = assert; x()`,
		nil,
		"bad argument #1 to 'x'",
	},
	{
		`t = {x = assert}; t.x()`,
		nil,
		"bad argument #1 to 'x'",
	},
}

func TestAssert(t *testing.T) {
	testExecCases(t, "test_assert", testAsserts)
}

var testErrors = []execCase{
	{
		`

		error("test_message")
		`,
		nil,
		`runtime: "test_message" from test_error:3`,
	},
	{
		`
		function x()
		  error("test_message")
		end

		x()
		`,
		nil,
		`runtime: "test_message" from test_error:3 via test_error:6`,
	},
}

func TestError(t *testing.T) {
	testExecCases(t, "test_error", testErrors)
}

var testCollectGarbages = []execCase{
	{
		`return collectgarbage()`,
		[]object.Value{object.Integer(0)},
		"",
	},
	{
		`return collectgarbage("collect")`,
		[]object.Value{object.Integer(0)},
		"",
	},
	{
		`n = collectgarbage("count"); return type(n)`,
		[]object.Value{object.String("number")},
		"",
	},
	{
		`return collectgarbage("isrunning")`,
		[]object.Value{object.True},
		"",
	},
	{
		`collectgarbage("stop")`,
		nil,
		"not implemented",
	},
	{
		`collectgarbage("restart")`,
		nil,
		"not implemented",
	},
	{
		`collectgarbage("step")`,
		nil,
		"not implemented",
	},
	{
		`collectgarbage("setpause")`,
		nil,
		"not implemented",
	},
	{
		`collectgarbage("setstepmul")`,
		nil,
		"not implemented",
	},
	{
		`collectgarbage("testtesttest")`,
		nil,
		"invalid option",
	},
}

func TestCollectGarbage(t *testing.T) {
	testExecCases(t, "test_collectgarbage", testCollectGarbages)
}

var testDoFiles = []execCase{
	{
		`return dofile("testdata/do.lua")`,
		[]object.Value{object.True},
		"",
	},
	{
		`assert(dofile("testdata/notexist"))`,
		nil,
		"no such file",
	},
	{
		`assert(dofile("testdata/not.lua"))`,
		nil,
		"compiler/parser",
	},
	{
		`dofile("testdata/do_err.lua")`,
		nil,
		"true",
	},
}

func TestDoFile(t *testing.T) {
	testExecCases(t, "test_dofile", testDoFiles)
}

var testSelects = []execCase{
	{
		`return select(1, 2, 3)`,
		[]object.Value{object.Integer(2), object.Integer(3)},
		"",
	},
	{
		`return select(-1, 2, 3)`,
		[]object.Value{object.Integer(3)},
		"",
	},
	{
		`return select("#", 2, 3)`,
		[]object.Value{object.Integer(2)},
		"",
	},
}

func TestSelect(t *testing.T) {
	testExecCases(t, "test_select", testSelects)
}

var testTypes = []execCase{
	{
		`return type(nil)`,
		[]object.Value{object.String("nil")},
		"",
	},
	{
		`return type(5)`,
		[]object.Value{object.String("number")},
		"",
	},
	{
		`return type("hello")`,
		[]object.Value{object.String("string")},
		"",
	},
	{
		`return type(false)`,
		[]object.Value{object.String("boolean")},
		"",
	},
	{
		`return type({1 = 5, n = 9})`,
		[]object.Value{object.String("table")},
		"",
	},
	{
		`return type(print)`,
		[]object.Value{object.String("function")},
		"",
	},
}

func TestType(t *testing.T) {
	testExecCases(t, "test_type", testTypes)
}

func testExecCases(t *testing.T, testname string, tests []execCase) {
	c := compiler.NewCompiler()

	for _, test := range tests {
		proto, err := c.Compile(strings.NewReader(test.Code), "="+testname)

		p := runtime.NewProcess()

		p.Require("_G", base.Open)

		rets, err := p.Exec(proto)
		if err != nil {
			if test.ErrString == "" {
				t.Fatal(err)
			}
			if !strings.Contains(err.Error(), test.ErrString) {
				t.Errorf("code: `%s`, err: %v, expected: %v\n", test.Code, err, test.ErrString)
			}
		}
		if !reflect.DeepEqual(rets, test.Rets) {
			t.Errorf("code: `%s`, rets: %v, expected: %v\n", test.Code, rets, test.Rets)
		}
	}
}
