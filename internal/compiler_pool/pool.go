package compiler_pool

import (
	"bytes"
	"io"
	"os"
	"strings"
	"sync"

	"github.com/hirochachacha/plua/compiler"
	"github.com/hirochachacha/plua/object"
)

var pool = &sync.Pool{
	New: func() interface{} {
		return new(compiler.Compiler)
	},
}

func CompileFile(fname string) (*object.Proto, error) {
	var r io.Reader

	if len(fname) == 0 {
		r = os.Stdin

		fname = "=stdin"
	} else {
		f, err := os.Open(fname)
		if err != nil {
			return nil, err
		}
		r = f

		fname = "@" + fname
	}

	c := pool.Get().(*compiler.Compiler)

	p, err := c.Compile(r, fname)

	pool.Put(c)

	return p, err
}

func CompileString(s, source string) (*object.Proto, error) {
	c := pool.Get().(*compiler.Compiler)

	p, err := c.Compile(strings.NewReader(s), source)

	pool.Put(c)

	return p, err
}

func CompileTextFile(fname string) (*object.Proto, error) {
	var r io.Reader

	if len(fname) == 0 {
		r = os.Stdin

		fname = "=stdin"
	} else {
		f, err := os.Open(fname)
		if err != nil {
			return nil, err
		}
		r = f
		fname = "@" + fname
	}

	c := pool.Get().(*compiler.Compiler)

	p, err := c.CompileText(r, fname)

	pool.Put(c)

	return p, err
}

func CompileTextString(s, source string) (*object.Proto, error) {
	c := pool.Get().(*compiler.Compiler)

	p, err := c.CompileText(strings.NewReader(s), source)

	pool.Put(c)

	return p, err
}

func CompileBinaryFile(fname string) (*object.Proto, error) {
	var r io.Reader

	if len(fname) == 0 {
		r = os.Stdin
	} else {
		f, err := os.Open(fname)
		if err != nil {
			return nil, err
		}
		r = f
	}

	c := pool.Get().(*compiler.Compiler)

	p, err := c.CompileBinary(r)

	pool.Put(c)

	return p, err
}

func CompileBinaryString(s string) (*object.Proto, error) {
	c := pool.Get().(*compiler.Compiler)

	p, err := c.CompileBinary(strings.NewReader(s))

	pool.Put(c)

	return p, err
}

func DumpToString(p *object.Proto, strip bool) (string, error) {
	c := pool.Get().(*compiler.Compiler)

	if strip {
		c.SetMode(compiler.StripDebugInfo)
	}

	buf := new(bytes.Buffer)

	err := c.DumpTo(buf, p)
	if err != nil {
		return "", err
	}

	return buf.String(), nil
}
