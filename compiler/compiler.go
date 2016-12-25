package compiler

import (
	"bufio"
	"fmt"
	"io"
	"os"

	"github.com/hirochachacha/plua/compiler/codegen"
	"github.com/hirochachacha/plua/compiler/parser"
	"github.com/hirochachacha/plua/compiler/scanner"
	"github.com/hirochachacha/plua/compiler/undump"
	"github.com/hirochachacha/plua/internal/version"
	"github.com/hirochachacha/plua/object"
)

type FormatType uint

const (
	Either FormatType = iota
	Text
	Binary
)

func (typ FormatType) String() string {
	switch typ {
	case Either:
		return "either"
	case Text:
		return "text"
	case Binary:
		return "binary"
	default:
		return "unexpected"
	}
}

type Compiler struct {
	s *scanner.ScanState
	r *bufio.Reader
}

func NewCompiler() *Compiler {
	return new(Compiler)
}

func (c *Compiler) Compile(r io.Reader, srcname string, typ FormatType) (*object.Proto, error) {

	if c.r == nil {
		c.r = bufio.NewReader(r)
	} else {
		c.r.Reset(r)
	}

	b, err := c.r.Peek(1)

	switch {
	case err == io.EOF, err == nil && b[0] != version.LUA_SIGNATURE[0]:
		if typ != Either && typ != Text {
			return nil, &Error{fmt.Errorf("compiler: attempt to load a %s chunk (mode is '%s')", "text", typ)}
		}

		if c.s == nil {
			c.s = scanner.Scan(c.r, srcname, 0)
		} else {
			c.s.Reset(c.r, srcname, 0)
		}

		ast, err := parser.Parse(c.s, 0)
		if err != nil {
			return nil, err
		}

		return codegen.Generate(ast)
	case err == nil && b[0] == version.LUA_SIGNATURE[0]:
		if typ != Either && typ != Binary {
			return nil, &Error{fmt.Errorf("compiler: attempt to load a %s chunk (mode is '%s')", "binary", typ)}
		}

		return undump.Undump(c.r, 0)
	default:
		return nil, err
	}
}

func (c *Compiler) CompileFile(path string, typ FormatType) (*object.Proto, error) {
	f, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer f.Close()

	return c.Compile(f, "@"+f.Name(), typ)
}
