package compiler_pool

import (
	"bytes"
	"io"
	"os"
	"strings"
	"sync"

	"github.com/hirochachacha/plua/compiler"
	"github.com/hirochachacha/plua/compiler/dump"
	"github.com/hirochachacha/plua/object"
)

var pool = &sync.Pool{
	New: func() interface{} {
		return new(compiler.Compiler)
	},
}

func CompileFile(path string, typ compiler.FormatType) (*object.Proto, error) {
	var r io.Reader

	if len(path) == 0 {
		r = os.Stdin

		path = "=stdin"
	} else {
		f, err := os.Open(path)
		if err != nil {
			return nil, err
		}
		defer f.Close()

		r = f

		path = "@" + path
	}

	c := pool.Get().(*compiler.Compiler)

	p, err := c.Compile(r, path, typ)

	pool.Put(c)

	return p, err
}

func CompileString(s, srcname string, typ compiler.FormatType) (*object.Proto, error) {
	c := pool.Get().(*compiler.Compiler)

	p, err := c.Compile(strings.NewReader(s), srcname, typ)

	pool.Put(c)

	return p, err
}

func DumpToString(p *object.Proto, strip bool) (string, error) {
	var mode dump.Mode
	if strip {
		mode |= dump.StripDebugInfo
	}

	buf := new(bytes.Buffer)

	err := dump.DumpTo(buf, p, mode)
	if err != nil {
		return "", err
	}

	return buf.String(), nil
}
