// Original: src/go/ast/ast.go
//
// Copyright 2009 The Go Authors. All rights reserved.
// Portions Copyright 2016 Hiroshi Ioka. All rights reserved.
//
// Redistribution and use in source and binary forms, with or without
// modification, are permitted provided that the following conditions are
// met:
//
//    * Redistributions of source code must retain the above copyright
// notice, this list of conditions and the following disclaimer.
//    * Redistributions in binary form must reproduce the above
// copyright notice, this list of conditions and the following disclaimer
// in the documentation and/or other materials provided with the
// distribution.
//    * Neither the name of Google Inc. nor the names of its
// contributors may be used to endorse or promote products derived from
// this software without specific prior written permission.
//
// THIS SOFTWARE IS PROVIDED BY THE COPYRIGHT HOLDERS AND CONTRIBUTORS
// "AS IS" AND ANY EXPRESS OR IMPLIED WARRANTIES, INCLUDING, BUT NOT
// LIMITED TO, THE IMPLIED WARRANTIES OF MERCHANTABILITY AND FITNESS FOR
// A PARTICULAR PURPOSE ARE DISCLAIMED. IN NO EVENT SHALL THE COPYRIGHT
// OWNER OR CONTRIBUTORS BE LIABLE FOR ANY DIRECT, INDIRECT, INCIDENTAL,
// SPECIAL, EXEMPLARY, OR CONSEQUENTIAL DAMAGES (INCLUDING, BUT NOT
// LIMITED TO, PROCUREMENT OF SUBSTITUTE GOODS OR SERVICES; LOSS OF USE,
// DATA, OR PROFITS; OR BUSINESS INTERRUPTION) HOWEVER CAUSED AND ON ANY
// THEORY OF LIABILITY, WHETHER IN CONTRACT, STRICT LIABILITY, OR TORT
// (INCLUDING NEGLIGENCE OR OTHERWISE) ARISING IN ANY WAY OUT OF THE USE
// OF THIS SOFTWARE, EVEN IF ADVISED OF THE POSSIBILITY OF SUCH DAMAGE.

package ast

import (
	"strings"

	"github.com/hirochachacha/plua/compiler/token"
	"github.com/hirochachacha/plua/position"
)

// ----------------------------------------------------------------------------
// Interfaces
//

// All node types implement the Node interface.
type Node interface {
	Type() Type
	Pos() position.Position // position of first character belonging to the node
	End() position.Position // position of first character immediately after the node
}

// All expression nodes implement the Expr interface.
type Expr interface {
	Node

	exprNode()
}

// All statement nodes implement the Stmt interface.
type Stmt interface {
	Node

	stmtNode()
}

// ----------------------------------------------------------------------------
// Expressions

// An expression is represented by a tree consisting of one
// or more of the following concrete expression nodes.
//
type (
	// A BadExpr node is a placeholder for expressions containing
	// syntax errors for which no correct expression nodes can be
	// created.
	//
	BadExpr struct {
		From, To position.Position // position range of bad expression

		// Parent Node
	}

	// An Name node represents an identifier.
	Name struct {
		NamePos position.Position // identifier position
		Name    string            // identifier name

		// Version int
		// IsLHS   bool
		// Parent  Node
	}

	Vararg struct {
		Ellipsis position.Position // position of "..."

		// Parent Node
	}

	// A BasicLit node represents a literal of basic type.
	BasicLit struct {
		token.Token

		// Parent Node
	}

	// A FuncLit node represents a function literal.
	FuncLit struct {
		Func position.Position // position of "function" keyword
		Body *FuncBody         // function body

		// Parent Node
	}

	// A TableLit node represents a composite literal.
	TableLit struct {
		Lbrace position.Position // position of "{"
		Fields []Expr            // fields
		Rbrace position.Position // position of "}"

		// Parent Node
	}

	// A ParenExpr node represents a parenthesized expression.
	ParenExpr struct {
		Lparen position.Position // position of "("
		X      Expr              // parenthesized expression
		Rparen position.Position // position of ")"

		// Parent Node
	}

	// A SelectorExpr node represents an expression followed by a selector.
	SelectorExpr struct {
		X      Expr              // expression
		Period position.Position // position of "."
		Sel    *Name             // field selector

		// IsLHS  bool
		// Parent Node
	}

	// An IndexExpr node represents an expression followed by an index.
	IndexExpr struct {
		X      Expr              // expression
		Lbrack position.Position // position of "["
		Index  Expr              // index expression
		Rbrack position.Position // position of "]"

		// IsLHS  bool
		// Parent Node
	}

	// A CallExpr node represents an expression followed by an argument list.
	CallExpr struct {
		X      Expr              // function expression
		Colon  position.Position // position of ":" or position.NoPos
		Name   *Name             // method name; or nil
		Lparen position.Position // position of "(" or position.NoPos
		Args   []Expr            // function arguments; or nil
		Rparen position.Position // position of ")" or position.NoPos

		// Parent Node
	}

	// A UnaryExpr node represents a unary expression.
	UnaryExpr struct {
		OpPos position.Position // position of Op
		Op    token.Type        // operator
		X     Expr              // operand

		// Parent Node
	}

	// A BinaryExpr node represents a binary expression.
	BinaryExpr struct {
		X     Expr              // left operand
		OpPos position.Position // position of Op
		Op    token.Type        // operator
		Y     Expr              // right operand

		// Parent Node
	}

	// A KeyValueExpr node represents (key = value) pairs or (value)
	// in table literals.
	//
	KeyValueExpr struct {
		Lbrack position.Position // position of "["; or position.NoPos
		Key    Expr              // key; or nil
		Rbrack position.Position // position of "]"; or position.NoPos
		Equal  position.Position // position of "="; or position.NoPos
		Value  Expr

		// Parent Node
	}

	// z = Φ(x, y)
	// Phi struct {
	// Name string
	// X    int
	// Y    int
	// Z    int
	// }
)

func (*BadExpr) Type() Type      { return BAD_EXPR }
func (*Name) Type() Type         { return NAME }
func (*Vararg) Type() Type       { return VARARG }
func (*BasicLit) Type() Type     { return BASIC_LIT }
func (*FuncLit) Type() Type      { return FUNC_LIT }
func (*TableLit) Type() Type     { return TABLE_LIT }
func (*ParenExpr) Type() Type    { return PAREN_EXPR }
func (*SelectorExpr) Type() Type { return SELECTOR_EXPR }
func (*IndexExpr) Type() Type    { return INDEX_EXPR }
func (*CallExpr) Type() Type     { return CALL_EXPR }
func (*UnaryExpr) Type() Type    { return UNARY_EXPR }
func (*BinaryExpr) Type() Type   { return BINARY_EXPR }
func (*KeyValueExpr) Type() Type { return KEY_VALUE_EXPR }

// Pos and End implementations for expression/type nodes.
//
func (x *BadExpr) Pos() position.Position  { return x.From }
func (x *Name) Pos() position.Position     { return x.NamePos }
func (x *Vararg) Pos() position.Position   { return x.Ellipsis }
func (x *BasicLit) Pos() position.Position { return x.Token.Pos }
func (x *FuncLit) Pos() position.Position  { return x.Func }
func (x *TableLit) Pos() position.Position {
	return x.Lbrace
}
func (x *ParenExpr) Pos() position.Position    { return x.Lparen }
func (x *SelectorExpr) Pos() position.Position { return x.X.Pos() }
func (x *IndexExpr) Pos() position.Position    { return x.X.Pos() }
func (x *CallExpr) Pos() position.Position     { return x.X.Pos() }
func (x *UnaryExpr) Pos() position.Position    { return x.OpPos }
func (x *BinaryExpr) Pos() position.Position   { return x.X.Pos() }
func (x *KeyValueExpr) Pos() position.Position {
	if x.Lbrack.IsValid() {
		return x.Lbrack
	}

	return x.Key.Pos()
}

func (x *BadExpr) End() position.Position      { return x.To }
func (x *Name) End() position.Position         { return x.NamePos.Offset(x.Name) }
func (x *Vararg) End() position.Position       { return x.Ellipsis.OffsetColumn(3) }
func (x *BasicLit) End() position.Position     { return x.Token.Pos.Offset(x.Token.Lit) }
func (x *FuncLit) End() position.Position      { return x.Body.End() }
func (x *TableLit) End() position.Position     { return x.Rbrace.OffsetColumn(1) }
func (x *ParenExpr) End() position.Position    { return x.Rparen.OffsetColumn(1) }
func (x *SelectorExpr) End() position.Position { return x.Sel.End() }
func (x *IndexExpr) End() position.Position    { return x.Rbrack.OffsetColumn(1) }
func (x *CallExpr) End() position.Position     { return x.Rparen.OffsetColumn(1) }
func (x *UnaryExpr) End() position.Position    { return x.X.End() }
func (x *BinaryExpr) End() position.Position   { return x.Y.End() }
func (x *KeyValueExpr) End() position.Position { return x.Value.End() }

// exprNode() ensures that only expression/type nodes can be
// assigned to an ExprNode.
//
func (*BadExpr) exprNode()      {}
func (*Name) exprNode()         {}
func (*Vararg) exprNode()       {}
func (*BasicLit) exprNode()     {}
func (*FuncLit) exprNode()      {}
func (*TableLit) exprNode()     {}
func (*ParenExpr) exprNode()    {}
func (*SelectorExpr) exprNode() {}
func (*IndexExpr) exprNode()    {}
func (*CallExpr) exprNode()     {}
func (*UnaryExpr) exprNode()    {}
func (*BinaryExpr) exprNode()   {}
func (*KeyValueExpr) exprNode() {}

// ----------------------------------------------------------------------------
// Statements

// A statement is represented by a tree consisting of one
// or more of the following concrete statement nodes.
//
type (
	// A BadStmt node is a placeholder for statements containing
	// syntax errors for which no correct statement nodes can be
	// created.
	//
	BadStmt struct {
		From, To position.Position // position range of bad statement
	}

	// An EmptyStmt node represents an empty statement.
	// The "position" of the empty statement is the position
	// of the immediately preceding semicolon.
	//
	EmptyStmt struct {
		Semicolon position.Position // position of preceding ";"
	}

	LocalAssignStmt struct {
		Local position.Position // position of "local" keyword
		LHS   []*Name
		Equal position.Position // position of Tok
		RHS   []Expr
	}

	LocalFuncStmt struct {
		Local position.Position // position of "local" keyword
		Func  position.Position // position of "function" keyword
		Name  *Name
		Body  *FuncBody // function body
	}

	// A FuncStmt node represents a function statement.
	FuncStmt struct {
		Func       position.Position // position of "function" keyword
		NamePrefix []*Name
		AccessTok  token.Type        // "." or ":"
		AccessPos  position.Position // position of AccessTok
		Name       *Name
		Body       *FuncBody // function body
	}

	// A LabelStmt node represents a label statement.
	LabelStmt struct {
		Label    position.Position // position of "::"
		Name     *Name
		EndLabel position.Position // position of "::"

		// Postlude []*Phi
	}

	// An ExprStmt node represents a (stand-alone) expression
	// in a statement list.
	//
	ExprStmt struct {
		X *CallExpr // expression
	}

	// An AssignStmt node represents an assignment statement.
	AssignStmt struct {
		LHS   []Expr
		Equal position.Position // position of "="
		RHS   []Expr
	}

	// A GotoStmt node represents a goto statement.
	GotoStmt struct {
		Goto  position.Position
		Label *Name // label name
	}

	// A BreakStmt represents a break statement.
	BreakStmt struct {
		Break position.Position // position of "break"
	}

	// An IfStmt node represents an if statement.
	IfStmt struct {
		If   position.Position // position of "if" or "elseif" keyword
		Cond Expr              // condition
		Body *Block
		Else *Block // else branch; or nil

		// Postlude []*Phi
	}

	// A DoStmt represents a do statement.
	DoStmt struct {
		Body *Block

		// Postlude []*Phi
	}

	// A WhileStmt represents a while statement.
	WhileStmt struct {
		While position.Position // position of "while" keyword
		Cond  Expr              // condition
		Body  *Block

		// Prelude  []*Phi
		// Postlude []*Phi
	}

	// A RepeatStmt represents a repeat statement.
	RepeatStmt struct {
		Repeat position.Position // position of "repeat" keyword
		Body   *Block
		Until  position.Position // position of  "until" keyword
		Cond   Expr              // condition

		// Prelude  []*Phi
		// Postlude []*Phi
	}

	// A ReturnStmt represents a return statement.
	ReturnStmt struct {
		Return    position.Position // position of "return" keyword
		Results   []Expr
		Semicolon position.Position // position of ";" if exist
	}

	// A ForStmt represents a for statement.
	ForStmt struct {
		For    position.Position // position of "for" keyword
		Name   *Name
		Equal  position.Position // position of "="
		Start  Expr
		Finish Expr
		Step   Expr // or nil
		Body   *Block

		// Prelude  []*Phi
		// Postlude []*Phi
	}

	// A ForEachStmt represents a for statement with a range clause.
	ForEachStmt struct {
		For   position.Position // position of "for" keyword
		Names []*Name           // Key, Value may be nil
		In    position.Position // position of "in" keyword
		Exprs []Expr            // value to range over
		Body  *Block

		// Prelude  []*Phi
		// Postlude []*Phi
	}
)

func (*BadStmt) Type() Type         { return BAD_STMT }
func (*EmptyStmt) Type() Type       { return EMPTY_STMT }
func (*LocalAssignStmt) Type() Type { return LOCAL_ASSIGN_STMT }
func (*LocalFuncStmt) Type() Type   { return LOCAL_FUNC_STMT }
func (*FuncStmt) Type() Type        { return FUNC_STMT }
func (*LabelStmt) Type() Type       { return LABEL_STMT }
func (*ExprStmt) Type() Type        { return EXPR_STMT }
func (*AssignStmt) Type() Type      { return ASSIGN_STMT }
func (*GotoStmt) Type() Type        { return GOTO_STMT }
func (*IfStmt) Type() Type          { return IF_STMT }
func (*DoStmt) Type() Type          { return DO_STMT }
func (*WhileStmt) Type() Type       { return WHILE_STMT }
func (*RepeatStmt) Type() Type      { return REPEAT_STMT }
func (*BreakStmt) Type() Type       { return BREAK_STMT }
func (*ReturnStmt) Type() Type      { return RETURN_STMT }
func (*ForStmt) Type() Type         { return FOR_STMT }
func (*ForEachStmt) Type() Type     { return FOR_EACH_STMT }

// Pos and End implementations for statement nodes.
//
func (s *BadStmt) Pos() position.Position         { return s.From }
func (s *EmptyStmt) Pos() position.Position       { return s.Semicolon }
func (s *LocalAssignStmt) Pos() position.Position { return s.Local }
func (s *LocalFuncStmt) Pos() position.Position   { return s.Local }
func (s *FuncStmt) Pos() position.Position        { return s.Func }
func (s *LabelStmt) Pos() position.Position       { return s.Label }
func (s *ExprStmt) Pos() position.Position        { return s.X.Pos() }
func (s *AssignStmt) Pos() position.Position      { return s.LHS[0].Pos() }
func (s *GotoStmt) Pos() position.Position        { return s.Goto }
func (s *IfStmt) Pos() position.Position          { return s.If }
func (s *DoStmt) Pos() position.Position          { return s.Body.Pos() }
func (s *WhileStmt) Pos() position.Position       { return s.While }
func (s *RepeatStmt) Pos() position.Position      { return s.Repeat }
func (s *BreakStmt) Pos() position.Position       { return s.Break }
func (s *ReturnStmt) Pos() position.Position      { return s.Return }
func (s *ForStmt) Pos() position.Position         { return s.For }
func (s *ForEachStmt) Pos() position.Position     { return s.For }

func (s *BadStmt) End() position.Position         { return s.To }
func (s *EmptyStmt) End() position.Position       { return s.Semicolon.OffsetColumn(1) }
func (s *LocalAssignStmt) End() position.Position { return s.RHS[len(s.RHS)-1].End() }
func (s *LocalFuncStmt) End() position.Position   { return s.Body.End() }
func (s *FuncStmt) End() position.Position        { return s.Body.End() }
func (s *LabelStmt) End() position.Position       { return s.EndLabel.OffsetColumn(2) }
func (s *ExprStmt) End() position.Position        { return s.X.End() }
func (s *AssignStmt) End() position.Position      { return s.RHS[len(s.RHS)-1].End() }
func (s *GotoStmt) End() position.Position {
	return s.Label.End()
}
func (s *IfStmt) End() position.Position {
	if s.Else != nil {
		return s.Else.End()
	}
	return s.Body.End()
}
func (s *DoStmt) End() position.Position     { return s.Body.End() }
func (s *WhileStmt) End() position.Position  { return s.Body.End() }
func (s *RepeatStmt) End() position.Position { return s.Cond.End() }
func (s *BreakStmt) End() position.Position  { return s.Break.OffsetColumn(5) }
func (s *ReturnStmt) End() position.Position {
	if s.Semicolon.IsValid() {
		return s.Semicolon.OffsetColumn(1)
	}
	if s.Results != nil {
		return s.Results[len(s.Results)-1].End()
	}
	return s.Return.OffsetColumn(6)
}
func (s *ForStmt) End() position.Position     { return s.Body.End() }
func (s *ForEachStmt) End() position.Position { return s.Body.End() }

// stmtNode() ensures that only statement nodes can be
// assigned to a StmtNode.
//
func (*BadStmt) stmtNode()         {}
func (*EmptyStmt) stmtNode()       {}
func (*LocalAssignStmt) stmtNode() {}
func (*LocalFuncStmt) stmtNode()   {}
func (*FuncStmt) stmtNode()        {}
func (*LabelStmt) stmtNode()       {}
func (*ExprStmt) stmtNode()        {}
func (*AssignStmt) stmtNode()      {}
func (*GotoStmt) stmtNode()        {}
func (*IfStmt) stmtNode()          {}
func (*DoStmt) stmtNode()          {}
func (*WhileStmt) stmtNode()       {}
func (*RepeatStmt) stmtNode()      {}
func (*BreakStmt) stmtNode()       {}
func (*ReturnStmt) stmtNode()      {}
func (*ForStmt) stmtNode()         {}
func (*ForEachStmt) stmtNode()     {}

// ----------------------------------------------------------------------------
// Other Nodes

// ----------------------------------------------------------------------------
// File

// A File node represents a Lua source file.
//
// The Comments list contains all comments in the source file in order of
// appearance, including the comments that are pointed to from other nodes.
type File struct {
	Filename string
	Shebang  string
	Chunk    []Stmt
	Comments []*CommentGroup // list of all comments in the source file
}

func (f *File) Type() Type             { return FILE }
func (f *File) Pos() position.Position { return position.Position{Line: 1, Column: 1} }
func (f *File) End() position.Position {
	var pos position.Position
	if len(f.Chunk) > 0 {
		pos = f.Chunk[len(f.Chunk)-1].End()
	}
	if len(f.Comments) > 0 {
		cpos := f.Comments[len(f.Comments)-1].End()
		if pos.LessThan(cpos) {
			pos = cpos
		}
	}
	return pos
}

// ----------------------------------------------------------------------------
// Block

// A Block node represents a scoped statement list.
type Block struct {
	Opening position.Position // position of "do", "then", "repeat" or params.End()

	List []Stmt

	Closing position.Position // position of "end", "elseif", "else" or "until"
}

func (b *Block) Type() Type { return BLOCK }

func (b *Block) Pos() position.Position { return b.Opening }

func (b *Block) End() position.Position { return b.Closing }

// ----------------------------------------------------------------------------
// Function Block

// A Function Block node represents a function implementation.
type FuncBody struct {
	Params *ParamList
	Body   *Block
}

func (f *FuncBody) Type() Type { return FUNC_BODY }

func (f *FuncBody) Pos() position.Position { return f.Params.Pos() }

func (f *FuncBody) End() position.Position { return f.Body.End() }

// ----------------------------------------------------------------------------
// Param List

// A ParamList represents a list of Names, enclosed by parentheses or braces.
type ParamList struct {
	Lparen   position.Position // position of "("
	List     []*Name           // field list; or nil
	Ellipsis position.Position
	Rparen   position.Position // position of ")"
}

func (f *ParamList) Type() Type {
	return PARAM_LIST
}

func (f *ParamList) Pos() position.Position {
	if f.Lparen.IsValid() {
		return f.Lparen
	}
	// the list should not be empty in this case;
	// be conservative and guard against bad ASTs
	if len(f.List) > 0 {
		return f.List[0].Pos()
	}
	return position.NoPos
}

func (f *ParamList) End() position.Position {
	if f.Rparen.IsValid() {
		return f.Rparen.OffsetColumn(1)
	}
	// the list should not be empty in this case;
	// be conservative and guard against bad ASTs
	if n := len(f.List); n > 0 {
		return f.List[n-1].End()
	}
	return position.NoPos
}

// ----------------------------------------------------------------------------
// Comments

// A Comment node represents a single --style or --[[style comment.
type Comment struct {
	Hyphen position.Position // position of "-" starting the comment
	Text   string            // comment text (excluding '\n' for --style comments)
}

func (c *Comment) Type() Type             { return COMMENT }
func (c *Comment) Pos() position.Position { return c.Hyphen }
func (c *Comment) End() position.Position { return c.Hyphen.Offset(c.Text) }

// A CommentGroup represents a sequence of comments
// with no other tokens and no empty lines between.
//
type CommentGroup struct {
	List []*Comment // len(List) > 0
}

func (g *CommentGroup) Type() Type             { return COMMENT_GROUP }
func (g *CommentGroup) Pos() position.Position { return g.List[0].Pos() }
func (g *CommentGroup) End() position.Position { return g.List[len(g.List)-1].End() }

func isSpace(c byte) bool {
	return c == ' ' || uint(c)-'\t' < 5
}

func stripTrailingSpace(s string) string {
	i := len(s)
	for i > 0 && isSpace(s[i-1]) {
		i--
	}
	return s[0:i]
}

// Text returns the text of the comment.
// Comment markers (//, /*, and */), the first space of a line comment, and
// leading and trailing empty lines are removed. Multiple empty lines are
// reduced to one, and trailing space on lines is trimmed. Unless the result
// is empty, it is newline-terminated.
//
func (g *CommentGroup) Text() string {
	if g == nil {
		return ""
	}

	lines := make([]string, 0, len(g.List))
	for _, c := range g.List {
		text := c.Text
		if len(text) > 0 && text[0] == ' ' {
			text = text[1:]
		}

		// Split on newlines.
		tl := strings.Split(text, "\n")

		// Walk lines, stripping trailing white space and adding to list.
		for _, l := range tl {
			lines = append(lines, stripTrailingSpace(l))
		}
	}

	// Remove leading blank lines; convert runs of
	// interior blank lines to a single blank line.
	n := 0
	for _, line := range lines {
		if line != "" || n > 0 && lines[n-1] != "" {
			lines[n] = line
			n++
		}
	}
	lines = lines[0:n]

	// Add final "" entry to get trailing newline from Join.
	if n > 0 && lines[n-1] != "" {
		lines = append(lines, "")
	}

	return strings.Join(lines, "\n")
}
