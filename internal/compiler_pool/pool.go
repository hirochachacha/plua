package compiler_pool

import (
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
