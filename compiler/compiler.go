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

type readerAt interface {
	io.Reader
	io.ReaderAt
}

type Compiler struct {
	s *scanner.Scanner
	u *bufio.Reader // buffer for undump
	r readerAt
	b [1]byte
}

func NewCompiler() *Compiler {
	return new(Compiler)
}

func (c *Compiler) Compile(r io.Reader, srcname string, typ FormatType) (*object.Proto, error) {
	if r, ok := r.(readerAt); ok {
		c.r = r
	} else {
		c.r = &onceReadAt{r: r}
	}

	_, err := c.r.ReadAt(c.b[:], 0)

	switch {
	case err != nil && err != io.EOF:
		fallthrough
	case err == nil && c.b[0] != version.LUA_SIGNATURE[0]:
		if typ != Either && typ != Text {
			return nil, fmt.Errorf("compiler: attempt to load a %s chunk (mode is '%s')", "text", typ)
		}

		if c.s == nil {
			c.s = scanner.NewScanner(c.r, srcname, 0)
		} else {
			c.s.Reset(c.r, srcname, 0)
		}

		ast, err := parser.Parse(c.s, 0)
		if err != nil {
			return nil, err
		}

		return codegen.Generate(ast), nil
	case err == nil && c.b[0] == version.LUA_SIGNATURE[0]:
		if typ != Either && typ != Binary {
			return nil, fmt.Errorf("compiler: attempt to load a %s chunk (mode is '%s')", "binary", typ)
		}

		if c.u == nil {
			c.u = bufio.NewReader(c.r)
		} else {
			c.u.Reset(c.r)
		}

		return undump.Undump(c.u, 0)
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

type onceReadAt struct {
	r   io.Reader
	buf []byte
}

func (r *onceReadAt) ReadAt(p []byte, off int64) (n int, err error) {
	if r.buf != nil {
		panic("ReadAt can be called only once")
	}

	n, err = io.ReadFull(r.r, p)
	if err != nil {
		if err == io.ErrUnexpectedEOF {
			err = io.EOF
		}
		return
	}

	r.buf = p

	return
}

func (r *onceReadAt) Read(p []byte) (n int, err error) {
	if len(r.buf) > 0 && len(p) > 0 {
		n1 := copy(p, r.buf)

		r.buf = r.buf[n1:]

		n, err = r.r.Read(p[n1:])

		return n + n1, err
	}

	return r.r.Read(p)
}
