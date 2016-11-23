package pattern

import "unicode/utf8"

func _decodeByte(input string, off int) (r rune, rsize int) {
	if len(input) == off {
		return eos, 0
	}
	return rune(input[off]), 1
}

func _lastDecodeByte(input string, off int) (r rune, rsize int) {
	if off == 0 {
		return sos, 0
	}
	return rune(input[off-1]), 1
}

func _decodeRune(input string, off int) (r rune, rsize int) {
	if len(input) == off {
		return eos, 0
	}
	return utf8.DecodeRuneInString(input[off:])
}

func _lastDecodeRune(input string, off int) (r rune, rsize int) {
	if off == 0 {
		return sos, 0
	}
	return utf8.DecodeLastRuneInString(input[:off])
}
