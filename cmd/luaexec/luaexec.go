package main

import (
	"bufio"
	"flag"
	"fmt"
	"os"
	"strings"

	"github.com/hirochachacha/plua/compiler"
	"github.com/hirochachacha/plua/compiler/parser"
	"github.com/hirochachacha/plua/object"
	"github.com/hirochachacha/plua/position"
	"github.com/hirochachacha/plua/runtime"
	"github.com/hirochachacha/plua/stdlib"

	isatty "github.com/mattn/go-isatty"
)

func main() {
	flag.Usage = func() {
		fmt.Fprintln(os.Stderr, "usage: luaexec [file]")
		flag.PrintDefaults()
	}

	flag.Parse()

	switch {
	case len(os.Args) >= 2:
		c := compiler.NewCompiler()

		proto, err := c.CompileFile(os.Args[1], compiler.Either)
		if err != nil {
			object.PrintError(err)
			os.Exit(1)
		}

		exec(proto)
	case !isatty.IsTerminal(os.Stdin.Fd()):
		c := compiler.NewCompiler()

		proto, err := c.Compile(os.Stdin, "=stdin", compiler.Either)
		if err != nil {
			object.PrintError(err)
			os.Exit(1)
		}

		exec(proto)
	default:
		interact()
	}

}

func exec(proto *object.Proto) {
	p := runtime.NewProcess()

	var a object.Table

	if len(os.Args) >= 2 {
		a = p.NewTableSize(len(os.Args)-2, 2)
		for i, arg := range os.Args {
			a.Set(object.Integer(i-1), object.String(arg))
		}
	} else {
		a = p.NewTableSize(0, 1)
		a.Set(object.Integer(0), object.String(os.Args[0]))
	}

	p.Globals().Set(object.String("arg"), a)

	p.Require("", stdlib.Open)

	_, err := p.Exec(proto)
	if err != nil {
		object.PrintError(err)
		os.Exit(1)
	}
}

func eof(s string) position.Position {
	line := 0
	column := 0
	for _, r := range s {
		if r == '\n' {
			line++
			column = 0
		} else {
			column++
		}
	}
	line++
	column++
	return position.Position{Line: line, Column: column}
}

func isIncomplete(code string, err error) bool {
	if err, ok := err.(*parser.Error); ok {
		eof := eof(code)
		return err.Pos.Line == eof.Line && err.Pos.Column == eof.Column
	}

	return false
}

func interact() {
	c := compiler.NewCompiler()

	p := runtime.NewProcess()

	p.Require("", stdlib.Open)

	stdin := bufio.NewScanner(os.Stdin)

	var code string

	for {
		var err error

		if len(code) != 0 {
			_, err = fmt.Fprint(os.Stdout, ">> ")
		} else {
			_, err = fmt.Fprint(os.Stdout, "> ")
		}

		if err != nil {
			panic(err)
		}

		if !stdin.Scan() {
			if err := stdin.Err(); err != nil {
				panic(err)
			}

			return
		}

		line := stdin.Text()
		var proto *object.Proto

		if len(code) == 0 {
			if line == "exit" {
				return
			}

			code = "return " + line

			proto, err = c.Compile(strings.NewReader(code), "=stdin", compiler.Text)
			if err != nil {
				code = line

				proto, err = c.Compile(strings.NewReader(code), "=stdin", compiler.Text)
			}
		} else {
			code += "\n" + line

			proto, err = c.Compile(strings.NewReader(code), "=stdin", compiler.Text)
		}

		if err != nil {
			if isIncomplete(code, err) {
				continue
			}

			code = ""

			object.PrintError(err)

			continue
		}

		code = ""

		rets, err := p.Exec(proto)
		if err != nil {
			object.PrintError(err)
		} else {
			if len(rets) > 0 {
				fmt.Fprintln(os.Stdout, object.Repr(rets[0]))
			}
		}

		p = p.Fork()
	}
}
