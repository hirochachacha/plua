package compiler

import (
	"bufio"
	"fmt"
	"io"

	"github.com/hirochachacha/plua/compiler/codegen"
	"github.com/hirochachacha/plua/compiler/parser"
	"github.com/hirochachacha/plua/compiler/scanner"
	"github.com/hirochachacha/plua/compiler/undump"
	"github.com/hirochachacha/plua/internal/version"
	"github.com/hirochachacha/plua/object"
)

type readerAt interface {
	io.Reader
	io.ReaderAt
}

type Compiler struct {
	s *scanner.Scanner
	u *bufio.Reader // buffer for undump
	r readerAt
}

func NewCompiler() *Compiler {
	return new(Compiler)
}

func (c *Compiler) Compile(r io.Reader, source string) (*object.Proto, error) {
	if r, ok := r.(readerAt); ok {
		c.r = r
	} else {
		c.r = &onceReadAt{r: r}
	}

	bs := make([]byte, 1)

	_, err := c.r.ReadAt(bs, 0)
	if err != nil {
		if err == io.EOF {
			if c.s == nil {
				c.s = scanner.NewScanner(c.r, source, 0)
			} else {
				c.s.Reset(c.r, source, 0)
			}
			ast, err := parser.Parse(c.s, 0)
			if err != nil {
				return nil, err
			}

			return codegen.Generate(ast), nil
		}

		return nil, err
	}

	// is bytecode?
	if bs[0] == version.LUA_SIGNATURE[0] {
		if c.u == nil {
			c.u = bufio.NewReader(c.r)
		} else {
			c.u.Reset(c.r)
		}
		return undump.Undump(c.u, 0)
	}

	if c.s == nil {
		c.s = scanner.NewScanner(c.r, source, 0)
	} else {
		c.s.Reset(c.r, source, 0)
	}

	ast, err := parser.Parse(c.s, 0)
	if err != nil {
		return nil, err
	}

	return codegen.Generate(ast), nil
}

func (c *Compiler) CompileText(r io.Reader, source string) (*object.Proto, error) {
	if r, ok := r.(readerAt); ok {
		c.r = r
	} else {
		c.r = &onceReadAt{r: r}
	}

	bs := make([]byte, 1)

	_, err := c.r.ReadAt(bs, 0)
	if err != nil {
		if err == io.EOF {
			if c.s == nil {
				c.s = scanner.NewScanner(c.r, source, 0)
			} else {
				c.s.Reset(c.r, source, 0)
			}
			ast, err := parser.Parse(c.s, 0)
			if err != nil {
				return nil, err
			}

			return codegen.Generate(ast), nil
		}

		return nil, err
	}

	// is bytecode?
	if bs[0] == version.LUA_SIGNATURE[0] {
		return nil, fmt.Errorf("compiler: attempt to load a %s chunk (mode is '%s')", "binary", "text")
	}

	if c.s == nil {
		c.s = scanner.NewScanner(c.r, source, 0)
	} else {
		c.s.Reset(c.r, source, 0)
	}

	ast, err := parser.Parse(c.s, 0)
	if err != nil {
		return nil, err
	}

	return codegen.Generate(ast), nil
}

func (c *Compiler) CompileBinary(r io.Reader) (*object.Proto, error) {
	if r, ok := r.(readerAt); ok {
		c.r = r
	} else {
		c.r = &onceReadAt{r: r}
	}

	bs := make([]byte, 1)

	_, err := c.r.ReadAt(bs, 0)
	if err != nil {
		if err == io.EOF {
			return nil, fmt.Errorf("compiler: attempt to load a %s chunk (mode is '%s')", "text", "binary")
		}

		return nil, err
	}

	// is text?
	if bs[0] != version.LUA_SIGNATURE[0] {
		return nil, fmt.Errorf("compiler: attempt to load a %s chunk (mode is '%s')", "text", "binary")
	}

	if c.u == nil {
		c.u = bufio.NewReader(c.r)
	} else {
		c.u.Reset(c.r)
	}
	return undump.Undump(c.u, 0)
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
