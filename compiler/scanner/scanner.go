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

type ScanState struct {
	sourceName string
	shebang    string

	mode Mode

	r io.Reader

	buf    []byte
	start  int
	end    int
	_mark  int
	filled bool

	clip bytes.Buffer

	ch int

	offset     int
	lineOffset int
	lineNum    int

	err error
}

func Scan(r io.Reader, srcname string, mode Mode) *ScanState {
	s := &ScanState{
		r:          r,
		sourceName: srcname,
		buf:        make([]byte, 4096),
		mode:       mode,
		_mark:      -1,
	}

	s.init()

	switch s.ch {
	case bom1:
		s.skipBom()
	case '#':
		s.shebang = s.scanSheBang()
	}

	return s
}

func (s *ScanState) Reset(r io.Reader, srcname string, mode Mode) {
	s.sourceName = srcname
	s.shebang = ""

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
		s.shebang = s.scanSheBang()
	}
}

func (s *ScanState) SourceName() string {
	return s.sourceName
}

func (s *ScanState) Shebang() string {
	return s.shebang
}

func (s *ScanState) Err() error {
	return s.err
}

func (s *ScanState) Next() token.Token {
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
			switch p := s.peek(2); p {
			case "..":
				s.next()
				s.next()
				if s.ch == '.' {
					s.next()
					typ = token.ELLIPSIS
				} else {
					typ = token.CONCAT
				}
			default:
				if len(p) == 2 && '0' <= p[1] && p[1] <= '9' {
					typ, lit = s.scanNumber(true)
				} else {
					s.next()
					typ = token.PERIOD
				}
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

func (s *ScanState) skipBom() {
	if s.peek(2) == string(bom) {
		s.next()
		s.next()
	}
}

func trimRightCR(s string) string {
	if len(s) > 0 && s[len(s)-1] == '\r' {
		s = s[:len(s)-1]
	}
	return s
}

func (s *ScanState) scanSheBang() (shebang string) {
	s.mark()

	s.next()
	for s.ch != '\n' {
		if s.ch == -1 {
			return trimRightCR(string(s.capture()))
		}
		s.next()
	}

	shebang = trimRightCR(string(s.capture()))

	s.next()

	return
}

func (s *ScanState) scanComment() (lit string) {
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

	for s.ch != '\n' && s.ch >= 0 {
		s.next()
	}

	lit = trimRightCR(string(s.capture()))

	return
}

func (s *ScanState) scanIdentifier() (lit string) {
	s.mark()

	s.next()

	for isLetter(s.ch) || isDigit(s.ch) {
		s.next()
	}

	return string(s.capture())
}

func (s *ScanState) skipMantissa(base int) {
	for digitVal(s.ch) < base {
		s.next()
	}
}

func (s *ScanState) scanNumber(seenDecimalPoint bool) (tok token.Type, lit string) {
	s.mark()

	if seenDecimalPoint {
		s.next() // skip .
	}

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

func (s *ScanState) scanString(quote int) (lit string) {
	s.mark()

	s.next()

	for s.ch != quote {
		if s.ch == '\n' || s.ch == '\r' || s.ch < 0 {
			lit = string(s.capture())

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

func (s *ScanState) skipEscape(quote int) {
	s.next()

	pos := s.pos()

	var pred func(int) bool
	var i, base, max uint32

	switch s.ch {
	case '\r':
		s.next()
		if s.ch == '\n' { // CRLN
			s.next()
		}
	case 'a', 'b', 'f', 'n', 'r', 't', 'v', '\\', '\n', '\'', '"':
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

	j := i
	for ; j > 0 && s.ch != quote && pred(s.ch); j-- {
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

	// hex
	if i == 2 {
		if j > 0 {
			s.error(pos, errUnknownEscapeSequence)

			return
		}
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

func (s *ScanState) scanLongString(simple bool) (lit string) {
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

func (s *ScanState) skipLongString(comment bool, simple bool) (err error) {
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

func (s *ScanState) skipSpace() {
	for isSpace(s.ch) {
		s.next()
	}
}

func (s *ScanState) error(pos position.Position, err error) {
	if s.err == nil {
		pos.SourceName = s.sourceName

		s.err = &Error{
			Pos: pos,
			Err: err,
		}
	}
}

func (s *ScanState) pos() position.Position {
	return position.Position{
		Line:   s.lineNum + 1,
		Column: s.offset - s.lineOffset,
	}
}

func (s *ScanState) mark() {
	if s._mark != -1 {
		panic("mark twice")
	}

	s._mark = s.start
}

func (s *ScanState) capture() []byte {
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

func (s *ScanState) init() {
	s.fill()

	if s.start == s.end {
		s.ch = -1
		s.start = 0
		s.end = 0

		return
	}

	s.ch = int(s.buf[s.start])
}

func (s *ScanState) next() {
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

func (s *ScanState) peek(n int) string {
	if n > s.end-s.start {
		s.fill()
		if n > s.end-s.start {
			return string(s.buf[s.start:s.end])
		}
	}

	return string(s.buf[s.start : s.start+n])
}

func (s *ScanState) fill() {
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
