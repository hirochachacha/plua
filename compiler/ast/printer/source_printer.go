package printer

import (
	"io"
	"strings"
	"text/tabwriter"
	"unicode"

	"github.com/hirochachacha/blua/compiler/ast"
	"github.com/hirochachacha/blua/compiler/token"
	"github.com/hirochachacha/blua/position"
)

const (
	infinity = 1 << 30
	equal    = "="
	indent   = "  "
	period   = "."
	comma    = ","
	lbrace   = "{"
	rbrace   = "}"
	lparen   = "("
	rparen   = ")"
	lbrack   = "["
	rbrack   = "]"
	colon    = ":"

	tab       = '\t'
	blank     = ' '
	newline   = '\n'
	semicolon = ';'
	formfeed  = '\f'
	escape    = tabwriter.Escape
)

type mode uint

const (
	noExprIndent mode = 1 << iota
	noBlank
	insertSemi
	isLit
)

type printer struct {
	w          *tabwriter.Writer
	commentPos position.Position
	cindex     int
	comment    *ast.CommentGroup
	comments   []*ast.CommentGroup
	depth      int // indent depth
	lastPos    position.Position
	err        error
	skip       bool // skip first byte
}

func newPrinter(w io.Writer) *printer {
	return &printer{
		w:    tabwriter.NewWriter(w, 2, 2, 1, ' ', tabwriter.DiscardEmptyColumns|tabwriter.StripEscape),
		skip: true,
	}
}

// Nodes

func (pr *printer) printNode(node ast.Node) {
	switch node := node.(type) {
	case *ast.BadExpr:
		pr.print(node.Pos(), "BadExpr", 0)
	case *ast.Name:
		pr.printName(node, 0)
	case *ast.Vararg:
		pr.printVararg(node, 0)
	case *ast.BasicLit:
		pr.printBasicLit(node, 0)
	case *ast.FuncLit:
		pr.printFuncLit(node, 0)
	case *ast.TableLit:
		pr.printTableLit(node, 0)
	case *ast.ParenExpr:
		pr.printParenExpr(node, 0)
	case *ast.SelectorExpr:
		pr.printSelectorExpr(node, 0)
	case *ast.IndexExpr:
		pr.printIndexExpr(node, 0)
	case *ast.CallExpr:
		pr.printCallExpr(node, 0)
	case *ast.UnaryExpr:
		pr.printUnaryExpr(node, 0)
	case *ast.BinaryExpr:
		pr.printBinaryExpr(node, token.HighestPrec, 0)
	case *ast.KeyValueExpr:
		pr.print(node.Pos(), "BadKeyValueExpr", 0)

	case *ast.BadStmt:
		pr.print(node.Pos(), "BadStmt", 0)
	case *ast.EmptyStmt:
		// skip this
	case *ast.LocalAssignStmt:
		pr.printLocalAssignStmt(node)
	case *ast.LocalFuncStmt:
		pr.printLocalFuncStmt(node)
	case *ast.FuncStmt:
		pr.printFuncStmt(node)
	case *ast.LabelStmt:
		pr.printLabelStmt(node)
	case *ast.ExprStmt:
		pr.printExprStmt(node)
	case *ast.AssignStmt:
		pr.printAssignStmt(node)
	case *ast.GotoStmt:
		pr.printGotoStmt(node)
	case *ast.BreakStmt:
		pr.printBreakStmt(node)
	case *ast.IfStmt:
		pr.printIfStmt(node)
	case *ast.DoStmt:
		pr.printDoStmt(node)
	case *ast.WhileStmt:
		pr.printWhileStmt(node)
	case *ast.RepeatStmt:
		pr.printRepeatStmt(node)
	case *ast.ReturnStmt:
		pr.printReturnStmt(node)
	case *ast.ForStmt:
		pr.printForStmt(node)
	case *ast.ForEachStmt:
		pr.printForEachStmt(node)

	case *ast.File:
		pr.printFile(node)
	case *ast.Block:
		pr.printBlock(node)
	case *ast.FuncBody:
		pr.printFuncBody(node)
	case *ast.Comment:
	case *ast.CommentGroup:
	case *ast.ParamList:
		pr.printParams(node)

	default:
		panic("unreachable")
	}
}

// Statements

func (pr *printer) printStmt(stmt ast.Stmt) {
	switch stmt := stmt.(type) {
	case *ast.BadStmt:
		pr.print(stmt.Pos(), "BadStmt", 0)
	case *ast.EmptyStmt:
		// skip this
	case *ast.LocalAssignStmt:
		pr.printLocalAssignStmt(stmt)
	case *ast.LocalFuncStmt:
		pr.printLocalFuncStmt(stmt)
	case *ast.FuncStmt:
		pr.printFuncStmt(stmt)
	case *ast.LabelStmt:
		pr.printLabelStmt(stmt)
	case *ast.ExprStmt:
		pr.printExprStmt(stmt)
	case *ast.AssignStmt:
		pr.printAssignStmt(stmt)
	case *ast.GotoStmt:
		pr.printGotoStmt(stmt)
	case *ast.BreakStmt:
		pr.printBreakStmt(stmt)
	case *ast.IfStmt:
		pr.printIfStmt(stmt)
	case *ast.DoStmt:
		pr.printDoStmt(stmt)
	case *ast.WhileStmt:
		pr.printWhileStmt(stmt)
	case *ast.RepeatStmt:
		pr.printRepeatStmt(stmt)
	case *ast.ReturnStmt:
		pr.printReturnStmt(stmt)
	case *ast.ForStmt:
		pr.printForStmt(stmt)
	case *ast.ForEachStmt:
		pr.printForEachStmt(stmt)
	default:
		panic("unreachable")
	}
}

func (pr *printer) printLocalAssignStmt(stmt *ast.LocalAssignStmt) {
	pr.printNames(stmt.LHS, noExprIndent|insertSemi)
	pr.print(stmt.Equal, equal, 0)
	pr.printExprs(stmt.RHS, 0)
}

func (pr *printer) printLocalFuncStmt(stmt *ast.LocalFuncStmt) {
	pr.print(stmt.Local, "local", noExprIndent|insertSemi)
	pr.print(stmt.Func, "function", 0)
	pr.printName(stmt.Name, 0)

	pr.printFuncBody(stmt.Body)

	pr.print(stmt.Body.Body.Closing, "end", noExprIndent)
}

func (pr *printer) printFuncStmt(stmt *ast.FuncStmt) {
	pr.print(stmt.Func, "function", noExprIndent|insertSemi)
	for _, name := range stmt.NamePrefix {
		pr.print(name.Pos(), name.Name+period, 0)
	}

	if stmt.AccessTok != token.ILLEGAL {
		pr.print(stmt.AccessPos, stmt.AccessTok.String(), noBlank)
	}

	pr.printName(stmt.Name, 0)

	pr.printFuncBody(stmt.Body)

	pr.print(stmt.Body.Body.Closing, "end", noExprIndent)
}

func (pr *printer) printLabelStmt(stmt *ast.LabelStmt) {
	pr.print(stmt.Label, "::"+stmt.Name.Name+"::", noExprIndent|insertSemi)
}

func (pr *printer) printExprStmt(stmt *ast.ExprStmt) {
	pr.printCallExpr(stmt.X, noExprIndent|insertSemi)
}

func (pr *printer) printAssignStmt(stmt *ast.AssignStmt) {
	pr.printExprs(stmt.LHS, noExprIndent|insertSemi)
	pr.print(stmt.Equal, equal, 0)
	pr.printExprs(stmt.RHS, 0)
}

func (pr *printer) printGotoStmt(stmt *ast.GotoStmt) {
	pr.print(stmt.Goto, "goto", noExprIndent|insertSemi)
	pr.printName(stmt.Label, 0)
}

func (pr *printer) printBreakStmt(stmt *ast.BreakStmt) {
	pr.print(stmt.Break, "break", noExprIndent|insertSemi)
}

func getElseIf(stmt *ast.IfStmt) (*ast.IfStmt, bool) {
	if stmt.Else != nil && len(stmt.Else.List) == 1 {
		if ifStmt, ok := stmt.Else.List[0].(*ast.IfStmt); ok {
			if stmt.Else.Opening == ifStmt.Pos() {
				return ifStmt, true
			}
		}
	}

	return stmt, false
}

func (pr *printer) printIfStmt(stmt *ast.IfStmt) {
	pr.print(stmt.If, "if", noExprIndent|insertSemi)
	pr.printExpr(stmt.Cond, 0)
	pr.print(stmt.Body.Opening, "then", 0)
	pr.printBlock(stmt.Body)
	if stmt.Else != nil {
		if last, ok := getElseIf(stmt); ok {
			pr.print(last.If, "elseif", noExprIndent|insertSemi)
			pr.printExpr(last.Cond, 0)
			pr.print(last.Body.Opening, "then", 0)
			pr.printBlock(last.Body)

			for {
				last, ok = getElseIf(last)
				if !ok {
					break
				}
				pr.print(last.If, "elseif", noExprIndent|insertSemi)
				pr.printExpr(last.Cond, 0)
				pr.print(last.Body.Opening, "then", 0)
				pr.printBlock(last.Body)
			}

			if last.Else != nil {
				pr.print(last.Else.Opening, "else", 0)
				pr.printBlock(last.Else)
				pr.print(last.Else.Closing, "end", noExprIndent)
			} else {
				pr.print(last.Body.Closing, "end", noExprIndent)
			}

			return
		}

		pr.print(stmt.Else.Opening, "else", 0)
		pr.printBlock(stmt.Else)
		pr.print(stmt.Else.Closing, "end", noExprIndent)
	} else {
		pr.print(stmt.Body.Closing, "end", noExprIndent)
	}
}

func (pr *printer) printDoStmt(stmt *ast.DoStmt) {
	pr.print(stmt.Body.Opening, "do", noExprIndent|insertSemi)

	pr.printBlock(stmt.Body)

	pr.print(stmt.Body.Closing, "end", noExprIndent)
}

func (pr *printer) printWhileStmt(stmt *ast.WhileStmt) {
	pr.print(stmt.While, "while", noExprIndent|insertSemi)

	pr.printExpr(stmt.Cond, 0)

	pr.print(stmt.Body.Opening, "do", 0)

	pr.printBlock(stmt.Body)

	pr.print(stmt.Body.Closing, "end", noExprIndent)
}

func (pr *printer) printRepeatStmt(stmt *ast.RepeatStmt) {
	pr.print(stmt.Repeat, "repeat", noExprIndent|insertSemi)

	pr.printBlock(stmt.Body)

	pr.print(stmt.Until, "until", noExprIndent)

	pr.printExpr(stmt.Cond, 0)
}

func (pr *printer) printReturnStmt(stmt *ast.ReturnStmt) {
	pr.print(stmt.Return, "return", noExprIndent|insertSemi)

	if len(stmt.Results) == 1 {
		if expr, ok := stmt.Results[0].(*ast.ParenExpr); ok {
			pr.printExpr(expr.X, 0)
		} else {
			pr.printExpr(stmt.Results[0], 0)
		}
	} else {
		pr.printExprs(stmt.Results, 0)
	}
}

func (pr *printer) printForStmt(stmt *ast.ForStmt) {
	pr.print(stmt.For, "for", noExprIndent|insertSemi)

	pr.printName(stmt.Name, 0)
	pr.print(stmt.Equal, equal, 0)
	pr.printExpr(stmt.Start, 0)
	pr.print(pr.lastPos, comma, noBlank)
	if stmt.Step != nil {
		pr.printExpr(stmt.Finish, 0)
		pr.print(pr.lastPos, comma, noBlank)
		pr.printExpr(stmt.Step, 0)
	} else {
		pr.printExpr(stmt.Finish, 0)
	}
	pr.print(stmt.Body.Opening, "do", 0)

	pr.printBlock(stmt.Body)

	pr.print(stmt.Body.Closing, "end", noExprIndent)
}

func (pr *printer) printForEachStmt(stmt *ast.ForEachStmt) {
	pr.print(stmt.For, "for", noExprIndent|insertSemi)

	pr.printNames(stmt.Names, 0)
	pr.print(stmt.In, "in", 0)
	pr.printExprs(stmt.Exprs, 0)

	pr.print(stmt.Body.Opening, "do", 0)

	pr.printBlock(stmt.Body)

	pr.print(stmt.Body.Closing, "end", noExprIndent)
}

// Expression

func (pr *printer) printExpr(expr ast.Expr, mode mode) {
	switch expr := expr.(type) {
	case *ast.BadExpr:
		pr.print(expr.Pos(), "BadExpr", mode)
	case *ast.Name:
		pr.printName(expr, mode)
	case *ast.Vararg:
		pr.printVararg(expr, mode)
	case *ast.BasicLit:
		pr.printBasicLit(expr, mode)
	case *ast.FuncLit:
		pr.printFuncLit(expr, mode)
	case *ast.TableLit:
		pr.printTableLit(expr, mode)
	case *ast.ParenExpr:
		pr.printParenExpr(expr, mode)
	case *ast.SelectorExpr:
		pr.printSelectorExpr(expr, mode)
	case *ast.IndexExpr:
		pr.printIndexExpr(expr, mode)
	case *ast.CallExpr:
		pr.printCallExpr(expr, mode)
	case *ast.UnaryExpr:
		pr.printUnaryExpr(expr, mode)
	case *ast.BinaryExpr:
		pr.printBinaryExpr(expr, token.HighestPrec, mode)
	case *ast.KeyValueExpr:
		panic("unexpected")
	default:
		panic("unreachable")
	}
}

func (pr *printer) printName(expr *ast.Name, mode mode) {
	pr.print(expr.Pos(), expr.Name, mode)
}

func (pr *printer) printVararg(expr *ast.Vararg, mode mode) {
	pr.print(expr.Ellipsis, "...", mode)
}

func (pr *printer) printBasicLit(expr *ast.BasicLit, mode mode) {
	if expr.Token.Type == token.STRING {
		lit := expr.Token.Lit
		if lit[0] == '\'' {
			if strings.IndexByte(lit, '"') == -1 {
				lit = "\"" + lit[1:len(lit)-1] + "\""
			}
		}
		pr.print(expr.Token.Pos, lit, mode|isLit)
	} else {
		pr.print(expr.Token.Pos, expr.Token.Lit, mode)
	}
}

func (pr *printer) printFuncLit(expr *ast.FuncLit, mode mode) {
	pr.print(expr.Func, "function", mode)
	pr.printFuncBody(expr.Body)
	pr.print(expr.Body.Body.Closing, "end", noExprIndent)
}

func (pr *printer) printTableLit(expr *ast.TableLit, mode mode) {
	doIndent := pr.lastPos.Line != expr.Lbrace.Line

	pr.print(expr.Lbrace, lbrace, mode)

	// pr.depth++

	pr.printExprs(expr.Fields, 0)

	// pr.depth--

	if doIndent {
		pr.print(expr.Rbrace, rbrace, 0)
	} else {
		pr.print(expr.Rbrace, rbrace, noExprIndent)
	}
}

func (pr *printer) printParenExpr(expr *ast.ParenExpr, mode mode) {
	if x, ok := expr.X.(*ast.ParenExpr); ok {
		pr.printParenExpr(x, mode)
	} else {
		pr.print(expr.Lparen, lparen, mode)
		pr.printExpr(expr.X, noBlank)
		pr.print(expr.Rparen, rparen, noBlank)
	}
}

func (pr *printer) printSelectorExpr(expr *ast.SelectorExpr, mode mode) {
	pr.printExpr(expr.X, mode)
	pr.print(expr.Period, period, noBlank)
	pr.printName(expr.Sel, noBlank)
}

func (pr *printer) printIndexExpr(expr *ast.IndexExpr, mode mode) {
	pr.printExpr(expr.X, mode)

	doIndent := pr.lastPos.Line != expr.Lbrack.Line

	pr.print(expr.Lbrack, lbrack, noBlank)

	// pr.depth++

	pr.printExpr(expr.Index, noBlank)

	// pr.depth--

	if doIndent {
		pr.print(expr.Rbrack, rbrack, noBlank)
	} else {
		pr.print(expr.Rbrack, rbrack, noBlank|noExprIndent)
	}
}

func (pr *printer) printCallExpr(expr *ast.CallExpr, mode mode) {
	pr.printExpr(expr.X, mode)

	if expr.Colon != position.NoPos {
		pr.print(expr.Colon, colon, noBlank)
		pr.printName(expr.Name, noBlank)
	}

	pr.print(expr.Lparen, lparen, noBlank)
	pr.printExprs(expr.Args, noBlank)
	pr.print(expr.Rparen, rparen, noBlank)
}

func (pr *printer) printUnaryExpr(expr *ast.UnaryExpr, mode mode) {
	pr.print(expr.OpPos, expr.Op.String(), mode)
	switch x := expr.X.(type) {
	case *ast.UnaryExpr:
		pr.print(x.Pos(), lparen, noBlank)
		pr.printUnaryExpr(x, noBlank)
		pr.print(x.End(), rparen, noBlank)
	case *ast.BinaryExpr:
		if x.Op == token.POW {
			pr.print(x.Pos(), lparen, noBlank)
			pr.printBinaryExpr(x, token.HighestPrec, noBlank)
			pr.print(x.End(), rparen, noBlank)
		} else {
			pr.printBinaryExpr(x, token.HighestPrec, noBlank)
		}
	default:
		pr.printExpr(expr.X, noBlank)
	}
}

func (pr *printer) printBinaryExpr(expr *ast.BinaryExpr, prec1 int, mode mode) {
	prec, _ := expr.Op.Precedence()

	if prec > prec1 && prec > 3 { // should cutoff?
		if x, ok := expr.X.(*ast.BinaryExpr); ok {
			pr.printBinaryExpr(x, prec1, mode)
		} else {
			pr.printExpr(expr.X, mode)
		}

		pr.print(expr.OpPos, expr.Op.String(), noBlank)

		switch y := expr.Y.(type) {
		case *ast.UnaryExpr:
			switch expr.Op.String() + y.Op.String() {
			case "--", "~~":
				pr.print(y.Pos(), lparen, noBlank)
				pr.printUnaryExpr(y, noBlank)
				pr.print(y.End(), rparen, noBlank)
			default:
				pr.printUnaryExpr(y, noBlank)
			}
		case *ast.BinaryExpr:
			pr.printBinaryExpr(y, prec1, noBlank)
		default:
			pr.printExpr(expr.Y, noBlank)
		}
	} else {
		if x, ok := expr.X.(*ast.BinaryExpr); ok {
			pr.printBinaryExpr(x, prec, mode)
		} else {
			pr.printExpr(expr.X, mode)
		}

		pr.print(expr.OpPos, expr.Op.String(), 0)

		if y, ok := expr.Y.(*ast.BinaryExpr); ok {
			pr.printBinaryExpr(y, prec, 0)
		} else {
			pr.printExpr(expr.Y, 0)
		}
	}
}

func (pr *printer) printKeyValue(expr *ast.KeyValueExpr) {
	if expr.Lbrack != position.NoPos {
		doIndent := pr.lastPos.Line != expr.Lbrack.Line

		pr.print(expr.Lbrack, lbrack, 0)

		// pr.depth++

		pr.printExpr(expr.Key, noBlank)

		// pr.depth--

		if doIndent {
			pr.print(expr.Rbrack, rbrack, noBlank)
		} else {
			pr.print(expr.Rbrack, rbrack, noBlank|noExprIndent)
		}

		pr.print(expr.Equal, equal, 0)

		pr.printExpr(expr.Value, 0)
	} else {
		pr.printExpr(expr.Key, noBlank)
		pr.print(expr.Equal, equal, 0)
		pr.printExpr(expr.Value, 0)
	}
}

// Other

func (pr *printer) printFile(file *ast.File) {
	pr.comments = file.Comments
	pr.nextComment()
	for _, stmt := range file.Chunk {
		pr.printStmt(stmt)
	}

	pr.printLastComment()

	// consume remaining comments
	for pr.comment != nil {
		pr.writeNewLine(pr.commentPos)

		for _, c := range pr.comment.List {
			pr.writeIndent()
			pr.writeByte(escape)
			pr.writeString(trimRight(c.Text))
			pr.writeByte(escape)
		}

		pr.lastPos = pr.comment.End()

		pr.nextComment()
	}
}

func (pr *printer) printFuncBody(body *ast.FuncBody) {
	pr.printParams(body.Params)
	pr.printBlock(body.Body)
}

func (pr *printer) printParams(params *ast.ParamList) {
	pr.print(params.Lparen, lparen, 0)

	pr.printNames(params.List, 0)
	if params.Ellipsis != position.NoPos {
		if len(params.List) > 0 {
			pr.print(pr.lastPos, comma, noBlank)
		}
		pr.print(params.Ellipsis, "...", 0)
	}

	pr.print(params.Rparen, rparen, 0)
}

func (pr *printer) printBlock(block *ast.Block) {
	if len(block.List) > 0 {
		pr.incIndent(block.List[0].Pos())
	} else {
		pr.depth++
	}

	for _, stmt := range block.List {
		pr.printStmt(stmt)
	}

	pr.decIndent(block.Closing)
}

func (pr *printer) printNames(names []*ast.Name, mode mode) {
	if len(names) == 0 {
		return
	}

	pr.printName(names[0], mode)

	for _, name := range names[1:] {
		pr.print(pr.lastPos, comma, noBlank)
		pr.printName(name, 0)
	}
}

func (pr *printer) printExprs(exprs []ast.Expr, mode mode) {
	if len(exprs) == 0 {
		return
	}

	pr.printExpr(exprs[0], mode)

	for _, expr := range exprs[1:] {
		pr.print(pr.lastPos, comma, noBlank)
		pr.printExpr(expr, 0)
	}
}

// Util

func (pr *printer) incIndent(next position.Position) {
	if pr.lastPos.Line != next.Line {
		pr.printLastComment()

		pr.w.Flush()
	}

	pr.depth++
}

func (pr *printer) decIndent(next position.Position) {
	if pr.lastPos.Line != next.Line {
		pr.printLastComment()

		pr.printComments(next)

		pr.w.Flush()
	}

	pr.depth--
}

func (pr *printer) nextComment() {
	if pr.cindex == len(pr.comments) {
		pr.comment = nil
		pr.commentPos = position.Position{Line: infinity}
	} else {
		pr.comment = pr.comments[pr.cindex]
		pr.commentPos = pr.comment.Pos()
		pr.cindex++
	}
}

func (pr *printer) printLastComment() {
	if pr.commentPos.Line == pr.lastPos.Line {
		for _, c := range pr.comment.List {
			pr.writeByte(tab)
			pr.writeByte(escape)
			pr.writeString(trimRight(c.Text))
			pr.writeByte(escape)
		}

		pr.lastPos = pr.comment.End()

		pr.nextComment()
	}
}

func (pr *printer) printComments(pos position.Position) {
	for pr.commentPos.Line < pos.Line {
		pr.writeNewLine(pr.commentPos)

		for _, c := range pr.comment.List {
			pr.writeIndent()
			pr.writeByte(escape)
			pr.writeString(trimRight(c.Text))
			pr.writeByte(escape)
		}

		pr.lastPos = pr.comment.End()

		pr.nextComment()
	}
}

func (pr *printer) print(pos position.Position, s string, mode mode) {
	if pr.lastPos.Line == pos.Line {
		if mode&insertSemi != 0 {
			pr.writeByte(semicolon)
		}
		if mode&noBlank == 0 {
			pr.writeByte(blank)
		}
	} else {
		pr.printLastComment()

		if mode&noExprIndent == 0 {
			pr.depth++

			pr.printComments(pos)

			pr.writeNewLine(pos)

			pr.writeIndent()

			pr.depth--
		} else {
			pr.printComments(pos)

			pr.writeNewLine(pos)

			pr.writeIndent()
		}
	}

	if mode&isLit != 0 {
		pr.writeByte(escape)
		pr.writeString(s)
		pr.writeByte(escape)
	} else {
		pr.writeString(s)
	}

	pr.lastPos = pos
}

func (pr *printer) writeString(s string) {
	pr.write([]byte(s))
}

func (pr *printer) write(bs []byte) {
	if pr.skip {
		if len(bs) > 0 {
			pr.skip = false

			bs = bs[1:]
		}
	}

	if pr.err != nil {
		return
	}
	_, pr.err = pr.w.Write(bs)
}

func (pr *printer) writeByte(b byte) {
	pr.write([]byte{b})
}

func (pr *printer) writeIndent() {
	for i := 0; i < pr.depth; i++ {
		pr.writeString(indent)
	}
}

func (pr *printer) writeNewLine(pos position.Position) {
	pr.writeByte(newline)

	switch diffLine := pos.Line - pr.lastPos.Line; {
	case diffLine <= 0:
		// panic("unexpected")
	case diffLine > 1: // truncate multiple newlines
		pr.writeByte(formfeed)
	}
}

func trimRight(s string) string {
	return strings.TrimRightFunc(s, unicode.IsSpace)
}
