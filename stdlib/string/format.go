package string

import (
	"bytes"
	"fmt"
	"math"
	"strings"

	"github.com/hirochachacha/plua/internal/strconv"
	"github.com/hirochachacha/plua/object"
	"github.com/hirochachacha/plua/object/fnutil"
)

func format(th object.Thread, args ...object.Value) ([]object.Value, *object.RuntimeError) {
	ap := fnutil.NewArgParser(th, args)

	s, err := ap.ToGoString(0)
	if err != nil {
		return nil, err
	}

	b := &buffer{th: th, ap: ap}

	argn := 1

	for i := 0; i < len(s); i++ {
		c := s[i]

		if c != '%' {
			e := b.WriteByte(c)
			if e != nil {
				return nil, object.NewRuntimeError(e.Error())
			}
			continue
		}

		i++
		if i == len(s) {
			break
		}

		c = s[i]

		if c == '%' {
			e := b.WriteByte(c)
			if e != nil {
				return nil, object.NewRuntimeError(e.Error())
			}
			continue
		}

		if argn == len(args) {
			return nil, ap.ArgError(argn, "no value")
		}

		t, j := parseTerm(s[i:])
		if t == nil {
			return nil, ap.ArgError(0, "invalid format")
		}

		err := b.writeFormat(t, argn)
		if err != nil {
			return nil, err
		}

		i += j - 1

		argn++
	}

	return []object.Value{object.String(b.String())}, nil
}

type buffer struct {
	bytes.Buffer

	th object.Thread
	ap *fnutil.ArgParser
}

func (b *buffer) writeFormat(t *term, argn int) *object.RuntimeError {
	prefix, verb, err := b.verb(t, argn)
	if err != nil {
		return err
	}

	padSize := t.width - len(verb) - len(prefix)

	if t.minus {
		if len(prefix) != 0 {
			b.WriteString(prefix)
		}
		b.WriteString(verb)
		if t.zero {
			for i := 0; i < padSize; i++ {
				b.WriteByte('0')
			}
		} else {
			for i := 0; i < padSize; i++ {
				b.WriteByte(' ')
			}
		}
	} else {
		if t.zero {
			if len(prefix) != 0 {
				b.WriteString(prefix)
			}
			for i := 0; i < padSize; i++ {
				b.WriteByte('0')
			}
		} else {
			for i := 0; i < padSize; i++ {
				b.WriteByte(' ')
			}
			if len(prefix) != 0 {
				b.WriteString(prefix)
			}
		}
		b.WriteString(verb)
	}

	return nil
}

func (b *buffer) verb(t *term, argn int) (string, string, *object.RuntimeError) {
	switch t.verb {
	case 'c':
		return t.byte(b.ap, argn)
	case 'd', 'i', 'u', 'o', 'x', 'X':
		return t.int(b.ap, argn)
	case 'e', 'E', 'f', 'g', 'G':
		return t.float(b.ap, argn)
	case 'a', 'A':
		return t.hexFloat(b.ap, argn)
	case 'q':
		return t.quoteString(b.ap, argn)
	case 's':
		return t.string(b.th, b.ap, argn)
	}

	return "", "", b.ap.OptionError(0, string(t.verb))
}

type term struct {
	verb byte

	width int
	prec  int

	plus  bool
	minus bool
	space bool
	zero  bool
}

func parseTerm(s string) (*term, int) {
	t := &term{
		width: -1,
		prec:  -1,
	}

	var i int

	// parse flags
parseFlags:
	for ; i < len(s); i++ {
		switch s[i] {
		case '+':
			t.plus = true
		case '-':
			t.minus = true
		case ' ':
			t.space = true
		case '0':
			t.zero = true
		default:
			break parseFlags
		}
	}

	// parse width
	j := i
	for ; i < len(s); i++ {
		if !('0' <= s[i] && s[i] <= '9') {
			t.width, _ = strconv.Atoi(s[j:i])

			break
		}
	}

	// parse prec
	if i < len(s) && s[i] == '.' {
		i++
		j = i
		for ; i < len(s); i++ {
			if !('0' <= s[i] && s[i] <= '9') {
				t.prec, _ = strconv.Atoi(s[j:i])

				break
			}
		}
	}

	if i == len(s) {
		return nil, 0
	}

	// parse verb
	t.verb = s[i]

	i++

	return t, i
}

func (t *term) byte(ap *fnutil.ArgParser, argn int) (prefix, s string, err *object.RuntimeError) {
	i, err := ap.ToGoInt64(argn)
	if err != nil {
		return "", "", err
	}
	return "", string(byte(i)), nil
}

func (t *term) int(ap *fnutil.ArgParser, argn int) (prefix, s string, err *object.RuntimeError) {
	i, err := ap.ToGoInt64(argn)
	if err != nil {
		return "", "", err
	}

	var base int
	var toUpper bool
	var toUint bool

	switch t.verb {
	case 'i', 'd':
		base = 10
	case 'u':
		base = 10
		toUint = true
	case 'o':
		base = 8
		toUint = true
	case 'x':
		base = 16
		toUint = true
	case 'X':
		base = 16
		toUpper = true
		toUint = true
	default:
		panic("unreachable")
	}

	if toUint {
		s = strconv.FormatUint(uint64(i), base)
	} else {
		s = strconv.FormatInt(i, base)

		if s[0] == '-' {
			s = s[1:]
			prefix = "-"
		} else if t.plus {
			prefix = "+"
		}
	}

	if toUpper {
		s = strings.ToUpper(s)
	}

	var prec string

	if 0 < t.prec-len(s) {
		prec = strings.Repeat("0", t.prec-len(s))
	}

	return prefix, prec + s, nil
}

func (t *term) float(ap *fnutil.ArgParser, argn int) (prefix, s string, err *object.RuntimeError) {
	f, err := ap.ToGoFloat64(argn)
	if err != nil {
		return "", "", err
	}

	s = strconv.FormatFloat(f, t.verb, t.prec, 64)

	if s[0] == '-' {
		s = s[1:]
		prefix = "-"
	} else if t.plus {
		prefix = "+"
	}

	return prefix, s, nil
}

func (t *term) hexFloat(ap *fnutil.ArgParser, argn int) (prefix, s string, err *object.RuntimeError) {
	f, err := ap.ToGoFloat64(argn)
	if err != nil {
		return "", "", err
	}

	u := math.Float64bits(f)

	signBit := int(u >> 63)

	if f == 0 {
		if t.prec > 0 {
			s = "0." + strings.Repeat("0", t.prec) + "+0"
		} else {
			s = "0p+0"
		}
	} else {
		exponent := int64(u>>52&0x7ff) - 1023
		fraction := u & 0xfffffffffffff

		if t.prec > 0 {
			// TODO precision support

			s = fmt.Sprintf("1.%xp%+d", fraction, exponent)
		} else {
			s = fmt.Sprintf("1.%xp%+d", fraction, exponent)
		}
	}

	switch t.verb {
	case 'a':
		if signBit == 1 {
			prefix = "-0x"
		} else if t.plus {
			prefix = "+0x"
		} else {
			prefix = "0x"
		}
	case 'A':
		if signBit == 1 {
			prefix = "-0X"
		} else if t.plus {
			prefix = "+0X"
		} else {
			prefix = "0X"
		}
		s = strings.ToUpper(s)
	default:
		panic("unreachable")
	}

	return prefix, s, nil
}

func (t *term) quoteString(ap *fnutil.ArgParser, argn int) (prefix, s string, err *object.RuntimeError) {
	val, _ := ap.Get(argn)

	switch val := val.(type) {
	case nil:
		s = object.Repr(val)
	case object.Boolean:
		s = object.Repr(val)
	case object.String:
		s = strconv.Quote(string(val))
	case object.Integer:
		s = strconv.FormatInt(int64(val), 10)
	case object.Number:
		s = strconv.FormatFloat(float64(val), 'f', -1, 64)
	default:
		return "", "", ap.ArgError(argn, "value has no literal form")
	}

	return "", s, nil
}

func (t *term) string(th object.Thread, ap *fnutil.ArgParser, argn int) (prefix, s string, err *object.RuntimeError) {
	val, _ := ap.Get(argn)

	if fn := th.GetMetaField(val, "__tostring"); fn != nil {
		rets, err := th.Call(fn, nil)
		if err != nil {
			return "", "", err
		}

		if len(rets) == 0 {
			return "", "", object.NewRuntimeError("'tostring' must return a string to 'print'")
		}

		val = rets[0]
	}

	s = object.Repr(val)

	if 0 <= t.prec && t.prec < len(s) {
		s = s[:t.prec]
	}

	return "", s, nil
}
