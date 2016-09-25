// Original: src/go/scanner/scanner.go
//
// Copyright 2009 The Go Authors. All rights reserved.
// Portions Copyright 2016 Hiroshi Ioka. All rights reserved.
//
// Redistribution and use in source and binary forms, with or without
// modification, are permitted provided that the following conditions are
// met:
//
//    * Redistributions of source code must retain the above copyright
// notice, this list of conditions and the following disclaimer.
//    * Redistributions in binary form must reproduce the above
// copyright notice, this list of conditions and the following disclaimer
// in the documentation and/or other materials provided with the
// distribution.
//    * Neither the name of Google Inc. nor the names of its
// contributors may be used to endorse or promote products derived from
// this software without specific prior written permission.
//
// THIS SOFTWARE IS PROVIDED BY THE COPYRIGHT HOLDERS AND CONTRIBUTORS
// "AS IS" AND ANY EXPRESS OR IMPLIED WARRANTIES, INCLUDING, BUT NOT
// LIMITED TO, THE IMPLIED WARRANTIES OF MERCHANTABILITY AND FITNESS FOR
// A PARTICULAR PURPOSE ARE DISCLAIMED. IN NO EVENT SHALL THE COPYRIGHT
// OWNER OR CONTRIBUTORS BE LIABLE FOR ANY DIRECT, INDIRECT, INCIDENTAL,
// SPECIAL, EXEMPLARY, OR CONSEQUENTIAL DAMAGES (INCLUDING, BUT NOT
// LIMITED TO, PROCUREMENT OF SUBSTITUTE GOODS OR SERVICES; LOSS OF USE,
// DATA, OR PROFITS; OR BUSINESS INTERRUPTION) HOWEVER CAUSED AND ON ANY
// THEORY OF LIABILITY, WHETHER IN CONTRACT, STRICT LIABILITY, OR TORT
// (INCLUDING NEGLIGENCE OR OTHERWISE) ARISING IN ANY WAY OUT OF THE USE
// OF THIS SOFTWARE, EVEN IF ADVISED OF THE POSSIBILITY OF SUCH DAMAGE.

package scanner

import (
	"bytes"
	"errors"
	"fmt"
	"io"
	"unicode"
	"unicode/utf8"

	"github.com/hirochachacha/plua/compiler/token"
	"github.com/hirochachacha/plua/position"
)

const (
	maxConsecutiveEmptyReads = 100
	bom                      = 0xFEFF // byte order mark, only permitted as very first character
	bom1                     = 0xFE
)

var (
	errInvalidLongStringDelimiter       = errors.New("invalid long string delimiter")
	errIllegalHexadecimalNumber         = errors.New("illegal hexadecimal number")
	errInvalidEscapeSequence            = errors.New("escape sequence is invalid Unicode code point")
	errUnknownEscapeSequence            = errors.New("unknown escape sequence")
	errMissingBracketInEscapeSequence   = errors.New("missing bracket in escape sequence")
	errIllegalCharacterInEscapeSequence = errors.New("illegal character in escape sequence")
	errUnterminatedString               = errors.New("unterminated string literal")
	errUnterminatedLongString           = errors.New("unterminated long string literal")
)

type Mode uint

const (
	ScanComments = 1 << iota
)

type Scanner struct {
	SourceName string
	Shebang    string

	mode Mode

	r io.Reader

	buf    []byte
	start  int
	end    int
	_mark  int
	filled bool

	clip bytes.Buffer

	runeBytes [utf8.UTFMax]byte

	ch int

	offset     int
	lineOffset int
	lineNum    int

	err error
}

func NewScanner(r io.Reader, srcname string, mode Mode) *Scanner {
	s := &Scanner{
		r:          r,
		SourceName: srcname,
		buf:        make([]byte, 4096),
		mode:       mode,
		_mark:      -1,
	}

	s.init()

	switch s.ch {
	case bom1:
		s.skipBom()
	case '#':
		s.Shebang = s.scanSheBang()
	}

	return s
}

func (s *Scanner) Reset(r io.Reader, filename string, mode Mode) {
	s.SourceName = filename
	s.Shebang = ""

	s.mode = mode
	s.r = r

	s.start = 0
	s.end = 0
	s._mark = -1
	s.filled = false

	s.clip.Reset()

	s.offset = 0
	s.lineOffset = 0
	s.lineNum = 0

	s.err = nil

	s.init()

	switch s.ch {
	case bom1:
		s.skipBom()
	case '#':
		s.Shebang = s.scanSheBang()
	}
}

func (s *Scanner) Err() error {
	return s.err
}

func (s *Scanner) Scan() token.Token {
	var typ token.Type
	var pos position.Position
	var lit string

scanAgain:
	s.skipSpace()

	pos = s.pos()

	switch ch := s.ch; {
	case isLetter(ch):
		lit = s.scanIdentifier()
		if len(lit) > 1 {
			// keywords are longer than one letter - avoid lookup otherwise
			typ = token.Lookup(lit)
		} else {
			typ = token.NAME
		}
	case isDigit(ch):
		typ, lit = s.scanNumber(false)
	default:
		switch ch {
		case -1:
			typ = token.EOF
		case '"', '\'':
			typ = token.STRING
			lit = s.scanString(ch)
		case ':':
			s.next()

			if s.ch == ':' {
				s.next()
				typ = token.LABEL
			} else {
				typ = token.COLON
			}
		case '.':
			s.next()

			if '0' <= s.ch && s.ch <= '9' {
				typ, lit = s.scanNumber(true)
			} else if s.ch == '.' {
				s.next()
				if s.ch == '.' {
					s.next()
					typ = token.ELLIPSIS
				} else {
					typ = token.CONCAT
				}
			} else {
				typ = token.PERIOD
			}
		case ',':
			s.next()

			typ = token.COMMA
		case ';':
			s.next()

			typ = token.SEMICOLON
		case '(':
			s.next()

			typ = token.LPAREN
		case ')':
			s.next()

			typ = token.RPAREN
		case '{':
			s.next()

			typ = token.LBRACE
		case '}':
			s.next()

			typ = token.RBRACE
		case '[':
			switch s.peek(2) {
			case "[[":
				typ = token.STRING
				lit = s.scanLongString(true)
			case "[=":
				typ = token.STRING
				lit = s.scanLongString(false)
			default:
				s.next()

				typ = token.LBRACK
			}
		case ']':
			s.next()

			typ = token.RBRACK
		case '+':
			s.next()

			typ = token.ADD
		case '-':
			if s.peek(2) == "--" {
				typ = token.COMMENT

				lit = s.scanComment()

				if s.mode&ScanComments == 0 {
					goto scanAgain
				}
			} else {
				s.next()

				typ = token.SUB
			}
		case '*':
			s.next()

			typ = token.MUL
		case '%':
			s.next()

			typ = token.MOD
		case '^':
			s.next()

			typ = token.POW
		case '/':
			s.next()

			if s.ch == '/' {
				s.next()
				typ = token.IDIV
			} else {
				typ = token.DIV
			}
		case '&':
			s.next()

			typ = token.BAND
		case '|':
			s.next()

			typ = token.BOR
		case '~':
			s.next()

			if s.ch == '=' {
				s.next()
				typ = token.NE
			} else {
				typ = token.BXOR
			}
		case '<':
			s.next()

			switch s.ch {
			case '<':
				s.next()
				typ = token.SHL
			case '=':
				s.next()
				typ = token.LE
			default:
				typ = token.LT
			}
		case '>':
			s.next()

			switch s.ch {
			case '>':
				s.next()
				typ = token.SHR
			case '=':
				s.next()
				typ = token.GE
			default:
				typ = token.GT
			}
		case '=':
			s.next()

			if s.ch == '=' {
				s.next()
				typ = token.EQ
			} else {
				typ = token.ASSIGN
			}
		case '#':
			s.next()

			typ = token.LEN
		default:
			s.next()
			s.error(pos, fmt.Errorf("illegal character %c", ch))
			typ = token.ILLEGAL
			lit = string(ch)
		}
	}

	return token.Token{Type: typ, Pos: pos, Lit: lit}
}

func (s *Scanner) skipBom() {
	if s.peek(2) == string(bom) {
		s.next()
		s.next()
	}
}

func (s *Scanner) scanSheBang() (shebang string) {
	if s.peek(2) == "#!" {
		s.mark()

		s.next()
		s.next()
		for s.ch != '\n' {
			if s.ch == -1 {
				return string(s.capture())
			}
			s.next()
		}

		shebang = string(s.capture())

		s.next()

		return
	}

	s.next()

	for s.ch != '\n' {
		if s.ch == -1 {
			return ""
		}
		s.next()
	}

	s.next()

	return ""
}

func (s *Scanner) scanComment() (lit string) {
	var err error

	s.mark()

	s.next() // skip '-'
	s.next() // skip '-'

	if s.ch == '[' {
		s.next()
		switch s.ch {
		case '[':
			err = s.skipLongString(true, true)
			if err != nil {
				s.error(s.pos(), err)
			}

			lit = string(s.capture())

			return
		case '=':
			err = s.skipLongString(true, false)
			if err != nil {
				s.error(s.pos(), err)
			}

			lit = string(s.capture())

			return
		}
	}

	for s.ch != '\n' && s.ch != '\r' && s.ch >= 0 {
		s.next()
	}

	lit = string(s.capture())

	return
}

func (s *Scanner) scanIdentifier() (lit string) {
	s.mark()

	s.next()

	for isLetter(s.ch) || isDigit(s.ch) {
		s.next()
	}

	return string(s.capture())
}

func (s *Scanner) skipMantissa(base int) {
	for digitVal(s.ch) < base {
		s.next()
	}
}

func (s *Scanner) scanNumber(seenDecimalPoint bool) (tok token.Type, lit string) {
	s.mark()

	tok = token.INT

	if seenDecimalPoint {
		tok = token.FLOAT
		s.skipMantissa(10)
		goto exponent
	}

	if s.ch == '0' {
		// int or float
		offs := s.offset
		pos := s.pos()

		s.next()

		// hexadecimal int or float
		if s.ch == 'x' || s.ch == 'X' {
			s.next()
			s.skipMantissa(16)

			// hex fraction
			if s.ch == '.' {
				tok = token.FLOAT
				s.next()
				s.skipMantissa(16)

				if s.ch == 'e' || s.ch == 'E' {
					tok = token.FLOAT
					s.next()
					if s.ch == '-' || s.ch == '+' {
						s.next()
					}
					s.skipMantissa(16)
				}

			}

			if s.offset-offs <= 2 {
				// only scanned "0x" or "0X"
				s.error(pos, errIllegalHexadecimalNumber)
			}

			// hex exponent
			if s.ch == 'p' || s.ch == 'P' {
				tok = token.FLOAT
				s.next()
				if s.ch == '-' || s.ch == '+' {
					s.next()
				}
				s.skipMantissa(10)
			}

			goto exit
		}
	}

	// decimal int or float
	s.skipMantissa(10)

	// fraction:
	if s.ch == '.' {
		tok = token.FLOAT
		s.next()
		s.skipMantissa(10)
	}

exponent:
	if s.ch == 'e' || s.ch == 'E' {
		tok = token.FLOAT
		s.next()
		if s.ch == '-' || s.ch == '+' {
			s.next()
		}
		s.skipMantissa(10)
	}

exit:
	lit = string(s.capture())

	return
}

func (s *Scanner) scanString(quote int) (lit string) {
	s.mark()

	s.next()

	for s.ch != quote {
		if s.ch == '\n' || s.ch == '\r' || s.ch < 0 {
			s.error(s.pos(), errUnterminatedString)

			return
		}

		if s.ch == '\\' {
			s.skipEscape(quote)
		} else {
			s.next()
		}
	}

	s.next()

	lit = string(s.capture())

	return
}

func (s *Scanner) skipEscape(quote int) {
	s.next()

	pos := s.pos()

	var pred func(int) bool
	var i, base, max uint32

	switch s.ch {
	case 'a', 'b', 'f', 'n', 'r', 't', 'v', '\\', '\r', '\n', '\'', '"':
		s.next()
		return
	case 'z':
		s.next()
		s.skipSpace()
		return
	case '0', '1', '2', '3', '4', '5', '6', '7', '8', '9':
		i, base, max = 3, 10, 255
		pred = isDigit
	case 'x':
		s.next()
		i, base, max = 2, 16, 255
		pred = isXdigit
	case 'u':
		s.next()

		if s.ch != '{' {
			s.error(pos, errMissingBracketInEscapeSequence)

			return
		}

		s.next()

		i, base, max = 8, 16, unicode.MaxRune
		pred = isXdigit
	default:
		ch := s.ch
		s.next() // always make progress
		s.error(pos, fmt.Errorf("unknown escape sequence %c", ch))

		return
	}

	var x uint32

	for ; i > 0 && s.ch != quote && pred(s.ch); i-- {
		d := uint32(digitVal(s.ch))
		if d >= base {
			// if not unicode
			if max != unicode.MaxRune {
				s.error(pos, fmt.Errorf("illegal character %c in escape sequence", s.ch))
			}

			break
		}

		// check overflow
		if x > (unicode.MaxRune-d)/base {
			s.error(pos, fmt.Errorf("escape sequence is invalid Unicode code point %c", s.ch))

			return
		}

		x = x*base + d

		s.next()
	}

	// unicode
	if max == unicode.MaxRune {
		if s.ch != '}' {
			s.error(pos, errMissingBracketInEscapeSequence)

			return
		}

		s.next()

		if 0xD800 <= x && x < 0xE000 {
			s.error(pos, fmt.Errorf("escape sequence is invalid Unicode code point %c", s.ch))
		}

		return
	}

	if x > max {
		s.error(pos, errInvalidEscapeSequence)
	}
}

func (s *Scanner) scanLongString(simple bool) (lit string) {
	var err error

	s.mark()

	s.next()

	err = s.skipLongString(false, simple)
	if err != nil {
		s.error(s.pos(), err)
	}

	lit = string(s.capture())

	return
}

func (s *Scanner) skipLongString(comment bool, simple bool) (err error) {
	s.next()

	if simple {
		for {
			for s.ch != ']' {
				if s.ch < 0 {
					err = errUnterminatedLongString

					return
				}
				s.next()
			}

			s.next()

			if s.ch == ']' {
				s.next()
				break
			}
		}

		return
	}

	depth := 1

	for s.ch == '=' {
		depth++
		s.next()
	}

	if s.ch != '[' {
		if comment {
			for s.ch != '\n' && s.ch != '\r' && s.ch >= 0 {
				s.next()
			}
			return
		}

		err = errInvalidLongStringDelimiter

		return
	}

	s.next()

	for {
		_depth := depth
		for s.ch != ']' {
			if s.ch < 0 {
				err = errUnterminatedLongString

				return
			}
			s.next()
		}

		s.next()

		for s.ch == '=' {
			_depth--
			s.next()
		}

		if _depth != 0 {
			continue
		}

		if s.ch == ']' {
			s.next()
			break
		}
	}

	return
}

func (s *Scanner) skipSpace() {
	for isSpace(s.ch) {
		s.next()
	}
}

func (s *Scanner) error(pos position.Position, err error) {
	if s.err == nil {
		pos.SourceName = s.SourceName

		s.err = &Error{
			Pos: pos,
			Err: err,
		}
	}
}

func (s *Scanner) pos() position.Position {
	return position.Position{
		Line:   s.lineNum + 1,
		Column: s.offset - s.lineOffset,
	}
}

func (s *Scanner) mark() {
	if s._mark != -1 {
		panic("mark twice")
	}

	s._mark = s.start
}

func (s *Scanner) capture() []byte {
	if s._mark == -1 {
		panic("no mark")
	}

	buf := s.buf[s._mark:s.start]

	s._mark = -1

	if s.clip.Len() > 0 {
		s.clip.Write(buf)
		buf = s.clip.Bytes()
		s.clip.Reset()
	}

	return buf
}

func (s *Scanner) init() {
	s.fill()

	if s.start == s.end {
		s.ch = -1
		s.start = 0
		s.end = 0

		return
	}

	s.ch = int(s.buf[s.start])
}

func (s *Scanner) next() {
	if s.ch == -1 {
		return
	}

	if s.ch == '\n' {
		s.lineOffset = s.offset
		s.lineNum++
	}

	s.start++
	s.offset++

	if s.start == s.end {
		s.fill()
		if s.start == s.end {
			s.ch = -1
			s.start = 0
			s.end = 0

			return
		}
	}

	s.ch = int(s.buf[s.start])
}

func (s *Scanner) peek(n int) string {
	if n > s.end-s.start {
		s.fill()
		if n > s.end-s.start {
			return string(s.buf[s.start:s.end])
		}
	}

	return string(s.buf[s.start : s.start+n])
}

func (s *Scanner) fill() {
	if s.filled {
		return
	}

	if s.start > 0 {
		if s._mark != -1 {
			s.clip.Write(s.buf[s._mark:s.start])

			s._mark = 0
		}

		copy(s.buf, s.buf[s.start:s.end])
		s.end -= s.start
		s.start = 0
	}

	for i := maxConsecutiveEmptyReads; i > 0; i-- {
		n, err := s.r.Read(s.buf[s.end:])
		if err == io.EOF {
			s.filled = true

			return
		}
		if n < 0 {
			panic("reader returned negative count from Read")
		}
		s.end += n
		if err != nil {
			s.error(position.NoPos, err)
			return
		}

		if n > 0 {
			return
		}
	}
	s.error(position.NoPos, io.ErrNoProgress)
}

func digitVal(ch int) int {
	switch {
	case uint(ch)-'0' < 10:
		return int(ch - '0')
	case uint(ch)-'a' < 6:
		return int(ch - 'a' + 10)
	case uint(ch)-'A' < 6:
		return int(ch - 'A' + 10)
	}

	return 16 // larger than any legal digit val
}

func isSpace(ch int) bool {
	return ch == ' ' || uint(ch)-'\t' < 5
}

func isLetter(ch int) bool {
	return ch == '_' || (uint(ch)|32)-'a' < 26
}

func isDigit(ch int) bool {
	return uint(ch)-'0' < 10
}

func isXdigit(ch int) bool {
	return uint(ch)-'0' < 10 || (uint(ch)|32)-'a' < 6
}
