package format

import (
	"bytes"
	"fmt"
	"math"
	"unicode"
	"unicode/utf8"

	"github.com/hirochachacha/plua/internal/strconv"
	"github.com/hirochachacha/plua/object"
)

type ArgError struct {
	ArgNum   int
	ExtraMsg string
}

func (a *ArgError) Error() string {
	return ""
}

type TypeError struct {
	ArgNum   int
	Expected string
}

func (t *TypeError) Error() string {
	return ""
}

type OptionError struct {
	Opt string
}

func (o *OptionError) Error() string {
	return ""
}

type term struct {
	verb byte

	width int // use 0 instead of false
	prec  int // use 0 instead of false

	plus  bool
	minus bool
	space bool
	zero  bool
}

func (t *term) writeFormat(buf *bytes.Buffer, val object.Value) (n int, err error) {
	if t.minus {
		n, err = t.writeVerb(buf, val)
		if err != nil {
			return
		}

		if t.zero {
			if t.width > n {
				for i := 0; i < t.width-n; i++ {
					buf.WriteRune('0')
				}
				n = t.width
			}
		} else {
			if t.width > n {
				for i := 0; i < t.width-n; i++ {
					buf.WriteRune(' ')
				}
				n = t.width
			}
		}
	} else {
		tmp := new(bytes.Buffer)
		n, err = t.writeVerb(tmp, val)
		if err != nil {
			return
		}

		if t.zero {
			if t.width > n {
				for i := 0; i < t.width-n; i++ {
					buf.WriteRune('0')
				}
				n = t.width
			}
		} else {
			if t.width > n {
				for i := 0; i < t.width-n; i++ {
					buf.WriteRune(' ')
				}
				n = t.width
			}
		}

		_, err = tmp.WriteTo(buf)
	}

	return
}

func (t *term) writeVerb(buf *bytes.Buffer, val object.Value) (n int, err error) {
	switch t.verb {
	case 'c':
		return t.writeByte(buf, val)
	case 'd':
		return t.writeInt(buf, val, 10, false)
	case 'i':
		return t.writeInt(buf, val, 10, false)
	case 'o':
		return t.writeInt(buf, val, 8, false)
	case 'u':
		return t.writeUint(buf, val, 10)
	case 'x':
		return t.writeInt(buf, val, 16, false)
	case 'X':
		return t.writeInt(buf, val, 16, true)
	case 'e':
		return t.writeFloat(buf, val, 'e')
	case 'E':
		return t.writeFloat(buf, val, 'E')
	case 'f':
		return t.writeFloat(buf, val, 'f')
	case 'a':
		return t.writeHexFloat(buf, val, false)
	case 'A':
		return t.writeHexFloat(buf, val, true)
	case 'g':
		return t.writeFloat(buf, val, 'g')
	case 'G':
		return t.writeFloat(buf, val, 'G')
	case 'q':
		return t.writeQuoteString(buf, val)
	case 's':
		return t.writeString(buf, val)
	}

	return 0, &OptionError{"%" + string(t.verb)}
}

func (t *term) writeByte(buf *bytes.Buffer, val object.Value) (n int, err error) {
	i, ok := object.ToGoInt64(val)
	if !ok {
		return 0, &TypeError{Expected: "integer"}
	}

	return 1, buf.WriteByte(byte(i))
}

func (t *term) writeInt(buf *bytes.Buffer, val object.Value, base int, isUpper bool) (n int, err error) {
	i, ok := object.ToGoInt64(val)
	if !ok {
		return 0, &TypeError{Expected: "integer"}
	}

	if t.plus {
		if i < 0 {
			_, err = buf.WriteRune('-')
			i = -i
		} else {
			_, err = buf.WriteRune('+')
		}

		if err != nil {
			return
		}

		n++
	} else if i < 0 {
		_, err = buf.WriteRune('-')
		if err != nil {
			return
		}

		i = -i
		n++
	}

	ndigits := 1
	for n := i; n != 0; n /= 10 {
		ndigits++
	}

	if 0 < t.prec-ndigits {
		for j := 0; j < t.prec-ndigits; j++ {
			_, err = buf.WriteRune('0')
			if err != nil {
				return
			}
			n++
		}
	}

	dst := strconv.AppendInt(nil, i, base)

	if base > 10 && isUpper {
		for j := 0; j < len(dst); j++ {
			dst[j] = byte(unicode.ToUpper(rune(dst[j])))
		}
	}

	m, err := buf.Write(dst)
	if err != nil {
		return 0, err
	}

	n += m

	return
}

func (t *term) writeUint(buf *bytes.Buffer, val object.Value, base int) (n int, err error) {
	i, ok := object.ToGoInt64(val)
	if !ok {
		return 0, &TypeError{Expected: "integer"}
	}

	if t.plus {
		if i < 0 {
			_, err = buf.WriteRune('-')
		} else {
			_, err = buf.WriteRune('+')
		}

		if err != nil {
			return
		}

		n++
	}

	ndigits := 1
	for n := i; n != 0; n /= 10 {
		ndigits++
	}

	if 0 < t.prec-ndigits {
		for j := 0; j < t.prec-ndigits; j++ {
			_, err = buf.WriteRune('0')
			if err != nil {
				return
			}
			n++
		}
	}

	m, err := buf.WriteString(strconv.FormatUint(uint64(i), base))
	if err != nil {
		return 0, err
	}

	n += m

	return
}

func (t *term) writeFloat(buf *bytes.Buffer, val object.Value, fmt byte) (n int, err error) {
	f, ok := object.ToGoFloat64(val)
	if !ok {
		return 0, &TypeError{Expected: "number"}
	}

	if t.plus && f >= 0 {
		_, err = buf.WriteRune('+')
		if err != nil {
			return
		}
		n++
	}

	m, err := buf.WriteString(strconv.FormatFloat(f, fmt, t.prec, 64))
	if err != nil {
		return 0, err
	}

	n += m

	return
}

func (t *term) writeHexFloat(buf *bytes.Buffer, val object.Value, isUpper bool) (n int, err error) {
	f, ok := object.ToGoFloat64(val)
	if !ok {
		return 0, &TypeError{Expected: "number"}
	}

	u := math.Float64bits(f)

	sign := u >> 63
	exponent := int64(u>>52&0x7ff) - 1023
	fraction := u & 0xfffffffffffff

	if isUpper {
		if sign == 1 {
			n, err = fmt.Fprintf(buf, "-0X1.%Xp%+d", fraction, exponent)
		} else if t.plus {
			n, err = fmt.Fprintf(buf, "+0X1.%Xp%+d", fraction, exponent)
		} else {
			n, err = fmt.Fprintf(buf, "0X1.%Xp%+d", fraction, exponent)
		}
	} else {
		if sign == 1 {
			n, err = fmt.Fprintf(buf, "-0x1.%xp%+d", fraction, exponent)
		} else if t.plus {
			n, err = fmt.Fprintf(buf, "+0x1.%xp%+d", fraction, exponent)
		} else {
			n, err = fmt.Fprintf(buf, "0x1.%xp%+d", fraction, exponent)
		}
	}
	return
}

func (t *term) writeQuoteString(buf *bytes.Buffer, val object.Value) (n int, err error) {
	s, ok := object.ToGoString(val)
	if !ok {
		return 0, &TypeError{Expected: "string"}
	}

	s = strconv.Quote(s)

	_, err = buf.WriteString(s)

	n = utf8.RuneCountInString(s)

	return
}

func (t *term) writeString(buf *bytes.Buffer, val object.Value) (n int, err error) {
	s, ok := object.ToGoString(val)
	if !ok {
		return 0, &TypeError{Expected: "string"}
	}

	_, err = buf.WriteString(s)

	n = utf8.RuneCountInString(s)

	return
}

func Format(format string, vals ...object.Value) (string, error) {
	if len(format) == 0 {
		return "", nil
	}

	var lasti int
	var termi int

	buf := new(bytes.Buffer)
	length := len(format)

	i := 0
	for {
		if format[i] == '%' {
			i++

			if i == length {
				return "", &OptionError{"%<\\0>"}
			}

			if format[i] == '%' {
				buf.WriteString(format[lasti:i])

				i++

				if i == length {
					return buf.String(), nil
				}

				lasti = i

				continue
			}

			if termi == len(vals) {
				return "", &ArgError{ExtraMsg: "no value", ArgNum: termi}
			}

			buf.WriteString(format[lasti : i-1])

			t := new(term)

			// parse flag
		flag:
			for {
				switch format[i] {
				case '+':
					t.plus = true
				case '-':
					t.minus = true
				case ' ':
					t.space = true
				case '0':
					t.zero = true
				default:
					break flag
				}

				i++
				if i == length {
					return "", &OptionError{"%<\\0>"}
				}
			}

			// parse width
			woffset := i
			for {
				if !('0' <= format[i] && format[i] <= '9') {
					t.width, _ = strconv.Atoi(format[woffset:i])
					break
				}

				i++
				if i == length {
					return "", &OptionError{"%<\\0>"}
				}
			}

			// parse prec
			if format[i] == '.' {
				i++
				if i == length {
					return "", &OptionError{"%<\\0>"}
				}

				poffset := i

				for {
					if !('0' <= format[i] && format[i] <= '9') {
						t.prec, _ = strconv.Atoi(format[poffset:i])
						break
					}

					i++
					if i == length {
						return "", &OptionError{"%<\\0>"}
					}
				}
			}

			// parse verb
			t.verb = format[i]

			// do format
			_, err := t.writeFormat(buf, vals[termi])
			if err != nil {
				switch err := err.(type) {
				case *ArgError:
					err.ArgNum = termi

					return "", err
				case *TypeError:
					err.ArgNum = termi

					return "", err
				case *OptionError:
					return "", err
				default:
					return "", err
				}
			}

			termi++

			i++
			if i == length {
				return buf.String(), nil
			}

			lasti = i
		} else {
			i++
			if i == length {
				buf.WriteString(format[lasti:i])

				return buf.String(), nil
			}
		}
	}

	return "", nil
}
