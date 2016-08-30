package printer

import (
	"fmt"
	"io"
	"strings"

	"github.com/hirochachacha/plua/compiler/ast"
	"github.com/hirochachacha/plua/internal/strconv"
)

const treeIndent = "  "

type treeprinter struct {
	w io.Writer
}

func (pr treeprinter) print(node ast.Node, prefix string, depth int) {
	ind := strings.Repeat(treeIndent, depth)
	nind := ind + treeIndent

	switch node := node.(type) {
	case *ast.Comment:
		fmt.Fprintf(pr.w, "%s%s { %s-%s\n", prefix, node.Type(), node.Pos(), node.End())
		fmt.Fprintf(pr.w, "%sHyphen: %s\n", nind, node.Hyphen)
		fmt.Fprintf(pr.w, "%sText: %s\n", nind, strconv.Quote(node.Text))
		fmt.Fprintf(pr.w, "%s}\n", ind)
	case *ast.CommentGroup:
		fmt.Fprintf(pr.w, "%s%s { %s-%s\n", prefix, node.Type(), node.Pos(), node.End())
		fmt.Fprintf(pr.w, "%sList: {\n", nind)
		for _, e := range node.List {
			pr.print(e, nind+treeIndent, depth+2)
		}
		fmt.Fprintf(pr.w, "%s}\n", nind)
		fmt.Fprintf(pr.w, "%s}\n", ind)
	case *ast.ParamList:
		fmt.Fprintf(pr.w, "%s%s { %s-%s\n", prefix, node.Type(), node.Pos(), node.End())
		fmt.Fprintf(pr.w, "%sLparen: %s\n", nind, node.Lparen)
		fmt.Fprintf(pr.w, "%sList: {\n", nind)
		for _, e := range node.List {
			pr.print(e, nind+treeIndent, depth+2)
		}
		fmt.Fprintf(pr.w, "%s}\n", nind)
		fmt.Fprintf(pr.w, "%sEllipsis: %s\n", nind, node.Ellipsis)
		fmt.Fprintf(pr.w, "%sRparen: %s\n", nind, node.Rparen)
		fmt.Fprintf(pr.w, "%s}\n", ind)

	case *ast.BadExpr:
		fmt.Fprintf(pr.w, "%s%s { %s-%s\n", prefix, node.Type(), node.Pos(), node.End())
		fmt.Fprintf(pr.w, "%sFrom: %s\n", nind, node.From)
		fmt.Fprintf(pr.w, "%sTo: %s\n", nind, node.To)
		fmt.Fprintf(pr.w, "%s}\n", ind)
	case *ast.Name:
		fmt.Fprintf(pr.w, "%s%s { %s-%s\n", prefix, node.Type(), node.Pos(), node.End())
		fmt.Fprintf(pr.w, "%sNamePos: %s\n", nind, node.NamePos)
		fmt.Fprintf(pr.w, "%sName: %s\n", nind, node.Name)
		// fmt.Fprintf(pr.w, "%sIsLHS: %t\n", nind, node.IsLHS)
		fmt.Fprintf(pr.w, "%s}\n", ind)
	case *ast.Vararg:
		fmt.Fprintf(pr.w, "%s%s { %s-%s\n", prefix, node.Type(), node.Pos(), node.End())
		fmt.Fprintf(pr.w, "%sEllipsis: %s\n", nind, node.Ellipsis)
		fmt.Fprintf(pr.w, "%s}\n", ind)
	case *ast.BasicLit:
		fmt.Fprintf(pr.w, "%s%s { %s-%s\n", prefix, node.Type(), node.Pos(), node.End())
		fmt.Fprintf(pr.w, "%sToken.Type: %s\n", nind, node.Token.Type)
		fmt.Fprintf(pr.w, "%sToken.Pos: %s\n", nind, node.Token.Pos)
		fmt.Fprintf(pr.w, "%sToken.Lit: %s\n", nind, node.Token.Lit)
		fmt.Fprintf(pr.w, "%s}\n", ind)
	case *ast.FuncLit:
		fmt.Fprintf(pr.w, "%s%s { %s-%s\n", prefix, node.Type(), node.Pos(), node.End())
		fmt.Fprintf(pr.w, "%sFunc: %s\n", nind, node.Func)
		pr.print(node.Body, fmt.Sprintf("%sBody: ", nind), depth+1)
		fmt.Fprintf(pr.w, "%s}\n", ind)
	case *ast.TableLit:
		fmt.Fprintf(pr.w, "%s%s { %s-%s\n", prefix, node.Type(), node.Pos(), node.End())
		fmt.Fprintf(pr.w, "%sLbrace: %s\n", nind, node.Lbrace)
		fmt.Fprintf(pr.w, "%sFields: {\n", nind)
		for _, e := range node.Fields {
			pr.print(e, nind+treeIndent, depth+2)
		}
		fmt.Fprintf(pr.w, "%s}\n", nind)
		fmt.Fprintf(pr.w, "%sRbrace: %s\n", nind, node.Rbrace)
		fmt.Fprintf(pr.w, "%s}\n", ind)
	case *ast.ParenExpr:
		fmt.Fprintf(pr.w, "%s%s { %s-%s\n", prefix, node.Type(), node.Pos(), node.End())
		fmt.Fprintf(pr.w, "%sLparen: %s\n", nind, node.Lparen)
		pr.print(node.X, fmt.Sprintf("%sX: ", nind), depth+1)
		fmt.Fprintf(pr.w, "%sRparen: %s\n", nind, node.Rparen)
		fmt.Fprintf(pr.w, "%s}\n", ind)
	case *ast.SelectorExpr:
		fmt.Fprintf(pr.w, "%s%s { %s-%s\n", prefix, node.Type(), node.Pos(), node.End())
		pr.print(node.X, fmt.Sprintf("%sX: ", nind), depth+1)
		pr.print(node.Sel, fmt.Sprintf("%sSel: ", nind), depth+1)
		// fmt.Fprintf(pr.w, "%sIsLHS: %t\n", nind, node.IsLHS)
		fmt.Fprintf(pr.w, "%s}\n", ind)
	case *ast.IndexExpr:
		fmt.Fprintf(pr.w, "%s%s { %s-%s\n", prefix, node.Type(), node.Pos(), node.End())
		pr.print(node.X, fmt.Sprintf("%sX: ", nind), depth+1)
		fmt.Fprintf(pr.w, "%sLbrack: %s\n", nind, node.Lbrack)
		pr.print(node.Index, fmt.Sprintf("%sIndex: ", nind), depth+1)
		fmt.Fprintf(pr.w, "%sRbrack: %s\n", nind, node.Rbrack)
		// fmt.Fprintf(pr.w, "%sIsLHS: %t\n", nind, node.IsLHS)
		fmt.Fprintf(pr.w, "%s}\n", ind)
	case *ast.CallExpr:
		fmt.Fprintf(pr.w, "%s%s { %s-%s\n", prefix, node.Type(), node.Pos(), node.End())
		pr.print(node.X, fmt.Sprintf("%sX: ", nind), depth+1)
		fmt.Fprintf(pr.w, "%sColon: %s\n", nind, node.Colon)
		if node.Name != nil {
			pr.print(node.Name, fmt.Sprintf("%sName: ", nind), depth+1)
		} else {
			fmt.Fprintf(pr.w, "%sName: nil\n", nind)
		}
		fmt.Fprintf(pr.w, "%sLparen: %s\n", nind, node.Lparen)
		fmt.Fprintf(pr.w, "%sArgs: {\n", nind)
		for _, e := range node.Args {
			pr.print(e, nind+treeIndent, depth+2)
		}
		fmt.Fprintf(pr.w, "%s}\n", nind)
		fmt.Fprintf(pr.w, "%sRparen: %s\n", nind, node.Rparen)
		fmt.Fprintf(pr.w, "%s}\n", ind)
	case *ast.UnaryExpr:
		fmt.Fprintf(pr.w, "%s%s { %s-%s\n", prefix, node.Type(), node.Pos(), node.End())
		fmt.Fprintf(pr.w, "%sOpPos: %s\n", nind, node.OpPos)
		fmt.Fprintf(pr.w, "%sOp: %s\n", nind, node.Op)
		pr.print(node.X, fmt.Sprintf("%sX: ", nind), depth+1)
		fmt.Fprintf(pr.w, "%s}\n", ind)
	case *ast.BinaryExpr:
		fmt.Fprintf(pr.w, "%s%s { %s-%s\n", prefix, node.Type(), node.Pos(), node.End())
		pr.print(node.X, fmt.Sprintf("%sX: ", nind), depth+1)
		fmt.Fprintf(pr.w, "%sOpPos: %s\n", nind, node.OpPos)
		fmt.Fprintf(pr.w, "%sOp: %s\n", nind, node.Op)
		pr.print(node.Y, fmt.Sprintf("%sY: ", nind), depth+1)
		fmt.Fprintf(pr.w, "%s}\n", ind)
	case *ast.KeyValueExpr:
		fmt.Fprintf(pr.w, "%s%s { %s-%s\n", prefix, node.Type(), node.Pos(), node.End())
		fmt.Fprintf(pr.w, "%sLbrack: %s\n", nind, node.Lbrack)
		pr.print(node.Key, fmt.Sprintf("%sKey: ", nind), depth+1)
		fmt.Fprintf(pr.w, "%sRbrack: %s\n", nind, node.Rbrack)
		fmt.Fprintf(pr.w, "%sEqual: %s\n", nind, node.Equal)
		pr.print(node.Value, fmt.Sprintf("%sValue: ", nind), depth+1)
		fmt.Fprintf(pr.w, "%s}\n", ind)

	case *ast.BadStmt:
		fmt.Fprintf(pr.w, "%s%s { %s-%s\n", prefix, node.Type(), node.Pos(), node.End())
		fmt.Fprintf(pr.w, "%sFrom: %s\n", nind, node.From)
		fmt.Fprintf(pr.w, "%sTo: %s\n", nind, node.To)
		fmt.Fprintf(pr.w, "%s}\n", ind)
	case *ast.EmptyStmt:
		fmt.Fprintf(pr.w, "%s%s { %s-%s\n", prefix, node.Type(), node.Pos(), node.End())
		fmt.Fprintf(pr.w, "%sSemicolon: %s\n", nind, node.Semicolon)
		fmt.Fprintf(pr.w, "%s}\n", ind)
	case *ast.LocalAssignStmt:
		fmt.Fprintf(pr.w, "%s%s { %s-%s\n", prefix, node.Type(), node.Pos(), node.End())
		fmt.Fprintf(pr.w, "%sLHS: {\n", nind)
		for _, e := range node.LHS {
			pr.print(e, nind+treeIndent, depth+2)
		}
		fmt.Fprintf(pr.w, "%s}\n", nind)
		fmt.Fprintf(pr.w, "%sEqual: %s\n", nind, node.Equal)
		fmt.Fprintf(pr.w, "%sRHS: {\n", nind)
		for _, e := range node.RHS {
			pr.print(e, nind+treeIndent, depth+2)
		}
		fmt.Fprintf(pr.w, "%s}\n", nind)
		fmt.Fprintf(pr.w, "%s}\n", ind)
	case *ast.LocalFuncStmt:
		fmt.Fprintf(pr.w, "%s%s { %s-%s\n", prefix, node.Type(), node.Pos(), node.End())
		fmt.Fprintf(pr.w, "%sLocal: %s\n", nind, node.Local)
		fmt.Fprintf(pr.w, "%sFunc: %s\n", nind, node.Func)
		pr.print(node.Name, fmt.Sprintf("%sName: ", nind), depth+1)
		pr.print(node.Body, fmt.Sprintf("%sBody: ", nind), depth+1)
		fmt.Fprintf(pr.w, "%s}\n", ind)
	case *ast.FuncStmt:
		fmt.Fprintf(pr.w, "%s%s { %s-%s\n", prefix, node.Type(), node.Pos(), node.End())
		fmt.Fprintf(pr.w, "%sFunc: %s\n", nind, node.Func)
		fmt.Fprintf(pr.w, "%sNamePrefix: {\n", nind)
		for _, e := range node.NamePrefix {
			pr.print(e, nind+treeIndent, depth+2)
		}
		fmt.Fprintf(pr.w, "%s}\n", nind)
		pr.print(node.Name, fmt.Sprintf("%sName: ", nind), depth+1)
		fmt.Fprintf(pr.w, "%sAccessTok: '%s'\n", nind, node.AccessTok)
		fmt.Fprintf(pr.w, "%sAccessPos: %s\n", nind, node.AccessPos)
		pr.print(node.Body, fmt.Sprintf("%sBody: ", nind), depth+1)
		fmt.Fprintf(pr.w, "%s}\n", ind)
	case *ast.LabelStmt:
		fmt.Fprintf(pr.w, "%s%s { %s-%s\n", prefix, node.Type(), node.Pos(), node.End())
		fmt.Fprintf(pr.w, "%sLabel: %s\n", nind, node.Label)
		pr.print(node.Name, fmt.Sprintf("%sName: ", nind), depth+1)
		fmt.Fprintf(pr.w, "%sEndLabel: %s\n", nind, node.EndLabel)
		fmt.Fprintf(pr.w, "%s}\n", ind)
	case *ast.ExprStmt:
		fmt.Fprintf(pr.w, "%s%s { %s-%s\n", prefix, node.Type(), node.Pos(), node.End())
		pr.print(node.X, fmt.Sprintf("%sX: ", nind), depth+1)
		fmt.Fprintf(pr.w, "%s}\n", ind)
	case *ast.AssignStmt:
		fmt.Fprintf(pr.w, "%s%s { %s-%s\n", prefix, node.Type(), node.Pos(), node.End())
		fmt.Fprintf(pr.w, "%sLHS: {\n", nind)
		for _, e := range node.LHS {
			pr.print(e, nind+treeIndent, depth+2)
		}
		fmt.Fprintf(pr.w, "%s}\n", nind)
		fmt.Fprintf(pr.w, "%sEqual: %s\n", nind, node.Equal)
		fmt.Fprintf(pr.w, "%sRHS: {\n", nind)
		for _, e := range node.RHS {
			pr.print(e, nind+treeIndent, depth+2)
		}
		fmt.Fprintf(pr.w, "%s}\n", nind)
		fmt.Fprintf(pr.w, "%s}\n", ind)
	case *ast.GotoStmt:
		fmt.Fprintf(pr.w, "%s%s { %s-%s\n", prefix, node.Type(), node.Pos(), node.End())
		fmt.Fprintf(pr.w, "%sGoto: %s\n", nind, node.Goto)
		pr.print(node.Label, fmt.Sprintf("%sLabel: ", nind), depth+1)
		fmt.Fprintf(pr.w, "%s}\n", ind)
	case *ast.BreakStmt:
		fmt.Fprintf(pr.w, "%s%s { %s-%s\n", prefix, node.Type(), node.Pos(), node.End())
		fmt.Fprintf(pr.w, "%sBreak: %s\n", nind, node.Break)
		fmt.Fprintf(pr.w, "%s}\n", ind)
	case *ast.IfStmt:
		fmt.Fprintf(pr.w, "%s%s { %s-%s\n", prefix, node.Type(), node.Pos(), node.End())
		fmt.Fprintf(pr.w, "%sIf: %s\n", nind, node.If)
		pr.print(node.Cond, fmt.Sprintf("%sCond: ", nind), depth+1)
		pr.print(node.Body, fmt.Sprintf("%sBody: ", nind), depth+1)
		pr.print(node.Else, fmt.Sprintf("%sElse: ", nind), depth+1)
		fmt.Fprintf(pr.w, "%s}\n", ind)
	case *ast.DoStmt:
		fmt.Fprintf(pr.w, "%s%s { %s-%s\n", prefix, node.Type(), node.Pos(), node.End())
		pr.print(node.Body, fmt.Sprintf("%sBody: ", nind), depth+1)
		fmt.Fprintf(pr.w, "%s}\n", ind)
	case *ast.WhileStmt:
		fmt.Fprintf(pr.w, "%s%s { %s-%s\n", prefix, node.Type(), node.Pos(), node.End())
		fmt.Fprintf(pr.w, "%sWhile: %s\n", nind, node.While)
		pr.print(node.Cond, fmt.Sprintf("%sCond: ", nind), depth+1)
		pr.print(node.Body, fmt.Sprintf("%sBody: ", nind), depth+1)
		fmt.Fprintf(pr.w, "%s}\n", ind)
	case *ast.RepeatStmt:
		fmt.Fprintf(pr.w, "%s%s { %s-%s\n", prefix, node.Type(), node.Pos(), node.End())
		fmt.Fprintf(pr.w, "%sRepeat: %s\n", nind, node.Repeat)
		pr.print(node.Body, fmt.Sprintf("%sBody: ", nind), depth+1)
		fmt.Fprintf(pr.w, "%sUntil: %s\n", nind, node.Until)
		pr.print(node.Cond, fmt.Sprintf("%sCond: ", nind), depth+1)
		fmt.Fprintf(pr.w, "%s}\n", ind)
	case *ast.ReturnStmt:
		fmt.Fprintf(pr.w, "%s%s { %s-%s\n", prefix, node.Type(), node.Pos(), node.End())
		fmt.Fprintf(pr.w, "%sReturn: %s\n", nind, node.Return)
		fmt.Fprintf(pr.w, "%sResults: {\n", nind)
		for _, e := range node.Results {
			pr.print(e, nind+treeIndent, depth+2)
		}
		fmt.Fprintf(pr.w, "%s}\n", nind)
		fmt.Fprintf(pr.w, "%s}\n", ind)
	case *ast.ForStmt:
		fmt.Fprintf(pr.w, "%s%s { %s-%s\n", prefix, node.Type(), node.Pos(), node.End())
		fmt.Fprintf(pr.w, "%sFor: %s\n", nind, node.For)
		pr.print(node.Name, fmt.Sprintf("%sName: ", nind), depth+1)
		fmt.Fprintf(pr.w, "%sEqual: %s\n", nind, node.Equal)
		pr.print(node.Start, fmt.Sprintf("%sStart: ", nind), depth+1)
		pr.print(node.Finish, fmt.Sprintf("%sFinish: ", nind), depth+1)
		if node.Step != nil {
			pr.print(node.Step, fmt.Sprintf("%sStep: ", nind), depth+1)
		} else {
			fmt.Fprintf(pr.w, "%sStep: nil\n", nind)
		}
		pr.print(node.Body, fmt.Sprintf("%sBody: ", nind), depth+1)
		fmt.Fprintf(pr.w, "%s}\n", ind)
	case *ast.ForEachStmt:
		fmt.Fprintf(pr.w, "%s%s { %s-%s\n", prefix, node.Type(), node.Pos(), node.End())
		fmt.Fprintf(pr.w, "%sFor: %s\n", nind, node.For)
		fmt.Fprintf(pr.w, "%sNames: {\n", nind)
		for _, e := range node.Names {
			pr.print(e, nind+treeIndent, depth+2)
		}
		fmt.Fprintf(pr.w, "%s}\n", nind)
		fmt.Fprintf(pr.w, "%sIn: %s\n", nind, node.In)
		fmt.Fprintf(pr.w, "%sExprs: {\n", nind)
		for _, e := range node.Exprs {
			pr.print(e, nind+treeIndent, depth+2)
		}
		fmt.Fprintf(pr.w, "%s}\n", nind)
		pr.print(node.Body, fmt.Sprintf("%sBody: ", nind), depth+1)
		fmt.Fprintf(pr.w, "%s}\n", ind)

	case *ast.File:
		fmt.Fprintf(pr.w, "%s%s { %s-%s\n", prefix, node.Type(), node.Pos(), node.End())
		fmt.Fprintf(pr.w, "%sShebang: %s\n", nind, strconv.Quote(node.Shebang))
		fmt.Fprintf(pr.w, "%sChunk: {\n", nind)
		for _, e := range node.Chunk {
			pr.print(e, nind+treeIndent, depth+2)
		}
		fmt.Fprintf(pr.w, "%s}\n", nind)
		fmt.Fprintf(pr.w, "%sComments: {\n", nind)
		for _, e := range node.Comments {
			pr.print(e, nind+treeIndent, depth+2)
		}
		fmt.Fprintf(pr.w, "%s}\n", nind)
		fmt.Fprintf(pr.w, "%s}\n", ind)
	case *ast.Block:
		fmt.Fprintf(pr.w, "%s%s { %s-%s\n", prefix, node.Type(), node.Pos(), node.End())
		fmt.Fprintf(pr.w, "%sOpening: %s\n", nind, node.Opening)
		fmt.Fprintf(pr.w, "%sList: {\n", nind)
		for _, e := range node.List {
			pr.print(e, nind+treeIndent, depth+2)
		}
		fmt.Fprintf(pr.w, "%s}\n", nind)
		fmt.Fprintf(pr.w, "%sClosing: %s\n", nind, node.Closing)
		fmt.Fprintf(pr.w, "%s}\n", ind)
	case *ast.FuncBody:
		fmt.Fprintf(pr.w, "%s%s { %s-%s\n", prefix, node.Type(), node.Pos(), node.End())
		pr.print(node.Params, fmt.Sprintf("%sParams: ", nind), depth+1)
		pr.print(node.Body, fmt.Sprintf("%sBody: ", nind), depth+1)
		fmt.Fprintf(pr.w, "%s}\n", ind)
	default:
		panic("unreachable")
	}
}
