package strconv

func isXdigit(c byte) bool {
	return uint(c)-'0' < 10 || (uint(c)|32)-'a' < 6
}

func isSpace(c byte) bool {
	return c == ' ' || uint(c)-'\t' < 5
}

func isDigit(c byte) bool {
	return uint(c)-'0' < 10
}
