package pattern

import (
	"unicode"
	"unicode/utf8"
)

type predicates []func(rune) bool

var preds predicates = []func(rune) bool{
	unicode.IsLetter,
	unicode.IsControl,
	unicode.IsDigit,
	func(r rune) bool { return r != ' ' && unicode.IsGraphic(r) },
	unicode.IsLower,
	unicode.IsPunct,
	unicode.IsSpace,
	unicode.IsUpper,
	isAlphaNum,
	func(r rune) bool { return unicode.Is(unicode.Hex_Digit, r) },
}

func (p predicates) is(op, r rune) (bool, bool) {
	if op <= utf8.MaxRune {
		return op == r, true
	}

	if opLetter <= op && op <= opHexDigit {
		return p[op-opLetter](r), true
	}

	if op <= opNotHexDigit {
		return !p[op-opNotLetter](r), true
	}

	return false, false
}

func isAlphaNum(r rune) bool {
	return (uint(r)|32)-'a' < 26 || uint(r)-'0' < 10
}
