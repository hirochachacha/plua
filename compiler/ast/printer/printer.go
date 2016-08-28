package printer

import (
	"io"
	"os"

	"github.com/hirochachacha/blua/compiler/ast"
)

func FprintTree(w io.Writer, node ast.Node) {
	treeprinter{w}.print(node, "", 0)
}

func PrintTree(node ast.Node) {
	FprintTree(os.Stdout, node)
}

func Fprint(w io.Writer, node ast.Node) {
	p := newPrinter(w)

	p.printNode(node)

	p.w.Flush()
}

func Print(node ast.Node) {
	Fprint(os.Stdout, node)
}
