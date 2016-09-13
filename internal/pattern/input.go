package pattern

import "unicode/utf8"

const eot rune = -1

func decodeRune(s string, off int) (r rune, rsize int) {
	if len(s) <= off {
		return eot, 0
	}
	return utf8.DecodeRuneInString(s[off:])
}
