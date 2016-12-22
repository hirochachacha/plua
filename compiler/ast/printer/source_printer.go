package printer

import (
	"io"
	"strings"
	"text/tabwriter"

	"github.com/hirochachacha/plua/compiler/ast"
	"github.com/hirochachacha/plua/compiler/token"
	"github.com/hirochachacha/plua/internal/strconv"
	"github.com/hirochachacha/plua/position"
)

const (
	indent  = "  "
	rewrite = true
)

type mode uint

const (
	noBlank mode = 1 << iota
	escape
	noParen
	compact
	insertSemi
)

type printer struct {
	w          *tabwriter.Writer
	depth      int // indent depth
	exprDepth  int
	formfeed   bool
	stmtEnd    bool
	lastPos    position.Position
	commentPos position.Position
	cindex     int
	comments   []*ast.CommentGroup
	comment    *ast.CommentGroup
	err        error
}

var initPos = position.Position{Line: 1}

func newPrinter(w io.Writer) *printer {
	return &printer{
		w:       tabwriter.NewWriter(w, 2, 2, 1, ' ', tabwriter.DiscardEmptyColumns|tabwriter.StripEscape),
		lastPos: initPos,
	}
}

// Nodes

func (p *printer) printNode(node ast.Node) {
	switch node := node.(type) {
	case *ast.BadExpr:
		p.print(node.Pos(), "BadExpr", 0)
	case *ast.Name:
		p.printName(node, 0)
	case *ast.Vararg:
		p.printVararg(node, 0)
	case *ast.BasicLit:
		p.printBasicLit(node, 0)
	case *ast.FuncLit:
		p.printFuncLit(node, 0)
	case *ast.TableLit:
		p.printTableLit(node, 0)
	case *ast.ParenExpr:
		p.printParenExpr(node, 0)
	case *ast.SelectorExpr:
		p.printSelectorExpr(node, 0)
	case *ast.IndexExpr:
		p.printIndexExpr(node, 0)
	case *ast.CallExpr:
		p.printCallExpr(node, 0)
	case *ast.UnaryExpr:
		p.printUnaryExpr(node, 0)
	case *ast.BinaryExpr:
		p.printBinaryExpr(node, token.HighestPrec, 0)
	case *ast.KeyValueExpr:
		p.printKeyValueExpr(node, 0)
	case *ast.BadStmt:
		p.print(node.Pos(), "BadStmt", 0)
	case *ast.EmptyStmt:
		// skip this
	case *ast.LocalAssignStmt:
		p.printLocalAssignStmt(node)
	case *ast.LocalFuncStmt:
		p.printLocalFuncStmt(node)
	case *ast.FuncStmt:
		p.printFuncStmt(node)
	case *ast.LabelStmt:
		p.printLabelStmt(node)
	case *ast.ExprStmt:
		p.printExprStmt(node)
	case *ast.AssignStmt:
		p.printAssignStmt(node)
	case *ast.GotoStmt:
		p.printGotoStmt(node)
	case *ast.BreakStmt:
		p.printBreakStmt(node)
	case *ast.IfStmt:
		p.printIfStmt(node)
	case *ast.DoStmt:
		p.printDoStmt(node)
	case *ast.WhileStmt:
		p.printWhileStmt(node)
	case *ast.RepeatStmt:
		p.printRepeatStmt(node)
	case *ast.ReturnStmt:
		p.printReturnStmt(node)
	case *ast.ForStmt:
		p.printForStmt(node)
	case *ast.ForEachStmt:
		p.printForEachStmt(node)
	case *ast.File:
		p.printFile(node)
	case *ast.Block:
		p.printBlock(node)
	case *ast.FuncBody:
		p.printFuncBody(node)
	case *ast.Comment:
	case *ast.CommentGroup:
	case *ast.ParamList:
		p.printParams(node)

	default:
		panic("unreachable")
	}
}

// Statements

func (p *printer) printStmt(stmt ast.Stmt) {
	switch stmt := stmt.(type) {
	case *ast.BadStmt:
		p.print(stmt.Pos(), "BadStmt", 0)
	case *ast.EmptyStmt:
		// skip this
	case *ast.LocalAssignStmt:
		p.printLocalAssignStmt(stmt)
	case *ast.LocalFuncStmt:
		p.printLocalFuncStmt(stmt)
	case *ast.FuncStmt:
		p.printFuncStmt(stmt)
	case *ast.LabelStmt:
		p.printLabelStmt(stmt)
	case *ast.ExprStmt:
		p.printExprStmt(stmt)
	case *ast.AssignStmt:
		p.printAssignStmt(stmt)
	case *ast.GotoStmt:
		p.printGotoStmt(stmt)
	case *ast.BreakStmt:
		p.printBreakStmt(stmt)
	case *ast.IfStmt:
		p.printIfStmt(stmt)
	case *ast.DoStmt:
		p.printDoStmt(stmt)
	case *ast.WhileStmt:
		p.printWhileStmt(stmt)
	case *ast.RepeatStmt:
		p.printRepeatStmt(stmt)
	case *ast.ReturnStmt:
		p.printReturnStmt(stmt)
	case *ast.ForStmt:
		p.printForStmt(stmt)
	case *ast.ForEachStmt:
		p.printForEachStmt(stmt)
	default:
		panic("unreachable")
	}

	p.stmtEnd = true
	p.exprDepth = 0
}

func (p *printer) printLocalAssignStmt(stmt *ast.LocalAssignStmt) {
	p.print(stmt.Local, "local", insertSemi)
	p.exprDepth++
	p.printNames(stmt.LHS, 0)
	if stmt.Equal.IsValid() {
		p.print(stmt.Equal, "=", 0)
		p.printExprs(stmt.RHS, noParen)
	}
	p.exprDepth--
}

func (p *printer) printLocalFuncStmt(stmt *ast.LocalFuncStmt) {
	p.print(stmt.Local, "local", insertSemi)
	p.print(stmt.Func, "function", 0)
	p.exprDepth++
	p.printName(stmt.Name, 0)
	p.exprDepth--
	p.printFuncBody(stmt.Body)
	p.print(stmt.EndPos, "end", 0)
}

func (p *printer) printFuncStmt(stmt *ast.FuncStmt) {
	p.print(stmt.Func, "function", insertSemi)
	p.exprDepth++
	if len(stmt.PathList) > 0 {
		if len(stmt.PathList) == 1 {
			path := stmt.PathList[0]
			p.print(path.Pos(), path.Name, 0)
		} else {
			path := stmt.PathList[0]
			p.print(path.Pos(), path.Name+".", 0)
			for _, path := range stmt.PathList[1 : len(stmt.PathList)-1] {
				p.print(path.Pos(), path.Name+".", noBlank)
			}
			path = stmt.PathList[len(stmt.PathList)-1]
			p.print(path.Pos(), path.Name, noBlank)
		}
		p.print(stmt.AccessPos, stmt.AccessTok.String(), noBlank)
		p.printName(stmt.Name, noBlank)
	} else {
		p.printName(stmt.Name, 0)
	}
	p.exprDepth--
	p.printFuncBody(stmt.Body)
	p.print(stmt.EndPos, "end", 0)
}

func (p *printer) printLabelStmt(stmt *ast.LabelStmt) {
	p.print(stmt.Label, "::", insertSemi)
	p.exprDepth++
	p.print(stmt.Name.Pos(), stmt.Name.Name, noBlank)
	p.print(stmt.EndLabel, "::", noBlank)
	p.exprDepth--
}

func (p *printer) printExprStmt(stmt *ast.ExprStmt) {
	p.printCallExpr(stmt.X, insertSemi)
}

func (p *printer) printAssignStmt(stmt *ast.AssignStmt) {
	p.printExprs(stmt.LHS, noParen|insertSemi)
	p.exprDepth++
	p.print(stmt.Equal, "=", 0)
	p.printExprs(stmt.RHS, noParen)
	p.exprDepth--
}

func (p *printer) printGotoStmt(stmt *ast.GotoStmt) {
	p.print(stmt.Goto, "goto", insertSemi)
	p.exprDepth++
	p.printName(stmt.Label, 0)
	p.exprDepth--
}

func (p *printer) printBreakStmt(stmt *ast.BreakStmt) {
	p.print(stmt.Break, "break", insertSemi)
}

func (p *printer) printIfStmt(stmt *ast.IfStmt) {
	p.print(stmt.If, "if", insertSemi)
	p.exprDepth++
	p.printExpr(stmt.Cond, noParen)
	p.exprDepth--
	p.print(stmt.Then, "then", 0)
	p.printBlock(stmt.Body)
	for _, e := range stmt.ElseIfList {
		p.print(e.If, "elseif", 0)
		p.exprDepth++
		p.printExpr(e.Cond, noParen)
		p.exprDepth--
		p.print(e.Then, "then", 0)
		p.printBlock(e.Body)
	}
	if stmt.ElseBody != nil {
		p.print(stmt.Else, "else", 0)
		p.printBlock(stmt.ElseBody)
	}
	p.print(stmt.EndPos, "end", 0)
}

func (p *printer) printDoStmt(stmt *ast.DoStmt) {
	p.print(stmt.Do, "do", insertSemi)
	p.printBlock(stmt.Body)
	p.print(stmt.EndPos, "end", 0)
}

func (p *printer) printWhileStmt(stmt *ast.WhileStmt) {
	p.print(stmt.While, "while", insertSemi)
	p.exprDepth++
	p.printExpr(stmt.Cond, noParen)
	p.exprDepth--
	p.print(stmt.Do, "do", 0)
	p.printBlock(stmt.Body)
	p.print(stmt.EndPos, "end", 0)
}

func (p *printer) printRepeatStmt(stmt *ast.RepeatStmt) {
	p.print(stmt.Repeat, "repeat", insertSemi)
	p.printBlock(stmt.Body)
	p.print(stmt.Until, "until", 0)
	p.exprDepth++
	p.printExpr(stmt.Cond, noParen)
	p.exprDepth--
}

func (p *printer) printReturnStmt(stmt *ast.ReturnStmt) {
	p.print(stmt.Return, "return", insertSemi)
	p.printExprs(stmt.Results, noParen)
}

func (p *printer) printForStmt(stmt *ast.ForStmt) {
	p.print(stmt.For, "for", insertSemi)
	p.exprDepth++
	p.printName(stmt.Name, 0)
	p.print(stmt.Equal, "=", 0)
	p.printExpr(stmt.Start, noParen)
	p.print(p.lastPos, ",", noBlank)
	if stmt.Step != nil {
		p.printExpr(stmt.Finish, noParen)
		p.print(p.lastPos, ",", noBlank)
		p.printExpr(stmt.Step, noParen)
	} else {
		p.printExpr(stmt.Finish, noParen)
	}
	p.exprDepth--
	p.print(stmt.Do, "do", 0)
	p.printBlock(stmt.Body)
	p.print(stmt.EndPos, "end", 0)
}

func (p *printer) printForEachStmt(stmt *ast.ForEachStmt) {
	p.print(stmt.For, "for", insertSemi)
	p.exprDepth++
	p.printNames(stmt.Names, 0)
	p.print(stmt.In, "in", 0)
	p.printExprs(stmt.Exprs, noParen)
	p.exprDepth--
	p.print(stmt.Do, "do", 0)
	p.printBlock(stmt.Body)
	p.print(stmt.EndPos, "end", 0)
}

// Expression

func (p *printer) printExpr(expr ast.Expr, mode mode) {
	if rewrite {
		if mode&noParen != 0 {
			for { // ((expr)) => expr
				paren, ok := expr.(*ast.ParenExpr)
				if !ok {
					break
				}
				expr = paren.X
			}
			mode &^= noParen
		}
	}

	switch expr := expr.(type) {
	case *ast.BadExpr:
		p.print(expr.Pos(), "BadExpr", mode)
	case *ast.Name:
		p.printName(expr, mode)
	case *ast.Vararg:
		p.printVararg(expr, mode)
	case *ast.BasicLit:
		p.printBasicLit(expr, mode)
	case *ast.FuncLit:
		p.printFuncLit(expr, mode)
	case *ast.TableLit:
		p.printTableLit(expr, mode)
	case *ast.ParenExpr:
		p.printParenExpr(expr, mode)
	case *ast.SelectorExpr:
		p.printSelectorExpr(expr, mode)
	case *ast.IndexExpr:
		p.printIndexExpr(expr, mode)
	case *ast.CallExpr:
		p.printCallExpr(expr, mode)
	case *ast.UnaryExpr:
		p.printUnaryExpr(expr, mode)
	case *ast.BinaryExpr:
		p.printBinaryExpr(expr, token.HighestPrec, mode)
	case *ast.KeyValueExpr:
		p.printKeyValueExpr(expr, mode)
	default:
		panic("unreachable")
	}
}

func (p *printer) printName(expr *ast.Name, mode mode) {
	p.print(expr.Pos(), expr.Name, mode)
}

func (p *printer) printVararg(expr *ast.Vararg, mode mode) {
	p.print(expr.Ellipsis, "...", mode)
}

func (p *printer) printBasicLit(expr *ast.BasicLit, mode mode) {
	if expr.Token.Type == token.STRING {
		p.print(expr.Token.Pos, expr.Token.Lit, mode|escape)
	} else {
		p.print(expr.Token.Pos, expr.Token.Lit, mode)
	}
}

func (p *printer) printFuncLit(expr *ast.FuncLit, mode mode) {
	onLine := expr.Func.Line-p.lastPos.Line == 0
	if onLine {
		p.exprDepth--
	}
	p.print(expr.Func, "function", mode)
	p.printFuncBody(expr.Body)
	p.print(expr.EndPos, "end", 0)
	if onLine {
		p.exprDepth++
	}
}

func (p *printer) printTableLit(expr *ast.TableLit, mode mode) {
	onLine := expr.Lbrace.Line-p.lastPos.Line == 0
	if onLine {
		p.exprDepth--
	}
	if expr.Lbrace.Line == expr.Rbrace.Line {
		p.print(expr.Lbrace, "{", mode)
		p.printExprs(expr.Fields, noBlank|noParen)
		p.print(expr.Rbrace, "}", noBlank)
	} else {
		p.print(expr.Lbrace, "{", mode)
		p.incIndent(expr.Lbrace)
		p.printExprs(expr.Fields, noBlank|noParen)
		if len(expr.Fields) > 0 {
			if expr.Rbrace.Line-p.lastPos.Line > 0 {
				p.writeByte(',')
			}
		}
		p.decIndent(expr.Rbrace)
		p.print(expr.Rbrace, "}", noBlank)
	}
	if onLine {
		p.exprDepth++
	}
}

func (p *printer) printParenExpr(expr *ast.ParenExpr, mode mode) {
	onLine := expr.Lparen.Line-p.lastPos.Line == 0
	if onLine {
		p.exprDepth--
	}
	if expr.Lparen.Line == expr.Rparen.Line {
		p.print(expr.Lparen, "(", mode)
		p.printExpr(expr.X, noBlank|noParen|compact)
		p.print(expr.Rparen, ")", noBlank)
	} else {
		p.print(expr.Lparen, "(", mode)
		p.incIndent(expr.Lparen)
		p.printExpr(expr.X, noBlank|noParen|compact)
		p.decIndent(expr.Rparen)
		p.print(expr.Rparen, ")", noBlank)
	}
	if onLine {
		p.exprDepth++
	}
}

func (p *printer) printSelectorExpr(expr *ast.SelectorExpr, mode mode) {
	p.printExpr(expr.X, mode)
	p.print(expr.Period, ".", noBlank)
	p.printName(expr.Sel, noBlank)
}

func (p *printer) printIndexExpr(expr *ast.IndexExpr, mode mode) {
	p.printExpr(expr.X, mode)

	onLine := expr.Lbrack.Line-p.lastPos.Line == 0
	if onLine {
		p.exprDepth--
	}
	if expr.Lbrack.Line == expr.Rbrack.Line {
		p.print(expr.Lbrack, "[", noBlank)
		p.printExpr(expr.Index, noBlank|noParen|compact)
		p.print(expr.Rbrack, "]", noBlank)
	} else {
		p.print(expr.Lbrack, "[", noBlank)
		p.incIndent(expr.Lbrack)
		p.printExpr(expr.Index, noBlank|noParen|compact)
		p.decIndent(expr.Rbrack)
		p.print(expr.Rbrack, "]", noBlank)
	}
	if onLine {
		p.exprDepth++
	}
}

func (p *printer) printCallExpr(expr *ast.CallExpr, mode mode) {
	p.printExpr(expr.X, mode)
	if expr.Colon != position.NoPos {
		p.print(expr.Colon, ":", noBlank)
		p.printName(expr.Name, noBlank)
	}
	if expr.Lparen != position.NoPos {
		onLine := expr.Lparen.Line-p.lastPos.Line == 0
		if onLine {
			p.exprDepth--
		}
		if expr.Lparen.Line == expr.Rparen.Line {
			p.print(expr.Lparen, "(", noBlank)
			p.printExprs(expr.Args, noBlank|noParen)
			p.print(expr.Rparen, ")", noBlank)
		} else {
			p.print(expr.Lparen, "(", noBlank)
			p.incIndent(expr.Lparen)
			p.printExprs(expr.Args, noBlank|noParen)
			p.decIndent(expr.Rparen)
			p.print(expr.Rparen, ")", noBlank)
		}
		if onLine {
			p.exprDepth++
		}
	} else {
		p.printExprs(expr.Args, noParen)
	}
}

func (p *printer) printUnaryExpr(expr *ast.UnaryExpr, mode mode) {
	p.print(expr.OpPos, expr.Op.String(), mode)
	switch x := expr.X.(type) {
	case *ast.UnaryExpr:
		if rewrite {
			// - - 6 => -(-6)
			// not not true => not (not true)
			if expr.Op != token.NOT {
				p.print(x.Pos(), "(", noBlank)
			} else {
				p.print(x.Pos(), "(", 0)
			}
			p.printUnaryExpr(x, noBlank)
			p.print(x.End(), ")", noBlank)
		} else {
			// - - 6 => - -6
			p.printUnaryExpr(x, 0)
		}
	case *ast.BinaryExpr:
		if expr.Op != token.NOT {
			p.printBinaryExpr(x, token.HighestPrec, noBlank)
		} else {
			p.printBinaryExpr(x, token.HighestPrec, 0)
		}
	default:
		if expr.Op != token.NOT {
			p.printExpr(expr.X, noBlank)
		} else {
			p.printExpr(expr.X, 0)
		}
	}
}

func (p *printer) printBinaryExpr(expr *ast.BinaryExpr, prec1 int, mode mode) {
	prec, _ := expr.Op.Precedence()

	if (mode&compact != 0 || prec > prec1) && prec > 3 { // should cutoff?
		// (1 + 8 * 9) => (1+8*9)
		// 1 + 2 * 3 +4/5 +6 ^ 7 +8 => 1 + 2*3 + 4/5 + 6^7 + 8

		if x, ok := expr.X.(*ast.BinaryExpr); ok {
			p.printBinaryExpr(x, prec1, mode)
		} else {
			p.printExpr(expr.X, mode)
		}

		p.print(expr.OpPos, expr.Op.String(), noBlank)

		switch y := expr.Y.(type) {
		case *ast.UnaryExpr:
			p.printUnaryExpr(y, noBlank)
		case *ast.BinaryExpr:
			p.printBinaryExpr(y, prec1, noBlank)
		default:
			p.printExpr(expr.Y, noBlank)
		}
	} else {
		if x, ok := expr.X.(*ast.BinaryExpr); ok {
			p.printBinaryExpr(x, prec, mode)
		} else {
			p.printExpr(expr.X, mode)
		}

		p.print(expr.OpPos, expr.Op.String(), 0)

		switch y := expr.Y.(type) {
		case *ast.UnaryExpr:
			switch expr.Op.String() + y.Op.String() {
			case "--", "~~":
				if rewrite {
					// 1 - - 2 => 1 - (-2)
					// 1 ~ ~ 2 => 1 ~ (~2)
					p.print(y.Pos(), "(", 0)
					p.printUnaryExpr(y, noBlank)
					p.print(y.End(), ")", noBlank)
				} else {
					// 1 - - 2 => 1 - -2
					// 1 ~ ~ 2 => 1 ~ ~2
					p.printUnaryExpr(y, 0)
				}
			default:
				p.printUnaryExpr(y, 0)
			}
		case *ast.BinaryExpr:
			p.printBinaryExpr(y, prec, 0)
		default:
			p.printExpr(expr.Y, 0)
		}
	}
}

func (p *printer) printKeyValueExpr(expr *ast.KeyValueExpr, mode mode) {
	if expr.Lbrack != position.NoPos {
		onLine := expr.Lbrack.Line-p.lastPos.Line == 0
		if onLine {
			p.exprDepth--
		}
		if expr.Lbrack.Line == expr.Rbrack.Line {
			p.print(expr.Lbrack, "[", mode)
			p.printExpr(expr.Key, noBlank|noParen|compact)
			p.print(expr.Rbrack, "]", noBlank)
		} else {
			p.print(expr.Lbrack, "[", mode)
			p.incIndent(expr.Lbrack)
			p.printExpr(expr.Key, noBlank|noParen|compact)
			p.decIndent(expr.Rbrack)
			p.print(expr.Rbrack, "]", noBlank)
		}
		if onLine {
			p.exprDepth++
		}

		p.exprDepth++
		p.print(expr.Equal, "=", 0)
		p.printExpr(expr.Value, noParen)
		p.exprDepth--
	} else {
		p.printExpr(expr.Key, mode)
		p.exprDepth++
		p.print(expr.Equal, "=", 0)
		p.printExpr(expr.Value, noParen)
		p.exprDepth--
	}
}

func (p *printer) nextComment() {
	if p.cindex == len(p.comments) {
		p.comment = nil
		p.commentPos = position.Position{Line: 1 << 30}
	} else {
		p.comment = p.comments[p.cindex]
		p.commentPos = p.comment.Pos()
		p.cindex++
	}
}

// Other

func replaceEscape(s string) string {
	if i := strings.IndexByte(s, tabwriter.Escape); i != -1 {
		bs := make([]byte, len(s))
		for i := range bs {
			c := s[i]
			if c == tabwriter.Escape {
				bs[i] = 0x00
			} else {
				bs[i] = s[i]
			}
		}
		return string(bs)
	}
	return s
}

func (p *printer) printFile(file *ast.File) {
	if file.Shebang != "" {
		p.writeByte(tabwriter.Escape)
		p.writeString(replaceEscape(file.Shebang))
		p.writeByte(tabwriter.Escape)
	}

	p.comments = file.Comments

	p.nextComment()

	for _, stmt := range file.Chunk {
		p.printStmt(stmt)
	}

	p.insertComment(file.End())
}

func (p *printer) printFuncBody(body *ast.FuncBody) {
	p.printParams(body.Params)
	p.printBlock(body.Body)
}

func (p *printer) printParams(params *ast.ParamList) {
	onLine := params.Lparen.Line-p.lastPos.Line == 0
	if onLine {
		p.exprDepth--
	}
	if params.Lparen.Line == params.Rparen.Line {
		p.print(params.Lparen, "(", noBlank)
		p.printNames(params.List, noBlank)
		if params.Ellipsis != position.NoPos {
			if len(params.List) > 0 {
				p.print(p.lastPos, ",", noBlank)
			}
			p.print(params.Ellipsis, "...", 0)
		}
		p.print(params.Rparen, ")", noBlank)
	} else {
		p.print(params.Lparen, "(", noBlank)
		p.incIndent(params.Lparen)
		p.printNames(params.List, noBlank)
		if params.Ellipsis != position.NoPos {
			if len(params.List) > 0 {
				p.print(p.lastPos, ",", noBlank)
			}
			p.print(params.Ellipsis, "...", 0)
		}
		p.decIndent(params.Rparen)
		p.print(params.Rparen, ")", noBlank)
	}
	if onLine {
		p.exprDepth++
	}
}

func (p *printer) printBlock(block *ast.Block) {
	p.incIndent(block.Opening)
	for _, stmt := range block.List {
		p.printStmt(stmt)
	}
	p.decIndent(block.Closing)
}

func (p *printer) printNames(names []*ast.Name, mode mode) {
	if len(names) == 0 {
		return
	}
	p.printName(names[0], mode)
	for _, name := range names[1:] {
		p.print(p.lastPos, ",", noBlank)
		p.printName(name, 0)
	}
}

func (p *printer) printExprs(exprs []ast.Expr, mode mode) {
	switch len(exprs) {
	case 0:
		return
	case 1:
		p.printExpr(exprs[0], mode)
	default:
		p.printExpr(exprs[0], mode|compact)
		for _, expr := range exprs[1:] {
			p.print(p.lastPos, ",", noBlank)
			p.printExpr(expr, noParen|compact)
		}
	}
}

func (p *printer) incIndent(pos position.Position) {
	p.insertComment(pos)

	p.formfeed = true

	p.depth++
}

func (p *printer) decIndent(end position.Position) {
	p.insertComment(end)

	p.depth--

	p.formfeed = true
}

func (p *printer) insertComment(pos position.Position) {
	for p.commentPos.LessThan(pos) {
		for i, c := range p.comment.List {
			d := c.Pos().Line - p.lastPos.Line

			switch {
			case d == 0:
				if p.lastPos != initPos {
					p.writeByte('\t')
				}
			case d > 0:
				if i == 0 {
					p.writeByte('\f')
				} else {
					if p.formfeed {
						p.writeByte('\f')
					} else {
						p.writeByte('\n')
					}
				}
				if d > 1 {
					p.writeByte('\f')
				}
				for i := 0; i < p.depth; i++ {
					p.writeString(indent)
				}
				p.formfeed = false
			default:
				panic("unexpected")
			}

			p.writeByte(tabwriter.Escape)
			p.writeString(replaceEscape(strings.TrimRight(c.Text, "\t\n\v\f\r ")))
			p.writeByte(tabwriter.Escape)

			p.lastPos = c.End()
		}

		p.nextComment()
	}
}

func (p *printer) print(pos position.Position, s string, mode mode) {
	p.insertComment(pos)

	d := pos.Line - p.lastPos.Line

	switch {
	case d == 0:
		if mode&noBlank == 0 {
			if p.lastPos != initPos {
				if p.stmtEnd && mode&insertSemi != 0 {
					p.writeByte(';')
				}
				p.writeByte(' ')
			}
		}
	case d > 0:
		if p.formfeed {
			p.writeByte('\f')
		} else {
			p.writeByte('\n')
		}
		if d > 1 {
			p.writeByte('\f')
		}
		for i := 0; i < p.exprDepth; i++ {
			p.writeString(indent)
		}
		for i := 0; i < p.depth; i++ {
			p.writeString(indent)
		}
		p.formfeed = false
	default:
		panic("unexpected")
	}

	if mode&escape != 0 {
		p.writeByte(tabwriter.Escape)
		p.writeString(strconv.Escape(s))
		p.writeByte(tabwriter.Escape)
	} else {
		p.writeString(s)
	}

	p.lastPos = pos.Offset(s)

	p.stmtEnd = false
}

func (p *printer) writeByte(c byte) {
	if p.err != nil {
		return
	}
	_, p.err = p.w.Write([]byte{c})
}

func (p *printer) writeString(s string) {
	if p.err != nil {
		return
	}
	_, p.err = p.w.Write([]byte(s))
}
