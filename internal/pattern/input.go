package pattern

import (
	"bytes"
	"strings"
	"unicode/utf8"
)

const endOfText rune = -1

// input abstracts different representations of the input text. It provides
// one-character lookahead.
type input interface {
	String() string
	length() int
	slices(i, j int) (string, []byte)
	stepRune(pos int) (r rune, width int) // advance one rune
	stepByte(pos int) (r rune, width int) // advance one byte
	hasPrefix(m *machine) bool
	isPrefix(m *machine) bool
	hasSuffix(m *machine) bool
	index(m *machine, pos int) int
	submatch(begin, end, pos int) bool
}

// inputString scans a string.
type inputString struct {
	str string
}

func (i *inputString) String() string {
	return i.str
}

func (i *inputString) length() int {
	return len(i.str)
}

func (is *inputString) slices(i, j int) (string, []byte) {
	s := is.str[i:j]
	return s, []byte(s)
}

func (i *inputString) stepRune(pos int) (rune, int) {
	if pos < len(i.str) {
		c := i.str[pos]
		if c < utf8.RuneSelf {
			return rune(c), 1
		}
		return utf8.DecodeRuneInString(i.str[pos:])
	}
	return endOfText, 0
}

func (i *inputString) stepByte(pos int) (rune, int) {
	if pos < len(i.str) {
		return rune(i.str[pos]), 1
	}
	return endOfText, 0
}

func (i *inputString) hasPrefix(m *machine) bool {
	return strings.HasPrefix(i.str, m.prefix)
}

func (i *inputString) isPrefix(m *machine) bool {
	return i.str == m.prefix
}

func (i *inputString) hasSuffix(m *machine) bool {
	return strings.HasSuffix(i.str, m.prefix)
}

func (i *inputString) index(m *machine, pos int) int {
	if pos >= len(i.str) {
		return -1
	}
	return strings.Index(i.str[pos:], m.prefix)
}

func (i *inputString) submatch(begin, end, offset int) bool {
	return i.str[begin:end] == i.str[offset:offset+end-begin]
}

// inputBytes scans a byte slice.
type inputBytes struct {
	str []byte
}

func (i *inputBytes) String() string {
	return string(i.str)
}

func (i *inputBytes) length() int {
	return len(i.str)
}

func (ib *inputBytes) slices(i, j int) (string, []byte) {
	bs := ib.str[i:j]
	return string(bs), bs
}

func (i *inputBytes) stepRune(pos int) (rune, int) {
	if pos < len(i.str) {
		c := i.str[pos]
		if c < utf8.RuneSelf {
			return rune(c), 1
		}
		return utf8.DecodeRune(i.str[pos:])
	}
	return endOfText, 0
}

func (i *inputBytes) stepByte(pos int) (rune, int) {
	if pos < len(i.str) {
		return rune(i.str[pos]), 1
	}
	return endOfText, 0
}

func (i *inputBytes) hasPrefix(m *machine) bool {
	return bytes.HasPrefix(i.str, m.prefixBytes)
}

func (i *inputBytes) isPrefix(m *machine) bool {
	return bytes.Equal(i.str, m.prefixBytes)
}

func (i *inputBytes) hasSuffix(m *machine) bool {
	return bytes.HasSuffix(i.str, m.prefixBytes)
}

func (i *inputBytes) index(m *machine, pos int) int {
	if pos >= len(i.str) {
		return -1
	}
	return bytes.Index(i.str[pos:], m.prefixBytes)
}

func (i *inputBytes) submatch(begin, end, offset int) bool {
	return bytes.Equal(i.str[begin:end], i.str[offset:offset+end-begin])
}
