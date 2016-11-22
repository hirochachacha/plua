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

func Quote(s string) string {
	return quoteWith(s, '"')
}

const lowerhex = "0123456789abcdef"

func quoteWith(s string, quote byte) string {
	b := make([]byte, 0, len(s)*3/2)

	b = append(b, quote)

	runeTmp := make([]byte, utf8.UTFMax)

	i := 0
	for i < len(s) {
		c := s[i]

		switch c {
		case quote, '\\':
			b = append(b, '\\')
			b = append(b, c)
			i++
		case '\a':
			b = append(b, `\a`...)
			i++
		case '\b':
			b = append(b, `\b`...)
			i++
		case '\f':
			b = append(b, `\f`...)
			i++
		case '\n':
			b = append(b, `\n`...)
			i++
		case '\r':
			b = append(b, `\r`...)
			i++
		case '\t':
			b = append(b, `\t`...)
			i++
		case '\v':
			b = append(b, `\v`...)
			i++
		default:
			if c >= utf8.RuneSelf {
				r, rsize := utf8.DecodeRuneInString(s[i:])
				if r != utf8.RuneError {
					if unicode.IsPrint(r) {
						n := utf8.EncodeRune(runeTmp[:], r)
						b = append(b, runeTmp[:n]...)
					} else {
						b = append(b, `\u{`...)
						for j := 28; j >= 0; j -= 4 {
							b = append(b, lowerhex[r>>uint(j)&0xF])
						}
						b = append(b, '}')
					}

					i += rsize

					continue
				}
			}

			if isPrint(c) {
				b = append(b, c)
			} else {
				b = append(b, '\\')
				b = AppendInt(b, int64(c), 10)
			}
			i++
		}
	}

	b = append(b, quote)

	return string(b)
}

func isPrint(c byte) bool {
	return uint(c)-0x20 < 0x5f
}

func Unquote(s string) (string, error) {
	if len(s) < 2 {
		return "", ErrSyntax
	}

	quote := s[0]

	if quote == '[' {
		return unquoteLong(s)
	}

	if quote != '"' && quote != '\'' {
		return "", ErrSyntax
	}

	if quote != s[len(s)-1] {
		return "", ErrSyntax
	}

	s = s[1 : len(s)-1]

	if !strings.ContainsAny(s, `\`+string(quote)) {
		return s, nil
	}

	b := make([]byte, 0, len(s))

	runeTmp := make([]byte, utf8.UTFMax)

	i := 0
	for i < len(s) {
		c := s[i]

		switch c {
		case quote:
			return "", ErrSyntax
		case '\\':
			i++
			if i == len(s) {
				return "", ErrSyntax
			}
			switch c = s[i]; c {
			case 'a':
				b = append(b, '\a')
				i++
			case 'b':
				b = append(b, '\b')
				i++
			case 'f':
				b = append(b, '\f')
				i++
			case 'n':
				b = append(b, '\n')
				i++
			case 'r':
				b = append(b, '\r')
				i++
			case 't':
				b = append(b, '\t')
				i++
			case 'v':
				b = append(b, '\v')
				i++
			case 'x':
				i++
				if len(s)-i < 2 {
					return "", ErrSyntax
				}
				var c1 int
				for lim := i + 2; i < lim; i++ {
					d := digitVal(s[i])
					if d == 16 {
						return "", ErrSyntax
					}
					c1 = c1<<4 | d
				}
				b = append(b, byte(c1))
			case 'u':
				i++
				if len(s)-i < 3 {
					return "", ErrSyntax
				}
				if s[i] != '{' {
					return "", ErrSyntax
				}
				i++
				var r rune
				for lim := i + 8; i < min(lim, len(s)); i++ {
					d := digitVal(s[i])
					if d == 16 {
						break
					}
					r = r<<4 | rune(d)
				}
				if s[i] != '}' {
					return "", ErrSyntax
				}
				i++
				if r < 0 || r > utf8.MaxRune {
					return "", ErrSyntax
				}
				n := utf8.EncodeRune(runeTmp, r)
				b = append(b, runeTmp[:n]...)
			case '0', '1', '2', '3', '4', '5', '6', '7', '8', '9':
				var c1 int
				for lim := i + 3; i < min(lim, len(s)); i++ {
					d := s[i] - '0'
					if d < 0 || d > 9 {
						break
					}
					c1 = c1*10 + int(d)
				}
				if c1 < 0 || c1 > 255 {
					return "", ErrSyntax
				}
				b = append(b, byte(c1))
			case '\\':
				b = append(b, c)
				i++
			case '\'', '"':
				b = append(b, c)
				i++
			case '\n':
				b = append(b, c)
				i++
			case '\r':
				i++
				if i < len(s) && s[i] == '\n' {
					b = append(b, '\n')
					i++
				} else {
					b = append(b, '\r')
				}
			case 'z':
				i++
				for i < len(s) && isSpace(s[i]) {
					i++
				}
			default:
				return "", ErrSyntax
			}
		default:
			b = append(b, c)
			i++
		}
	}

	return string(b), nil
}

func unquoteLong(s string) (string, error) {
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

func min(x, y int) int {
	if x < y {
		return x
	}
	return y
}
