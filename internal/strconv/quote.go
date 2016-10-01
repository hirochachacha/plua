// Original: src/strconv/quote.go
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

package strconv

import (
	"strings"
	"unicode"
	"unicode/utf8"
)

// Quote returns a double-quoted Go string literal representing s.  The
// returned string uses Lua escape sequences (\t, \n, \xFF, \u{0100}) for
// control characters and non-printable characters as defined by
// IsPrint.
func Quote(s string) string {
	return quoteWith(s, '"')
}

func SQuote(s string) string {
	return quoteWith(s, '\'')
}

const lowerhex = "0123456789abcdef"

func quoteWith(s string, quote byte) string {
	var runeTmp [utf8.UTFMax]byte
	buf := make([]byte, 0, 3*len(s)/2) // Try to avoid more allocations.
	buf = append(buf, quote)
	for width := 0; len(s) > 0; s = s[width:] {
		r := rune(s[0])
		width = 1
		if r >= utf8.RuneSelf {
			r, width = utf8.DecodeRuneInString(s)
		}
		if width == 1 && r == utf8.RuneError {
			buf = append(buf, '\\')
			buf = AppendInt(buf, int64(s[0]), 10)
			continue
		}
		if r == rune(quote) || r == '\\' { // always backslashed
			buf = append(buf, '\\')
			buf = append(buf, byte(r))
			continue
		}
		if unicode.IsPrint(r) {
			n := utf8.EncodeRune(runeTmp[:], r)
			buf = append(buf, runeTmp[:n]...)
			continue
		}
		switch r {
		case '\a':
			// buf = append(buf, "\\\a"...)
			buf = append(buf, `\a`...)
		case '\b':
			// buf = append(buf, "\\\b"...)
			buf = append(buf, `\b`...)
		case '\f':
			// buf = append(buf, "\\\f"...)
			buf = append(buf, `\f`...)
		case '\n':
			// buf = append(buf, "\\\n"...)
			buf = append(buf, `\n`...)
		case '\r':
			// buf = append(buf, "\\\r"...)
			buf = append(buf, `\r`...)
		case '\t':
			// buf = append(buf, "\\\t"...)
			buf = append(buf, `\t`...)
		case '\v':
			// buf = append(buf, "\\\v"...)
			buf = append(buf, `\v`...)
		default:
			if r <= 0xff {
				buf = append(buf, '\\')
				buf = AppendInt(buf, int64(r), 10)
			} else {
				buf = append(buf, `\u{`...)
				for s := 28; s >= 0; s -= 4 {
					buf = append(buf, lowerhex[r>>uint(s)&0xF])
				}
				buf = append(buf, '}')
			}
		}
	}
	buf = append(buf, quote)
	return string(buf)

}

func unhex(b byte) (v rune, ok bool) {
	c := rune(b)
	switch {
	case '0' <= c && c <= '9':
		return c - '0', true
	case 'a' <= c && c <= 'f':
		return c - 'a' + 10, true
	case 'A' <= c && c <= 'F':
		return c - 'A' + 10, true
	}
	return
}

// UnquoteChar decodes the first character or byte in the escaped string
// or character literal represented by the string s.
// It returns four values:
//
//	1) value, the decoded Unicode code point or byte value;
//	2) multibyte, a boolean indicating whether the decoded character requires a multibyte UTF-8 representation;
//	3) tail, the remainder of the string after the character; and
//	4) an error that will be nil if the character is syntactically valid.
//
// The second argument, quote, specifies the type of literal being parsed
// and therefore which escaped quote character is permitted.
// If set to a single quote, it permits the sequence \' and disallows unescaped '.
// If set to a double quote, it permits \" and disallows unescaped ".
// If set to zero, it does not permit either escape and allows both quote characters to appear unescaped.
func unquoteChar(s string, quote byte) (value rune, multibyte bool, tail string, err error) {
	// easy cases
	switch c := s[0]; {
	case c == quote && (quote == '\'' || quote == '"'):
		err = ErrSyntax
		return
	case c >= utf8.RuneSelf:
		r, size := utf8.DecodeRuneInString(s)
		return r, true, s[size:], nil
	case c != '\\':
		return rune(s[0]), false, s[1:], nil
	}

	// hard case: c is backslash
	if len(s) <= 1 {
		err = ErrSyntax
		return
	}
	c := s[1]
	s = s[2:]

	switch c {
	case 'a':
		value = '\a'
	case 'b':
		value = '\b'
	case 'f':
		value = '\f'
	case 'n':
		value = '\n'
	case 'r':
		value = '\r'
	case 't':
		value = '\t'
	case 'v':
		value = '\v'
	case 'x':
		var v rune
		if len(s) < 2 {
			err = ErrSyntax
			return
		}
		for j := 0; j < 2; j++ {
			x, ok := unhex(s[j])
			if !ok {
				err = ErrSyntax
				return
			}
			v = v<<4 | x
		}
		s = s[2:]
		value = v
	case 'u':
		var v rune
		if len(s) < 3 {
			err = ErrSyntax
			return
		}
		if s[0] != '{' {
			err = ErrSyntax
			return
		}
		s = s[1:]
		var j int
		for j = 0; j < min(8, len(s)); j++ {
			x, ok := unhex(s[j])
			if !ok {
				break
			}
			v = v<<4 | x
		}
		if s[j] != '}' {
			err = ErrSyntax
			return
		}
		s = s[j+1:]
		if v > utf8.MaxRune {
			err = ErrSyntax
			return
		}
		value = v
		multibyte = true
	case '0', '1', '2', '3', '4', '5', '6', '7', '8', '9':
		v := rune(c) - '0'
		var j int
		for j = 0; j < min(2, len(s)); j++ { // one digit already; two more
			x := rune(s[j]) - '0'
			if x < 0 || x > 9 {
				break
			}
			v = v*10 + x
		}
		s = s[j:]
		if v > 255 {
			err = ErrSyntax
			return
		}
		value = v
	case '\\':
		value = '\\'
	case '\'', '"':
		value = rune(c)
	case '\n':
		value = '\n'
	case '\r':
		value = '\r'
		if len(s) > 0 && s[0] == '\n' {
			value = '\n'
			s = s[1:]
		}
	case 'z':
		s = strings.TrimLeft(s, " \t\r\n")
		value = -1
	default:
		err = ErrSyntax
		return
	}
	tail = s
	return
}

func unquoteLong(s string) (t string, err error) {
	n := len(s)

	switch prefix := s[:2]; prefix {
	case "[[":
		if n < 4 {
			return "", ErrSyntax
		}
		if "]]" != s[n-2:] {
			return "", ErrSyntax
		}

		s = s[2 : n-2]

		s = strings.Replace(s, "\r", "", -1)

		if len(s) > 0 && s[0] == '\n' {
			s = s[1:]
		}

		return s, nil
	case "[=":
		j := 2
		if n == j {
			return "", ErrSyntax
		}
		for s[j] == '=' {
			j++
			if n == j {
				return "", ErrSyntax
			}
		}
		if s[j] != '[' {
			return "", ErrSyntax
		}
		j++
		if n < 2*j {
			return "", ErrSyntax
		}

		s = s[j : n-j]

		s = strings.Replace(s, "\r", "", -1)

		if len(s) > 0 && s[0] == '\n' {
			s = s[1:]
		}

		return s, nil
	}

	return "", ErrSyntax
}

// Unquote interprets s as a single-quoted, double-quoted,
// or backquoted Go string literal, returning the string value
// that s quotes.  (If s is single-quoted, it would be a Go
// character literal; Unquote returns the corresponding
// one-character string.)
func Unquote(s string) (t string, err error) {
	n := len(s)
	if n < 2 {
		return "", ErrSyntax
	}

	quote := s[0]

	if quote == '[' {
		return unquoteLong(s)
	}

	if quote != '"' && quote != '\'' {
		return "", ErrSyntax
	}

	if quote != s[n-1] {
		return "", ErrSyntax
	}

	s = s[1 : n-1]

	// Is it trivial?  Avoid allocation.
	if !containsByte(s, '\\') && !containsByte(s, quote) {
		return s, nil
	}

	var runeTmp [utf8.UTFMax]byte
	buf := make([]byte, 0, 3*len(s)/2) // Try to avoid more allocations.
	for len(s) > 0 {
		c, multibyte, ss, err := unquoteChar(s, quote)
		if err != nil {
			return "", err
		}
		s = ss
		if c < 0 {
			continue
		}
		if c < utf8.RuneSelf || !multibyte {
			buf = append(buf, byte(c))
		} else {
			n := utf8.EncodeRune(runeTmp[:], c)
			buf = append(buf, runeTmp[:n]...)
		}
	}
	return string(buf), nil
}

// containsByte reports whether the string containsByte the byte c.
func containsByte(s string, c byte) bool {
	return strings.IndexByte(s, c) != -1
}

func min(x, y int) int {
	if x > y {
		return y
	}
	return x
}
