package printer

import (
	"io"

	"github.com/hirochachacha/plua/compiler/ast"
)

func FprintTree(w io.Writer, node ast.Node) error {
	p := treeprinter{w: w}
	p.printNode(node, "", 0)
	return p.err
}

func Fprint(w io.Writer, node ast.Node) error {
	p := newPrinter(w)
	p.printNode(node)
	if p.err != nil {
		return p.err
	}
	if err := p.w.Flush(); err != nil {
		return err
	}
	return nil
}
