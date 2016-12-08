package printer

import (
	"fmt"
	"io"
	"reflect"
	"strings"

	"github.com/hirochachacha/plua/compiler/ast"
	"github.com/hirochachacha/plua/internal/strconv"
)

const treeIndent = "  "

type treeprinter struct {
	w   io.Writer
	err error
}

func (p treeprinter) printf(s string, args ...interface{}) {
	if p.err != nil {
		return
	}
	_, p.err = fmt.Fprintf(p.w, s, args...)
}

func (p treeprinter) printNode(node ast.Node, prefix string, depth int) {
	if p.err != nil {
		return
	}

	ind := strings.Repeat(treeIndent, depth)
	nind := ind + treeIndent

	if reflect.ValueOf(node).IsNil() {
		p.printf("%snil\n", prefix)

		return
	}

	switch node := node.(type) {
	case *ast.Comment:
		p.printf("%s%s { %s-%s\n", prefix, node.Type(), node.Pos(), node.End())
		p.printf("%sHyphen: %s\n", nind, node.Hyphen)
		p.printf("%sText: %s\n", nind, strconv.Quote(node.Text))
		p.printf("%s}\n", ind)
	case *ast.CommentGroup:
		p.printf("%s%s { %s-%s\n", prefix, node.Type(), node.Pos(), node.End())
		p.printf("%sList: {\n", nind)
		for _, e := range node.List {
			p.printNode(e, nind+treeIndent, depth+2)
		}
		p.printf("%s}\n", nind)
		p.printf("%s}\n", ind)
	case *ast.ParamList:
		p.printf("%s%s { %s-%s\n", prefix, node.Type(), node.Pos(), node.End())
		p.printf("%sLparen: %s\n", nind, node.Lparen)
		p.printf("%sList: {\n", nind)
		for _, e := range node.List {
			p.printNode(e, nind+treeIndent, depth+2)
		}
		p.printf("%s}\n", nind)
		p.printf("%sEllipsis: %s\n", nind, node.Ellipsis)
		p.printf("%sRparen: %s\n", nind, node.Rparen)
		p.printf("%s}\n", ind)

	case *ast.BadExpr:
		p.printf("%s%s { %s-%s\n", prefix, node.Type(), node.Pos(), node.End())
		p.printf("%sFrom: %s\n", nind, node.From)
		p.printf("%sTo: %s\n", nind, node.To)
		p.printf("%s}\n", ind)
	case *ast.Name:
		p.printf("%s%s { %s-%s\n", prefix, node.Type(), node.Pos(), node.End())
		p.printf("%sNamePos: %s\n", nind, node.NamePos)
		p.printf("%sName: %s\n", nind, node.Name)
		// p.printf("%sIsLHS: %t\n", nind, node.IsLHS)
		p.printf("%s}\n", ind)
	case *ast.Vararg:
		p.printf("%s%s { %s-%s\n", prefix, node.Type(), node.Pos(), node.End())
		p.printf("%sEllipsis: %s\n", nind, node.Ellipsis)
		p.printf("%s}\n", ind)
	case *ast.BasicLit:
		p.printf("%s%s { %s-%s\n", prefix, node.Type(), node.Pos(), node.End())
		p.printf("%sToken.Type: %s\n", nind, node.Token.Type)
		p.printf("%sToken.Pos: %s\n", nind, node.Token.Pos)
		p.printf("%sToken.Lit: %s\n", nind, node.Token.Lit)
		p.printf("%s}\n", ind)
	case *ast.FuncLit:
		p.printf("%s%s { %s-%s\n", prefix, node.Type(), node.Pos(), node.End())
		p.printf("%sFunc: %s\n", nind, node.Func)
		p.printNode(node.Body, fmt.Sprintf("%sBody: ", nind), depth+1)
		p.printf("%sEndPos: %s\n", nind, node.EndPos)
		p.printf("%s}\n", ind)
	case *ast.TableLit:
		p.printf("%s%s { %s-%s\n", prefix, node.Type(), node.Pos(), node.End())
		p.printf("%sLbrace: %s\n", nind, node.Lbrace)
		p.printf("%sFields: {\n", nind)
		for _, e := range node.Fields {
			p.printNode(e, nind+treeIndent, depth+2)
		}
		p.printf("%s}\n", nind)
		p.printf("%sRbrace: %s\n", nind, node.Rbrace)
		p.printf("%s}\n", ind)
	case *ast.ParenExpr:
		p.printf("%s%s { %s-%s\n", prefix, node.Type(), node.Pos(), node.End())
		p.printf("%sLparen: %s\n", nind, node.Lparen)
		p.printNode(node.X, fmt.Sprintf("%sX: ", nind), depth+1)
		p.printf("%sRparen: %s\n", nind, node.Rparen)
		p.printf("%s}\n", ind)
	case *ast.SelectorExpr:
		p.printf("%s%s { %s-%s\n", prefix, node.Type(), node.Pos(), node.End())
		p.printNode(node.X, fmt.Sprintf("%sX: ", nind), depth+1)
		p.printNode(node.Sel, fmt.Sprintf("%sSel: ", nind), depth+1)
		// p.printf("%sIsLHS: %t\n", nind, node.IsLHS)
		p.printf("%s}\n", ind)
	case *ast.IndexExpr:
		p.printf("%s%s { %s-%s\n", prefix, node.Type(), node.Pos(), node.End())
		p.printNode(node.X, fmt.Sprintf("%sX: ", nind), depth+1)
		p.printf("%sLbrack: %s\n", nind, node.Lbrack)
		p.printNode(node.Index, fmt.Sprintf("%sIndex: ", nind), depth+1)
		p.printf("%sRbrack: %s\n", nind, node.Rbrack)
		// p.printf("%sIsLHS: %t\n", nind, node.IsLHS)
		p.printf("%s}\n", ind)
	case *ast.CallExpr:
		p.printf("%s%s { %s-%s\n", prefix, node.Type(), node.Pos(), node.End())
		p.printNode(node.X, fmt.Sprintf("%sX: ", nind), depth+1)
		p.printf("%sColon: %s\n", nind, node.Colon)
		p.printNode(node.Name, fmt.Sprintf("%sName: ", nind), depth+1)
		p.printf("%sLparen: %s\n", nind, node.Lparen)
		p.printf("%sArgs: {\n", nind)
		for _, e := range node.Args {
			p.printNode(e, nind+treeIndent, depth+2)
		}
		p.printf("%s}\n", nind)
		p.printf("%sRparen: %s\n", nind, node.Rparen)
		p.printf("%s}\n", ind)
	case *ast.UnaryExpr:
		p.printf("%s%s { %s-%s\n", prefix, node.Type(), node.Pos(), node.End())
		p.printf("%sOpPos: %s\n", nind, node.OpPos)
		p.printf("%sOp: %s\n", nind, node.Op)
		p.printNode(node.X, fmt.Sprintf("%sX: ", nind), depth+1)
		p.printf("%s}\n", ind)
	case *ast.BinaryExpr:
		p.printf("%s%s { %s-%s\n", prefix, node.Type(), node.Pos(), node.End())
		p.printNode(node.X, fmt.Sprintf("%sX: ", nind), depth+1)
		p.printf("%sOpPos: %s\n", nind, node.OpPos)
		p.printf("%sOp: %s\n", nind, node.Op)
		p.printNode(node.Y, fmt.Sprintf("%sY: ", nind), depth+1)
		p.printf("%s}\n", ind)
	case *ast.KeyValueExpr:
		p.printf("%s%s { %s-%s\n", prefix, node.Type(), node.Pos(), node.End())
		p.printf("%sLbrack: %s\n", nind, node.Lbrack)
		p.printNode(node.Key, fmt.Sprintf("%sKey: ", nind), depth+1)
		p.printf("%sRbrack: %s\n", nind, node.Rbrack)
		p.printf("%sEqual: %s\n", nind, node.Equal)
		p.printNode(node.Value, fmt.Sprintf("%sValue: ", nind), depth+1)
		p.printf("%s}\n", ind)

	case *ast.BadStmt:
		p.printf("%s%s { %s-%s\n", prefix, node.Type(), node.Pos(), node.End())
		p.printf("%sFrom: %s\n", nind, node.From)
		p.printf("%sTo: %s\n", nind, node.To)
		p.printf("%s}\n", ind)
	case *ast.EmptyStmt:
		p.printf("%s%s { %s-%s\n", prefix, node.Type(), node.Pos(), node.End())
		p.printf("%sSemicolon: %s\n", nind, node.Semicolon)
		p.printf("%s}\n", ind)
	case *ast.LocalAssignStmt:
		p.printf("%s%s { %s-%s\n", prefix, node.Type(), node.Pos(), node.End())
		p.printf("%sLHS: {\n", nind)
		for _, e := range node.LHS {
			p.printNode(e, nind+treeIndent, depth+2)
		}
		p.printf("%s}\n", nind)
		p.printf("%sEqual: %s\n", nind, node.Equal)
		p.printf("%sRHS: {\n", nind)
		for _, e := range node.RHS {
			p.printNode(e, nind+treeIndent, depth+2)
		}
		p.printf("%s}\n", nind)
		p.printf("%s}\n", ind)
	case *ast.LocalFuncStmt:
		p.printf("%s%s { %s-%s\n", prefix, node.Type(), node.Pos(), node.End())
		p.printf("%sLocal: %s\n", nind, node.Local)
		p.printf("%sFunc: %s\n", nind, node.Func)
		p.printNode(node.Name, fmt.Sprintf("%sName: ", nind), depth+1)
		p.printNode(node.Body, fmt.Sprintf("%sBody: ", nind), depth+1)
		p.printf("%sEndPos: %s\n", nind, node.EndPos)
		p.printf("%s}\n", ind)
	case *ast.FuncStmt:
		p.printf("%s%s { %s-%s\n", prefix, node.Type(), node.Pos(), node.End())
		p.printf("%sFunc: %s\n", nind, node.Func)
		p.printf("%sPathList: {\n", nind)
		for _, e := range node.PathList {
			p.printNode(e, nind+treeIndent, depth+2)
		}
		p.printf("%s}\n", nind)
		p.printNode(node.Name, fmt.Sprintf("%sName: ", nind), depth+1)
		p.printf("%sAccessTok: '%s'\n", nind, node.AccessTok)
		p.printf("%sAccessPos: %s\n", nind, node.AccessPos)
		p.printNode(node.Body, fmt.Sprintf("%sBody: ", nind), depth+1)
		p.printf("%sEndPos: %s\n", nind, node.EndPos)
		p.printf("%s}\n", ind)
	case *ast.LabelStmt:
		p.printf("%s%s { %s-%s\n", prefix, node.Type(), node.Pos(), node.End())
		p.printf("%sLabel: %s\n", nind, node.Label)
		p.printNode(node.Name, fmt.Sprintf("%sName: ", nind), depth+1)
		p.printf("%sEndLabel: %s\n", nind, node.EndLabel)
		p.printf("%s}\n", ind)
	case *ast.ExprStmt:
		p.printf("%s%s { %s-%s\n", prefix, node.Type(), node.Pos(), node.End())
		p.printNode(node.X, fmt.Sprintf("%sX: ", nind), depth+1)
		p.printf("%s}\n", ind)
	case *ast.AssignStmt:
		p.printf("%s%s { %s-%s\n", prefix, node.Type(), node.Pos(), node.End())
		p.printf("%sLHS: {\n", nind)
		for _, e := range node.LHS {
			p.printNode(e, nind+treeIndent, depth+2)
		}
		p.printf("%s}\n", nind)
		p.printf("%sEqual: %s\n", nind, node.Equal)
		p.printf("%sRHS: {\n", nind)
		for _, e := range node.RHS {
			p.printNode(e, nind+treeIndent, depth+2)
		}
		p.printf("%s}\n", nind)
		p.printf("%s}\n", ind)
	case *ast.GotoStmt:
		p.printf("%s%s { %s-%s\n", prefix, node.Type(), node.Pos(), node.End())
		p.printf("%sGoto: %s\n", nind, node.Goto)
		p.printNode(node.Label, fmt.Sprintf("%sLabel: ", nind), depth+1)
		p.printf("%s}\n", ind)
	case *ast.BreakStmt:
		p.printf("%s%s { %s-%s\n", prefix, node.Type(), node.Pos(), node.End())
		p.printf("%sBreak: %s\n", nind, node.Break)
		p.printf("%s}\n", ind)
	case *ast.IfStmt:
		p.printf("%s%s { %s-%s\n", prefix, node.Type(), node.Pos(), node.End())
		p.printf("%sIf: %s\n", nind, node.If)
		p.printNode(node.Cond, fmt.Sprintf("%sCond: ", nind), depth+1)
		p.printf("%sThen: %s\n", nind, node.Then)
		p.printNode(node.Body, fmt.Sprintf("%sBody: ", nind), depth+1)
		p.printf("%sElseIfList: {\n", nind)
		for _, e := range node.ElseIfList {
			p.printf("%sElseIf: %s\n", nind+treeIndent, e.If)
			p.printNode(e.Cond, fmt.Sprintf("%sCond: ", nind+treeIndent), depth+2)
			p.printf("%sThen: %s\n", nind+treeIndent, e.Then)
			p.printNode(e.Body, fmt.Sprintf("%sBody: ", nind+treeIndent), depth+2)
		}
		p.printf("%s}\n", nind)
		p.printf("%sElse: %s\n", nind, node.Else)
		p.printNode(node.ElseBody, fmt.Sprintf("%sElseBody: ", nind), depth+1)
		p.printf("%sEndPos: %s\n", nind, node.EndPos)
		p.printf("%s}\n", ind)
	case *ast.DoStmt:
		p.printf("%s%s { %s-%s\n", prefix, node.Type(), node.Pos(), node.End())
		p.printf("%sDo: %s\n", nind, node.Do)
		p.printNode(node.Body, fmt.Sprintf("%sBody: ", nind), depth+1)
		p.printf("%sEndPos: %s\n", nind, node.EndPos)
		p.printf("%s}\n", ind)
	case *ast.WhileStmt:
		p.printf("%s%s { %s-%s\n", prefix, node.Type(), node.Pos(), node.End())
		p.printf("%sWhile: %s\n", nind, node.While)
		p.printNode(node.Cond, fmt.Sprintf("%sCond: ", nind), depth+1)
		p.printf("%sDo: %s\n", nind, node.Do)
		p.printNode(node.Body, fmt.Sprintf("%sBody: ", nind), depth+1)
		p.printf("%sEndPos: %s\n", nind, node.EndPos)
		p.printf("%s}\n", ind)
	case *ast.RepeatStmt:
		p.printf("%s%s { %s-%s\n", prefix, node.Type(), node.Pos(), node.End())
		p.printf("%sRepeat: %s\n", nind, node.Repeat)
		p.printNode(node.Body, fmt.Sprintf("%sBody: ", nind), depth+1)
		p.printf("%sUntil: %s\n", nind, node.Until)
		p.printNode(node.Cond, fmt.Sprintf("%sCond: ", nind), depth+1)
		p.printf("%s}\n", ind)
	case *ast.ReturnStmt:
		p.printf("%s%s { %s-%s\n", prefix, node.Type(), node.Pos(), node.End())
		p.printf("%sReturn: %s\n", nind, node.Return)
		p.printf("%sResults: {\n", nind)
		for _, e := range node.Results {
			p.printNode(e, nind+treeIndent, depth+2)
		}
		p.printf("%s}\n", nind)
		p.printf("%s}\n", ind)
	case *ast.ForStmt:
		p.printf("%s%s { %s-%s\n", prefix, node.Type(), node.Pos(), node.End())
		p.printf("%sFor: %s\n", nind, node.For)
		p.printNode(node.Name, fmt.Sprintf("%sName: ", nind), depth+1)
		p.printf("%sEqual: %s\n", nind, node.Equal)
		p.printNode(node.Start, fmt.Sprintf("%sStart: ", nind), depth+1)
		p.printNode(node.Finish, fmt.Sprintf("%sFinish: ", nind), depth+1)
		p.printNode(node.Step, fmt.Sprintf("%sStep: ", nind), depth+1)
		p.printf("%sDo: %s\n", nind, node.Do)
		p.printNode(node.Body, fmt.Sprintf("%sBody: ", nind), depth+1)
		p.printf("%sEndPos: %s\n", nind, node.EndPos)
		p.printf("%s}\n", ind)
	case *ast.ForEachStmt:
		p.printf("%s%s { %s-%s\n", prefix, node.Type(), node.Pos(), node.End())
		p.printf("%sFor: %s\n", nind, node.For)
		p.printf("%sNames: {\n", nind)
		for _, e := range node.Names {
			p.printNode(e, nind+treeIndent, depth+2)
		}
		p.printf("%s}\n", nind)
		p.printf("%sIn: %s\n", nind, node.In)
		p.printf("%sExprs: {\n", nind)
		for _, e := range node.Exprs {
			p.printNode(e, nind+treeIndent, depth+2)
		}
		p.printf("%s}\n", nind)
		p.printf("%sDo: %s\n", nind, node.Do)
		p.printNode(node.Body, fmt.Sprintf("%sBody: ", nind), depth+1)
		p.printf("%sEndPos: %s\n", nind, node.EndPos)
		p.printf("%s}\n", ind)
	case *ast.File:
		p.printf("%s%s { %s-%s\n", prefix, node.Type(), node.Pos(), node.End())
		p.printf("%sShebang: %s\n", nind, strconv.Quote(node.Shebang))
		p.printf("%sChunk: {\n", nind)
		for _, e := range node.Chunk {
			p.printNode(e, nind+treeIndent, depth+2)
		}
		p.printf("%s}\n", nind)
		p.printf("%sComments: {\n", nind)
		for _, e := range node.Comments {
			p.printNode(e, nind+treeIndent, depth+2)
		}
		p.printf("%s}\n", nind)
		p.printf("%s}\n", ind)
	case *ast.Block:
		p.printf("%s%s { %s-%s\n", prefix, node.Type(), node.Pos(), node.End())
		p.printf("%sOpening: %s\n", nind, node.Opening)
		p.printf("%sList: {\n", nind)
		for _, e := range node.List {
			p.printNode(e, nind+treeIndent, depth+2)
		}
		p.printf("%s}\n", nind)
		p.printf("%sClosing: %s\n", nind, node.Closing)
		p.printf("%s}\n", ind)
	case *ast.FuncBody:
		p.printf("%s%s { %s-%s\n", prefix, node.Type(), node.Pos(), node.End())
		p.printNode(node.Params, fmt.Sprintf("%sParams: ", nind), depth+1)
		p.printNode(node.Body, fmt.Sprintf("%sBody: ", nind), depth+1)
		p.printf("%s}\n", ind)
	default:
		panic("unreachable")
	}
}
