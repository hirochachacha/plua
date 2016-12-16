package io

import (
	"io"
	"strings"

	"github.com/hirochachacha/plua/internal/strconv"
	"github.com/hirochachacha/plua/object"
)

type scanner struct {
	sc  io.ByteScanner
	c   byte
	err error
	buf []byte
}

func newScanner(sc io.ByteScanner) *scanner {
	s := &scanner{
		sc: sc,
	}

	s.next()

	return s
}

func (s *scanner) next() {
	if s.err == nil {
		s.c, s.err = s.sc.ReadByte()
	}
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

func (s *scanner) scanNumber() (object.Value, error) {
	s.skipSpace()

	s.acceptSign()

	base, zero, _ := s.acceptBase()
	if zero {
		return object.Integer(0), nil
	}

	if base == 16 {
		s.acceptHexDigits()
		if s.acceptByte('.') {
			s.acceptHexDigits()
		}
		if s.accepts("pP") {
			s.acceptSign()
			s.acceptDigits()
		} else if s.accepts("eE") {
			s.acceptSign()
			s.acceptDigits()
		}
	} else {
		s.acceptDigits()
		if s.acceptByte('.') {
			s.acceptDigits()
		}
		if s.accepts("eE") {
			s.acceptSign()
			s.acceptDigits()
		}
	}

	if s.err != nil {
		return nil, s.err
	}

	if err := s.sc.UnreadByte(); err != nil {
		return nil, err
	}

	str := string(s.buf)

	if i, err := strconv.ParseInt(str); err == nil {
		return object.Integer(i), err
	}

	f, err := strconv.ParseFloat(str)
	if err != nil {
		return nil, err
	}

	return object.Number(f), nil
}

func isXdigit(c byte) bool {
	return uint(c)-'0' < 10 || (uint(c)|32)-'a' < 6
}

func isSpace(c byte) bool {
	return c == ' ' || uint(c)-'\t' < 5
}

func isDigit(c byte) bool {
	return uint(c)-'0' < 10
}
