package compiler_pool

import (
	"io"
	"os"
	"strings"
	"sync"

	"github.com/hirochachacha/plua/compiler"
	"github.com/hirochachacha/plua/compiler/codegen"
	"github.com/hirochachacha/plua/compiler/dump"
	"github.com/hirochachacha/plua/compiler/parser"
	"github.com/hirochachacha/plua/compiler/scanner"
	"github.com/hirochachacha/plua/compiler/undump"
	"github.com/hirochachacha/plua/object"
	"github.com/hirochachacha/plua/position"
)

var pool = &sync.Pool{
	New: func() interface{} {
		return new(compiler.Compiler)
	},
}

func CompileFile(path string, typ compiler.FormatType) (*object.Proto, *object.RuntimeError) {
	var r io.Reader

	if len(path) == 0 {
		r = os.Stdin

		path = "=stdin"
	} else {
		f, err := os.Open(path)
		if err != nil {
			return nil, newRuntimeError(err)
		}
		defer f.Close()

		r = f

		path = "@" + path
	}

	c := pool.Get().(*compiler.Compiler)

	p, err := c.Compile(r, path, typ)

	pool.Put(c)

	return p, newRuntimeError(err)
}

func CompileString(s, srcname string, typ compiler.FormatType) (*object.Proto, *object.RuntimeError) {
	c := pool.Get().(*compiler.Compiler)

	p, err := c.Compile(strings.NewReader(s), srcname, typ)

	pool.Put(c)

	return p, newRuntimeError(err)
}

func newRuntimeError(err error) *object.RuntimeError {
	switch err := err.(type) {
	case nil:
		return nil
	case *compiler.Error:
		return object.NewRuntimeError(err.Err.Error())
	case *scanner.Error:
		return &object.RuntimeError{
			Value:     object.String(err.Err.Error()),
			Level:     1,
			Traceback: []position.Position{err.Pos},
		}
	case *parser.Error:
		return &object.RuntimeError{
			Value:     object.String(err.Err.Error()),
			Level:     1,
			Traceback: []position.Position{err.Pos},
		}
	case *dump.Error:
		return object.NewRuntimeError(err.Err.Error())
	case *undump.Error:
		return object.NewRuntimeError(err.Err.Error())
	case *codegen.Error:
		return object.NewRuntimeError(err.Err.Error())
	default:
		return object.NewRuntimeError(err.Error())
	}
}
