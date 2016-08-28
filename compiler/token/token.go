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
