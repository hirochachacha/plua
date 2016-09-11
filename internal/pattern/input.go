package pattern

import (
	"strings"
	"unicode/utf8"
)

const endOfText rune = -1

// input abstracts different representations of the input text. It provides
// one-character lookahead.
type input interface {
	String() string
	length() int
	slice(i, j int) string
	stepRune(pos int) (r rune, width int) // advance one rune
	stepByte(pos int) (r rune, width int) // advance one byte
	hasPrefix(m *machine) bool
	isPrefix(m *machine) bool
	hasSuffix(m *machine) bool
	index(m *machine, pos int) int
	submatch(begin, end, pos int) bool
}

// inputString scans a string.
type inputString string

func (s inputString) String() string {
	return string(s)
}

func (s inputString) length() int {
	return len(s)
}

func (s inputString) slice(i, j int) string {
	return string(s[i:j])
}

func (s inputString) stepRune(pos int) (rune, int) {
	if pos < len(s) {
		c := s[pos]
		if c < utf8.RuneSelf {
			return rune(c), 1
		}
		return utf8.DecodeRuneInString(string(s[pos:]))
	}
	return endOfText, 0
}

func (s inputString) stepByte(pos int) (rune, int) {
	if pos < len(s) {
		return rune(s[pos]), 1
	}
	return endOfText, 0
}

func (s inputString) hasPrefix(m *machine) bool {
	return strings.HasPrefix(string(s), m.prefix)
}

func (s inputString) isPrefix(m *machine) bool {
	return string(s) == m.prefix
}

func (s inputString) hasSuffix(m *machine) bool {
	return strings.HasSuffix(string(s), m.prefix)
}

func (s inputString) index(m *machine, pos int) int {
	if pos >= len(s) {
		return -1
	}
	return strings.Index(string(s[pos:]), m.prefix)
}

func (s inputString) submatch(begin, end, offset int) bool {
	return s[begin:end] == s[offset:offset+end-begin]
}
