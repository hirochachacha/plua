package pattern

import (
	"unicode"
	"unicode/utf8"
)

type predicates []func(rune) bool

func (preds predicates) is(op, r rune) (bool, bool) {
	if op <= utf8.MaxRune {
		return op == r, true
	}

	if opLetter <= op && op <= opHexDigit {
		return preds[op-opLetter](r), true
	}

	if op <= opNotHexDigit {
		return !preds[op-opNotLetter](r), true
	}

	return false, false
}

var upreds predicates

var (
	_upreds = [...]func(rune) bool{
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
)

func init() {
	upreds = predicates(_upreds[:])
}

func isLetter(r rune) bool {
	return (uint(r)|32)-'a' < 26
}

func isControl(r rune) bool {
	return r == 0x7f || uint(r) < 0x20
}

func isDigit(r rune) bool {
	return uint(r)-'0' < 10
}

func isGraphic(r rune) bool {
	return uint(r)-0x21 < 0x5e
}

func isLower(r rune) bool {
	return uint(r)-'a' < 26
}

func isPunct(r rune) bool {
	return isGraphic(r) && !isAlphaNum(r)
}

func isSpace(r rune) bool {
	return r == ' ' || uint(r)-'\t' < 5
}

func isUpper(r rune) bool {
	return uint(r)-'A' < 26
}

func isAlphaNum(r rune) bool {
	return isLetter(r) || isDigit(r)
}

func isXdigit(r rune) bool {
	return isDigit(r) || (uint(r)|32)-'a' < 6
}
