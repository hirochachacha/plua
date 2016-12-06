package base_test

import (
	"path/filepath"
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
		`runtime: test_error:3: test_message`,
	},
	{
		`
		function x()
		  error("test_message")
		end

		x()
		`,
		nil,
		`runtime: test_error:3: test_message`,
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
		`return dofile("testdata/do/do.lua")`,
		[]object.Value{object.True},
		"",
	},
	{
		`assert(dofile("testdata/do/notexist"))`,
		nil,
		"no such file",
	},
	{
		`assert(dofile("testdata/do/not.lua"))`,
		nil,
		"expected",
	},
	{
		`dofile("testdata/do/do_err.lua")`,
		nil,
		"true",
	},
}

func TestDoFile(t *testing.T) {
	testExecCases(t, "test_dofile", testDoFiles)
}

func testExecCases(t *testing.T, testname string, tests []execCase) {
	c := compiler.NewCompiler()

	for _, test := range tests {
		proto, err := c.Compile(strings.NewReader(test.Code), "="+testname, 0)
		if err != nil {
			t.Fatal(err)
		}

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

func TestBase(t *testing.T) {
	c := compiler.NewCompiler()

	matches, err := filepath.Glob("testdata/*.lua")
	if err != nil {
		t.Fatal(err)
	}

	for _, fname := range matches {
		proto, err := c.CompileFile(fname, 0)
		if err != nil {
			t.Fatal(err)
		}

		p := runtime.NewProcess()

		p.Require("_G", base.Open)

		_, err = p.Exec(proto)
		if err != nil {
			t.Error(err)
		}
	}
}
