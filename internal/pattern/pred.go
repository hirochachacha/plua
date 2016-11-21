package pattern

func isAlphaNum(r rune) bool {
	return isLetter(r) || isDigit(r)
}

func isLetter(r rune) bool {
	return (uint(r)|32)-'a' < 26
}

func isControl(r rune) bool {
	return uint(r) < 0x20 || r == 0x7f
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

func isPrint(r rune) bool {
	return uint(r)-0x20 < 0x5f
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

func isHexDigit(r rune) bool {
	return isDigit(r) || (uint(r)|32)-'a' < 6
}
