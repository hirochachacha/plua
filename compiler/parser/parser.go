package parser

import (
	"fmt"

	"github.com/hirochachacha/blua/compiler/ast"
	"github.com/hirochachacha/blua/compiler/scanner"
	"github.com/hirochachacha/blua/compiler/token"
	"github.com/hirochachacha/blua/errors"
	"github.com/hirochachacha/blua/position"
)

var (
	errUnexpectedToken = errors.SyntaxError.New("expected %s, found %s")
	errIllegalVararg   = errors.SyntaxError.New("cannot use '...' outside of vararg function")
	errIllegalBreak    = errors.SyntaxError.New("cannot use 'break' outside of loop")
)

type Mode uint // currently, no mode are defined

func Parse(s *scanner.Scanner, mode Mode) (f *ast.File, err error) {
	p := &parser{
		scanner: s,
	}

	defer func() {
		if r := recover(); r != nil {
			_ = r.(bailout)

			err = p.err
		}
	}()

	p.next()

	f = p.parseFile()

	if p.err == nil {
		if serr := p.scanner.Err(); serr != nil {
			p.err = serr
		}
	}

	err = p.err

	return
}

// The parser structure holds the parser's internal state.
type parser struct {
	// scanner
	scanner *scanner.Scanner

	// Comments
	comments    []*ast.CommentGroup
	leadComment *ast.CommentGroup // last lead comment
	lineComment *ast.CommentGroup // last line comment

	// Next token
	tok token.Token

	// allow vararg in this function scope?
	allowEllipsis bool

	// allow break in this scope?
	allowBreak bool

	err error
}

type bailout struct{}

func (p *parser) markLHS(x ast.Expr) {
	// switch t := x.(type) {
	// case *ast.Name:
	// t.IsLHS = true
	// case *ast.SelectorExpr:
	// t.IsLHS = true
	// case *ast.IndexExpr:
	// t.IsLHS = true
	// }
}

func (p *parser) markRHS(x ast.Expr) {
	// switch t := x.(type) {
	// case *ast.Name:
	// if !t.IsLHS {
	// panic("name already marked as RHS")
	// }

	// t.IsLHS = false
	// case *ast.SelectorExpr:
	// if !t.IsLHS {
	// panic("selector expression already marked as RHS")
	// }

	// t.IsLHS = false
	// case *ast.IndexExpr:
	// if !t.IsLHS {
	// panic("index expression already marked as RHS")
	// }

	// t.IsLHS = false
	// }
}

// ----------------------------------------------------------------------------
// Parsing support

// Advance to the next token.
func (p *parser) next0() {
	p.tok = p.scanner.Scan()
}

// Consume a comment and return it and the line on which it ends.
func (p *parser) consumeComment() (comment *ast.Comment, endLine int) {
	endLine = p.tok.Pos.Line
	for _, r := range p.tok.Lit {
		if r == '\n' {
			endLine++
		}
	}

	if p.tok.Type == token.COMMENT {
		comment = &ast.Comment{Hyphen: p.tok.Pos, Text: p.tok.Lit}
		p.next0()

		return
	}

	comment = &ast.Comment{Hyphen: p.tok.Pos, Text: p.tok.Lit}
	p.next0()

	return
}

// Consume a group of adjacent comments, add it to the parser's
// comments list, and return it together with the line at which
// the last comment in the group ends. A non-comment token or n
// empty lines terminate a comment group.
//
func (p *parser) consumeCommentGroup(n int) (comments *ast.CommentGroup, endLine int) {
	var list []*ast.Comment
	endLine = p.tok.Pos.Line
	for p.tok.Type == token.COMMENT && p.tok.Pos.Line <= endLine+n {
		var comment *ast.Comment
		comment, endLine = p.consumeComment()
		list = append(list, comment)
	}

	// add comment group to the comments list
	comments = &ast.CommentGroup{List: list}
	p.comments = append(p.comments, comments)

	return
}

// Advance to the next non-comment token. In the process, collect
// any comment groups encountered, and remember the last lead and
// and line comments.
//
// A lead comment is a comment group that starts and ends in a
// line without any other tokens and that is followed by a non-comment
// token on the line immediately after the comment group.
//
// A line comment is a comment group that follows a non-comment
// token on the same line, and that has no tokens after it on the line
// where it ends.
//
// Lead and line comments may be considered documentation that is
// stored in the AST.
//
func (p *parser) next() {
	p.leadComment = nil
	p.lineComment = nil
	prevLine := p.tok.Pos.Line
	p.next0()

	if p.tok.Type == token.COMMENT {
		var comment *ast.CommentGroup
		var endLine int

		if p.tok.Pos.Line == prevLine {
			// The comment is on same line as the previous token; it
			// cannot be a lead comment but may be a line comment.
			comment, endLine = p.consumeCommentGroup(0)
			if p.tok.Pos.Line != endLine {
				// The next token is on a different line, thus
				// the last comment group is a line comment.
				p.lineComment = comment
			}
		}

		// consume successor comments, if any
		endLine = -1
		for p.tok.Type == token.COMMENT {
			comment, endLine = p.consumeCommentGroup(1)
		}

		if endLine+1 == p.tok.Pos.Line {
			// The next token is following on the line immediately after the
			// comment group, thus the last comment group is a lead comment.
			p.leadComment = comment
		}
	}
}

func (p *parser) error(pos position.Position, err error, args ...interface{}) {
	if serr := p.scanner.Err(); serr != nil {
		p.err = serr
	} else {
		pos.Filename = p.scanner.Filename

		var e *errors.Error
		if err, ok := err.(*errors.Error); ok {
			e = err.With(pos, args...)
		} else {
			e = errors.SyntaxError.WrapWith(pos, err)
		}

		p.err = e
	}

	panic(bailout{})
}

func (p *parser) errorExpected(actual token.Token, expected string) {
	found := "'" + actual.Type.String() + "'"
	if len(actual.Lit) > 0 {
		found += " " + actual.Lit
	}
	p.error(actual.Pos, errUnexpectedToken, expected, found)
}

func (p *parser) expect(expected token.Type) position.Position {
	if p.tok.Type != expected {
		p.errorExpected(p.tok, "'"+expected.String()+"'")

		return position.NoPos
	}

	pos := p.tok.Pos

	p.next() // make progress

	return pos
}

func (p *parser) skip() position.Position {
	pos := p.tok.Pos

	p.next() // make progress

	return pos
}

func (p *parser) accept(tok token.Type, toks ...token.Type) bool {
	if p.tok.Type == tok {
		p.next() // make progress
		return true
	}

	for _, tok := range toks {
		if p.tok.Type == tok {
			p.next() // make progress
			return true
		}
	}
	return false
}

// ----------------------------------------------------------------------------
// Name

func (p *parser) parseName() *ast.Name {
	pos := p.tok.Pos
	name := "_"
	if p.tok.Type == token.NAME {
		name = p.tok.Lit
		p.next()
	} else {
		p.expect(token.NAME) // use expect() error handling
	}
	// return &ast.Name{NamePos: pos, Name: name, IsLHS: true}
	return &ast.Name{NamePos: pos, Name: name}
}

func (p *parser) parseNameList() (list []*ast.Name) {
	list = append(list, p.parseName())
	for p.accept(token.COMMA) {
		list = append(list, p.parseName())
	}

	return
}

// ----------------------------------------------------------------------------
// Common productions

func (p *parser) parseCallExprOrLHSList() (CallExpr *ast.CallExpr, LHS []ast.Expr) {
	expr := p.parseBinaryExpr(true, token.LowestPrec)

	if expr, ok := expr.(*ast.CallExpr); ok {
		return expr, nil
	}

	LHS = append(LHS, p.checkLHS(expr))
	for p.accept(token.COMMA) {
		LHS = append(LHS, p.parseLHS())
	}

	return
}

func (p *parser) parseRHSList() (rhs []ast.Expr) {
	rhs = append(rhs, p.parseRHS())
	for p.accept(token.COMMA) {
		rhs = append(rhs, p.parseRHS())
	}

	return
}

func (p *parser) parseReturnList() (rhs []ast.Expr) {
	rhs = append(rhs, p.parseRHS())
	for p.accept(token.COMMA) {
		rhs = append(rhs, p.parseRHS())
	}

	return
}

func (p *parser) parseParameterList() (params []*ast.Name) {
	if p.tok.Type == token.ELLIPSIS {
		return
	}

	name := p.parseName()
	params = append(params, name)

	for p.accept(token.COMMA) {
		if p.tok.Type == token.ELLIPSIS {
			break
		}

		name := p.parseName()

		params = append(params, name)
	}

	return
}

func (p *parser) parseParameters() *ast.ParamList {
	var params []*ast.Name

	ellipsis := position.NoPos

	lparen := p.expect(token.LPAREN)
	if p.tok.Type != token.RPAREN {
		params = p.parseParameterList()
		if p.tok.Type == token.ELLIPSIS {
			ellipsis = p.tok.Pos
			p.next()
		}
	}
	rparen := p.expect(token.RPAREN)

	return &ast.ParamList{Lparen: lparen, List: params, Ellipsis: ellipsis, Rparen: rparen}
}

// ----------------------------------------------------------------------------
// Blocks

func (p *parser) isEndOfBlock() bool {
	switch p.tok.Type {
	case token.EOF, token.ELSE, token.ELSEIF, token.END:
		return true
	case token.UNTIL:
		return true
	}

	return false
}

func (p *parser) parseStmtList() []ast.Stmt {
	var list []ast.Stmt

loop:
	for {
		switch p.tok.Type {
		case token.EOF, token.ELSE, token.ELSEIF, token.END, token.UNTIL:
			break loop
		case token.RETURN:
			list = append(list, p.parseReturnStmt())
			break loop
		}
		list = append(list, p.parseStmt())
	}

	return list
}

func (p *parser) parseChunk() []ast.Stmt {
	p.allowEllipsis = true

	return p.parseStmtList()
}

func (p *parser) parseThenBlock() *ast.Block {
	opening := p.expect(token.THEN)

	list := p.parseStmtList()

	var closing position.Position

	if p.tok.Type == token.ELSEIF || p.tok.Type == token.ELSE {
		closing = p.tok.Pos
	} else {
		closing = p.expect(token.END).OffsetColumn(3)
	}

	return &ast.Block{
		Opening: opening,
		List:    list,
		Closing: closing,
	}
}

func (p *parser) parseRepeatBlock() *ast.Block {
	opening := p.skip()

	list := p.parseStmtList()

	closing := p.expect(token.UNTIL).OffsetColumn(5)

	return &ast.Block{
		Opening: opening,
		List:    list,
		Closing: closing,
	}
}

func (p *parser) parseBlock() *ast.Block {
	opening := p.skip()

	list := p.parseStmtList()

	closing := p.expect(token.END).OffsetColumn(3)

	return &ast.Block{
		Opening: opening,
		List:    list,
		Closing: closing,
	}
}

func (p *parser) parseBody(pos position.Position) *ast.Block {
	list := p.parseStmtList()

	closing := p.expect(token.END).OffsetColumn(3)

	return &ast.Block{
		Opening: pos,
		List:    list,
		Closing: closing,
	}
}

func (p *parser) parseFuncBody() *ast.FuncBody {
	params := p.parseParameters()

	old := p.allowEllipsis
	p.allowEllipsis = params.Ellipsis.IsValid()

	body := p.parseBody(params.End())

	p.allowEllipsis = old

	return &ast.FuncBody{
		Params: params,
		Body:   body,
	}
}

// ----------------------------------------------------------------------------
// Expressions

func (p *parser) parseSelector(x ast.Expr) ast.Expr {
	period := p.skip()

	sel := p.parseName()

	return &ast.SelectorExpr{X: x, Period: period, Sel: sel}
}

func (p *parser) parseIndex(x ast.Expr) ast.Expr {
	lbrack := p.skip()

	index := p.parseRHS()

	rbrack := p.expect(token.RBRACK)

	return &ast.IndexExpr{X: x, Lbrack: lbrack, Index: index, Rbrack: rbrack}
}

func (p *parser) parseCall(isMethod bool, x ast.Expr) ast.Expr {
	colon := position.NoPos
	var name *ast.Name

	if isMethod {
		colon = p.skip()
		name = p.parseName()
	}

	tok := p.tok

	switch tok.Type {
	case token.LPAREN:
		lparen := tok.Pos
		p.next()
		var args []ast.Expr

		for p.tok.Type != token.RPAREN && p.tok.Type != token.EOF {
			args = append(args, p.parseRHS())
			if !p.accept(token.COMMA) {
				break
			}
		}

		rparen := p.expect(token.RPAREN)

		return &ast.CallExpr{X: x, Colon: colon, Name: name, Lparen: lparen, Args: args, Rparen: rparen}
	case token.LBRACE:
		y := p.parseTableLit()

		return &ast.CallExpr{X: x, Colon: colon, Name: name, Args: []ast.Expr{y}}
	case token.STRING:
		y := &ast.BasicLit{Token: tok}

		p.next()

		return &ast.CallExpr{X: x, Colon: colon, Name: name, Args: []ast.Expr{y}}
	}

	// we have an error
	p.errorExpected(tok, "callable")

	// syncStmt(p)

	return &ast.BadExpr{From: tok.Pos, To: p.tok.Pos}
}

func (p *parser) parseElement() ast.Expr {
	lbrack := position.NoPos
	rbrack := position.NoPos

	var x ast.Expr

	switch p.tok.Type {
	case token.LBRACK: // [key]
		lbrack = p.tok.Pos

		p.next()

		x = p.parseRHS()

		rbrack = p.expect(token.RBRACK)
	default: // key
		x = p.parseBinaryExpr(true, token.LowestPrec)
	}

	if p.tok.Type == token.ASSIGN {
		// key = value
		p.next()

		y := p.parseRHS()

		return &ast.KeyValueExpr{Lbrack: lbrack, Key: x, Rbrack: rbrack, Equal: p.tok.Pos, Value: y}
	}

	// value
	// p.markRHS(x)

	return x
}

func (p *parser) parseElementList() (fields []ast.Expr) {
	for p.tok.Type != token.RBRACE && p.tok.Type != token.EOF {
		e := p.parseElement()

		fields = append(fields, e)

		if !p.accept(token.COMMA, token.SEMICOLON) {
			break
		}
	}

	return
}

func (p *parser) parseTableLit() *ast.TableLit {
	lbrace := p.expect(token.LBRACE)

	var fields []ast.Expr
	if p.tok.Type != token.RBRACE {
		fields = p.parseElementList()
	}

	rbrace := p.expect(token.RBRACE)

	return &ast.TableLit{Lbrace: lbrace, Fields: fields, Rbrace: rbrace}
}

func (p *parser) parseFuncLit() *ast.FuncLit {
	pos := p.expect(token.FUNCTION)

	f := p.parseFuncBody()

	return &ast.FuncLit{
		Func: pos,
		Body: f,
	}
}

// checkExpr checks that x is an expression.
func (p *parser) checkExpr(x ast.Expr) ast.Expr {
	switch unparen(x).(type) {
	case *ast.BadExpr:
	case *ast.Name:
	case *ast.Vararg:
	case *ast.BasicLit:
	case *ast.FuncLit:
	case *ast.TableLit:
	case *ast.ParenExpr:
		panic("unreachable")
	case *ast.SelectorExpr:
	case *ast.IndexExpr:
	case *ast.CallExpr:
	case *ast.UnaryExpr:
	case *ast.BinaryExpr:
	default:
		// all other nodes are not proper expressions
		p.error(x.Pos(), errUnexpectedToken, "expression", fmt.Sprintf("%T", x))
		x = &ast.BadExpr{From: x.Pos(), To: x.End()}
	}
	return x
}

// checkLHS checks that x is an LHS expression.
func (p *parser) checkLHS(x ast.Expr) ast.Expr {
	// switch t := x.(type) {
	switch x.(type) {
	case *ast.Name:
		// if !t.IsLHS {
		// goto error
		// }
	case *ast.SelectorExpr:
		// if !t.IsLHS {
		// goto error
		// }
	case *ast.IndexExpr:
		// if !t.IsLHS {
		// goto error
		// }
	default:
		goto error
	}

	return x

error:
	// all other nodes are not proper expressions
	p.error(x.Pos(), errUnexpectedToken, "LHS", fmt.Sprintf("%T", x))

	x = &ast.BadExpr{From: x.Pos(), To: x.End()}

	return x
}

// If x is of the form (T), unparen returns unparen(T), otherwise it returns x.
func unparen(x ast.Expr) ast.Expr {
	if p, isParen := x.(*ast.ParenExpr); isParen {
		x = unparen(p.X)
	}
	return x
}

func (p *parser) parsePrimaryExpr() ast.Expr {
	tok := p.tok

	switch tok.Type {
	case token.NAME:
		x := p.parseName()
		// x.IsLHS = false
		return x
	case token.LPAREN:
		lparen := tok.Pos
		p.next()
		x := p.parseRHS()
		rparen := p.expect(token.RPAREN)

		return &ast.ParenExpr{Lparen: lparen, X: x, Rparen: rparen}
	}

	// we have an error
	p.errorExpected(tok, "NAME or '('")

	// syncStmt(p)

	return &ast.BadExpr{From: tok.Pos, To: p.tok.Pos}
}

func (p *parser) parseSuffixedExpr(isLHS bool) ast.Expr {
	x := p.parsePrimaryExpr()
L:
	for {
		switch p.tok.Type {
		case token.PERIOD:
			x = p.parseSelector(p.checkExpr(x))
		case token.LBRACK:
			x = p.parseIndex(p.checkExpr(x))
		case token.LPAREN, token.LBRACE, token.STRING:
			x = p.parseCall(false, p.checkExpr(x))
		case token.COLON:
			x = p.parseCall(true, p.checkExpr(x))
		default:
			if isLHS {
				// p.markLHS(x)
			}

			break L
		}
	}

	return x
}

func (p *parser) parseSimpleExpr(isLHS bool) ast.Expr {
	tok := p.tok

	switch tok.Type {
	case token.ELLIPSIS:
		var x ast.Expr

		if !p.allowEllipsis {
			p.error(tok.Pos, errIllegalVararg)

			p.next()

			x = &ast.BadExpr{From: tok.Pos, To: p.tok.Pos}

			return x
		}

		x = &ast.Vararg{Ellipsis: tok.Pos}

		p.next()

		return x

	case token.INT, token.FLOAT, token.STRING:
		x := &ast.BasicLit{Token: tok}
		p.next()
		return x

	case token.FALSE, token.TRUE, token.NIL:
		x := &ast.BasicLit{Token: tok}
		p.next()
		return x

	case token.FUNCTION:
		return p.parseFuncLit()

	case token.LBRACE:
		return p.parseTableLit()
	default:
		return p.parseSuffixedExpr(isLHS)
	}
}

// If isLHS is set and the result is an name, it is not resolved.
func (p *parser) parseUnaryExpr(isLHS bool) ast.Expr {
	tok := p.tok

	switch tok.Type {
	case token.UNM, token.BNOT, token.NOT, token.LEN:
		p.next()
		x := p.parseBinaryExpr(false, token.UnaryPrec)
		return &ast.UnaryExpr{OpPos: tok.Pos, Op: tok.Type, X: p.checkExpr(x)}
	}

	return p.parseSimpleExpr(isLHS)
}

func (p *parser) tokPrec() (token.Type, int, int) {
	tok := p.tok.Type
	l, r := tok.Precedence()
	return tok, l, r
}

// If isLHS is set and the result is an name, it is not resolved.
func (p *parser) parseBinaryExpr(isLHS bool, prec1 int) ast.Expr {
	x := p.parseUnaryExpr(isLHS)

	for {
		op, lprec, rprec := p.tokPrec()

		if lprec <= prec1 {
			break
		}

		if isLHS {
			// p.markRHS(x)

			isLHS = false
		}

		olprec, orprec := lprec, rprec

		for {
			pos := p.expect(op)

			y := p.parseBinaryExpr(false, orprec)

			x = &ast.BinaryExpr{X: p.checkExpr(x), OpPos: pos, Op: op, Y: p.checkExpr(y)}

			op, olprec, orprec = p.tokPrec()

			if olprec < rprec {
				break
			}
		}
	}

	return x
}

func (p *parser) parseLHS() ast.Expr {
	return p.checkLHS(p.parseBinaryExpr(true, token.LowestPrec))
}

func (p *parser) parseRHS() ast.Expr {
	return p.checkExpr(p.parseBinaryExpr(false, token.LowestPrec))
}

// ----------------------------------------------------------------------------
// Statements

func (p *parser) parseAssign(LHS []ast.Expr) ast.Stmt {
	pos := p.skip()

	rhs := p.parseRHSList()

	assign := &ast.AssignStmt{
		LHS:   LHS,
		Equal: pos,
		RHS:   rhs,
	}

	return assign
}

func (p *parser) parseLocalAssignStmt(pos position.Position) ast.Stmt {
	LHS := p.parseNameList()

	var stmt ast.Stmt

	if p.tok.Type == token.ASSIGN {
		eq := p.tok.Pos
		p.next()

		rhs := p.parseRHSList()

		stmt = &ast.LocalAssignStmt{
			Local: pos,
			LHS:   LHS,
			Equal: eq,
			RHS:   rhs,
		}
	} else {
		stmt = &ast.LocalAssignStmt{
			Local: pos,
			LHS:   LHS,
		}
	}

	return stmt
}

func (p *parser) parseLocalStmt() ast.Stmt {
	pos := p.skip()

	if p.tok.Type == token.FUNCTION {
		pos := p.skip()

		name := p.parseName()

		f := p.parseFuncBody()

		return &ast.LocalFuncStmt{
			Local: pos,
			Func:  pos,
			Name:  name,
			Body:  f,
		}
	}

	return p.parseLocalAssignStmt(pos)
}

func (p *parser) parseFuncStmt() *ast.FuncStmt {
	pos := p.skip()

	var prefix []*ast.Name
	var accessTok token.Type
	var accessPos position.Position
	var name *ast.Name

	var list []*ast.Name
	var periodPos position.Position

	list = append(list, p.parseName())

	for p.tok.Type == token.PERIOD {
		periodPos = p.tok.Pos

		p.next() // make progress

		list = append(list, p.parseName())
	}

	if p.tok.Type == token.COLON {
		accessTok = token.COLON
		accessPos = p.tok.Pos

		p.next()

		name = p.parseName()
		prefix = list
	} else if len(list) == 1 {
		name = list[0]
		accessTok = token.ILLEGAL
		accessPos = position.NoPos
	} else {
		name = list[len(list)-1]
		prefix = list[:len(list)-1]
		accessTok = token.PERIOD
		accessPos = periodPos
	}

	f := p.parseFuncBody()

	fn := &ast.FuncStmt{
		Func:       pos,
		NamePrefix: prefix,
		AccessTok:  accessTok,
		AccessPos:  accessPos,
		Name:       name,
		Body:       f,
	}

	return fn
}

func (p *parser) parseExprOrAssignStmt() (stmt ast.Stmt) {
	tok := p.tok

	expr, LHS := p.parseCallExprOrLHSList()

	if p.tok.Type == token.ASSIGN { // assign stmt
		stmt = p.parseAssign(LHS)
	} else { // expr stmt
		if expr == nil {
			p.errorExpected(tok, "callable")

			return &ast.BadStmt{From: tok.Pos, To: p.tok.Pos}
		}

		stmt = &ast.ExprStmt{expr}
	}

	return
}

func (p *parser) parseReturnStmt() *ast.ReturnStmt {
	pos := p.skip()
	semi := position.NoPos

	if p.isEndOfBlock() {
		return &ast.ReturnStmt{
			Return:    pos,
			Semicolon: semi,
		}
	}

	if p.tok.Type == token.SEMICOLON {
		semi = p.tok.Pos

		p.next()

		return &ast.ReturnStmt{
			Return:    pos,
			Semicolon: semi,
		}
	}

	results := p.parseReturnList()

	if p.tok.Type == token.SEMICOLON {
		semi = p.tok.Pos

		p.next()
	}

	return &ast.ReturnStmt{
		Return:    pos,
		Results:   results,
		Semicolon: semi,
	}
}

func (p *parser) parseLabelStmt() *ast.LabelStmt {
	pos := p.skip()

	name := p.parseName()

	end := p.expect(token.LABEL)

	label := &ast.LabelStmt{
		Label:    pos,
		Name:     name,
		EndLabel: end,
	}

	return label
}

func (p *parser) parseGotoStmt() *ast.GotoStmt {
	pos := p.skip()

	label := p.parseName()

	return &ast.GotoStmt{Goto: pos, Label: label}
}

func (p *parser) parseBreak() ast.Stmt {
	if !p.allowBreak {
		pos := p.tok.Pos
		p.error(pos, errIllegalBreak)

		// syncStmt(p)

		return &ast.BadStmt{From: pos, To: p.tok.Pos}
	}

	pos := p.skip()

	return &ast.BreakStmt{Break: pos}
}

func (p *parser) parseDoStmt() *ast.DoStmt {
	body := p.parseBlock()

	return &ast.DoStmt{
		Body: body,
	}
}

func (p *parser) parseRepeatStmt() *ast.RepeatStmt {
	old := p.allowBreak

	p.allowBreak = true

	body := p.parseRepeatBlock()

	p.allowBreak = old

	cond := p.parseRHS()

	return &ast.RepeatStmt{
		Body: body,
		Cond: cond,
	}
}

func (p *parser) parseWhileStmt() *ast.WhileStmt {
	pos := p.skip()

	cond := p.parseRHS()

	old := p.allowBreak

	p.allowBreak = true

	body := p.parseBlock()

	p.allowBreak = old

	return &ast.WhileStmt{
		While: pos,
		Cond:  cond,
		Body:  body,
	}
}

func (p *parser) parseIfStmt() *ast.IfStmt {
	pos := p.skip()

	cond := p.parseRHS()

	body := p.parseThenBlock()

	ifRoot := &ast.IfStmt{
		If:   pos,
		Cond: cond,
		Body: body,
	}

	last := ifRoot
	for {
		pos := p.tok.Pos
		if !p.accept(token.ELSEIF) {
			break
		}

		cond := p.parseRHS()
		body := p.parseThenBlock()

		ifLeaf := &ast.IfStmt{
			If:   pos,
			Cond: cond,
			Body: body,
		}

		last.Else = &ast.Block{
			Opening: ifLeaf.Pos(),
			List:    []ast.Stmt{ifLeaf},
			Closing: ifLeaf.End(),
		}

		last = ifLeaf
	}

	if p.tok.Type == token.ELSE {
		last.Else = p.parseBlock()
	}

	return ifRoot
}

func (p *parser) parseForStmt() ast.Stmt {
	pos := p.skip()

	names := p.parseNameList()

	var in, eq position.Position

	if len(names) > 1 {
		p.expect(token.IN)
		goto foreach
	}

	switch p.tok.Type {
	case token.ASSIGN:
		eq = p.tok.Pos
		p.next()
	case token.IN:
		p.next()
		goto foreach
	default:
		tok := p.tok

		p.errorExpected(tok, "'=' or 'in'")

		// syncStmt(p)

		return &ast.BadStmt{From: tok.Pos, To: p.tok.Pos}
	}

	// fornum
	{

		start := p.parseRHS()

		p.expect(token.COMMA)

		finish := p.parseRHS()

		var step ast.Expr

		if p.accept(token.COMMA) {
			step = p.parseRHS()
		}

		old := p.allowBreak

		p.allowBreak = true

		body := p.parseBlock()

		p.allowBreak = old

		return &ast.ForStmt{
			For:    pos,
			Name:   names[0],
			Equal:  eq,
			Start:  start,
			Finish: finish,
			Step:   step,
			Body:   body,
		}
	}

foreach:
	{
		exprs := p.parseRHSList()

		old := p.allowBreak

		p.allowBreak = true

		body := p.parseBlock()

		p.allowBreak = old

		return &ast.ForEachStmt{
			For:   pos,
			Names: names,
			In:    in,
			Exprs: exprs,
			Body:  body,
		}
	}
}

func (p *parser) parseStmt() (s ast.Stmt) {
	switch p.tok.Type {
	case token.IF:
		s = p.parseIfStmt()
	case token.WHILE:
		s = p.parseWhileStmt()
	case token.DO:
		s = p.parseDoStmt()
	case token.FOR:
		s = p.parseForStmt()
	case token.REPEAT:
		s = p.parseRepeatStmt()
	case token.FUNCTION:
		s = p.parseFuncStmt()
	case token.LOCAL:
		s = p.parseLocalStmt()
	case token.SEMICOLON:
		s = &ast.EmptyStmt{Semicolon: p.tok.Pos}

		p.next()
	case token.LABEL:
		s = p.parseLabelStmt()
	case token.RETURN:
		s = p.parseReturnStmt()
	case token.GOTO:
		s = p.parseGotoStmt()
	case token.BREAK:
		s = p.parseBreak()
	default:
		s = p.parseExprOrAssignStmt()
	}

	return
}

// ----------------------------------------------------------------------------
// Source files

func (p *parser) parseFile() *ast.File {
	chunk := p.parseChunk()

	return &ast.File{
		Filename: p.scanner.Filename,
		Shebang:  p.scanner.Shebang,
		Chunk:    chunk,
		Comments: p.comments,
	}
}
