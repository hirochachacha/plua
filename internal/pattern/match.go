package pattern

import "unicode/utf8"

func simpleMatch(op opCode, r rune) bool {
	switch op {
	case opAny:
		return true
	case opLetter:
		return isLetter(r)
	case opControl:
		return isControl(r)
	case opDigit:
		return isDigit(r)
	case opGraphic:
		return isGraphic(r)
	case opLower:
		return isLower(r)
	case opPunct:
		return isPunct(r)
	case opSpace:
		return isSpace(r)
	case opUpper:
		return isUpper(r)
	case opAlphaNum:
		return isAlphaNum(r)
	case opHexDigit:
		return isHexDigit(r)
	case opZero:
		return r == 0
	case opNotLetter:
		return !isLetter(r)
	case opNotControl:
		return !isControl(r)
	case opNotDigit:
		return !isDigit(r)
	case opNotGraphic:
		return !isGraphic(r)
	case opNotLower:
		return !isLower(r)
	case opNotPunct:
		return !isPunct(r)
	case opNotSpace:
		return !isSpace(r)
	case opNotUpper:
		return !isUpper(r)
	case opNotAlphaNum:
		return !isAlphaNum(r)
	case opNotHexDigit:
		return !isHexDigit(r)
	case opNotZero:
		return r != 0
	default:
		if 0 <= op && op <= utf8.MaxRune {
			return rune(op) == r
		}
		return false
	}
}

type r32 struct {
	low int32
	hi  int32
}

type set struct {
	elems  []rune
	ranges []r32
}

func (s *set) match(r rune) bool {
	for _, e := range s.elems {
		if simpleMatch(opCode(e), r) {
			return true
		}
	}
	for _, rng := range s.ranges {
		if rng.low <= r && r <= rng.hi {
			return true
		}
	}
	return false
}
