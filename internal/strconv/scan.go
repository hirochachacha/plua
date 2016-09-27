package strconv

import (
	"io"
	"strings"
)

type scanner struct {
	sc  io.ByteScanner
	c   byte
	err error
	buf []byte
}

func newScanner(sc io.ByteScanner) *scanner {
	return &scanner{
		sc: sc,
	}
}

func (s *scanner) next() {
	s.c, s.err = s.sc.ReadByte()
}

func (s *scanner) accept() {
	s.buf = append(s.buf, s.c)
	s.next()
}

func (s *scanner) skip() {
	s.next()
}

func (s *scanner) acceptByte(c byte) bool {
	if s.c == c {
		s.accept()
		return true
	}
	return false
}

func (s *scanner) skipByte(c byte) bool {
	if s.c == c {
		s.skip()
		return true
	}
	return false
}

func (s *scanner) accepts(str string) bool {
	if strings.IndexByte(str, s.c) != -1 {
		s.accept()
		return true
	}
	return false
}

func (s *scanner) acceptUntil(pred func(c byte) bool) {
	for pred(s.c) {
		s.accept()
	}
}

func (s *scanner) skipUntil(pred func(c byte) bool) {
	for pred(s.c) {
		s.skip()
	}
}

func (s *scanner) skipSpace() {
	s.skipUntil(isSpace)
}

func (s *scanner) acceptSign() (neg bool, ok bool) {
	if s.acceptByte('+') {
		return false, true
	}
	if s.acceptByte('-') {
		return true, true
	}
	return false, false
}

func (s *scanner) acceptBase() (base int, zero bool, ok bool) {
	base = 10
	if s.acceptByte('0') {
		if s.accepts("xX") {
			base = 16
		} else {
			for s.c == '0' {
				s.accept()
			}
			if s.err == io.EOF {
				return 10, true, true
			}
		}
	}

	return base, false, true
}

func (s *scanner) acceptDigits() {
	s.acceptUntil(isDigit)
}

func (s *scanner) acceptHexDigits() {
	s.acceptUntil(isXdigit)
}

func (s *scanner) scanUint() (uint64, error) {
	neg, _ := s.acceptSign()
	if neg {
		return 0, ErrSyntax
	}

	base, zero, _ := s.acceptBase()
	if zero {
		return 0, nil
	}

	if s.err != nil {
		return 0, s.err
	}

	if base == 16 {
		s.acceptHexDigits()
	} else {
		s.acceptDigits()
	}

	if s.err != nil && s.err != io.EOF {
		return 0, s.err
	}

	err := s.sc.UnreadByte()
	if err != nil {
		return 0, err
	}

	return ParseUint(string(s.buf))
}

func (s *scanner) scanInt() (int64, error) {
	s.skipSpace()

	s.acceptSign()

	base, zero, _ := s.acceptBase()
	if zero {
		return 0, nil
	}

	if s.err != nil {
		return 0, s.err
	}

	if base == 16 {
		s.acceptHexDigits()
	} else {
		s.acceptDigits()
	}

	if s.err != nil && s.err != io.EOF {
		return 0, s.err
	}

	return ParseInt(string(s.buf))
}

func (s *scanner) scanFloat() (float64, error) {
	s.skipSpace()

	s.acceptSign()

	base, zero, _ := s.acceptBase()
	if zero {
		return 0, nil
	}

	if s.err != nil {
		return 0, s.err
	}

	if base == 16 {
		s.acceptHexDigits()
		s.acceptByte('.')
		s.acceptHexDigits()
		s.accepts("pP")
		s.acceptSign()
		s.acceptDigits()
	} else {
		s.acceptDigits()
		s.acceptByte('.')
		s.acceptDigits()
	}

	if s.err != nil && s.err != io.EOF {
		return 0, s.err
	}

	return ParseFloat(string(s.buf))
}
