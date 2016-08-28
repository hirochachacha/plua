// Original: src/go/token/token.go
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

package token

import (
	"github.com/hirochachacha/blua/position"
)

type Token struct {
	Type
	Pos position.Position
	Lit string
}

func Lookup(name string) Type {
	if typ, ok := keywords[name]; ok {
		return typ
	}
	return NAME
}

// A set of constants for precedence-based expression parsing.
// Non-operators have lowest precedence, followed by operators
// starting with precedence 1 up to unary operators. The highest
// precedence serves as "catch-all" precedence for selector,
// indexing, and other operator and delimiter tokens.
//
const (
	LowestPrec  = 0 // non-operators
	UnaryPrec   = 12
	HighestPrec = 14
)

// Precedence returns the operator precedence of the binary
// operator op. If op is not a binary operator, the result
// is LowestPrecedence.
//
func (op Type) Precedence() (left int, right int) {
	switch op {
	case OR:
		return 1, 1
	case AND:
		return 2, 2
	case EQ, NE, LT, LE, GT, GE:
		return 3, 3
	case BOR:
		return 4, 4
	case BXOR:
		return 5, 5
	case BAND:
		return 6, 6
	case SHL, SHR:
		return 7, 7
	case CONCAT: // right associative
		return 9, 8
	case ADD, SUB:
		return 10, 10
	case MUL, DIV, IDIV, MOD:
		return 11, 11
	case POW: // right associative
		return 14, 13
	}
	return LowestPrec, LowestPrec
}

func (typ Type) IsLiteral() bool {
	return literal_beg < typ && typ < literal_end
}

func (typ Type) IsOperator() bool {
	return operator_beg < typ && typ < operator_end
}

func (typ Type) IsKeyword() bool {
	return keyword_beg < typ && typ < keyword_end
}
