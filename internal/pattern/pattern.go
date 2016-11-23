package pattern

import (
	"bytes"
	"errors"

	"github.com/hirochachacha/plua/object"
)

/*

pattern    : '^'? fragment* '$'?
fragment   : repetition | capture | captured | frontier | balance
repetition : single
           | single '*'
           | single '+'
           | single '-'
           | single '?'
capture    : '(' fragment* ')'
captured   : '%' <digit>
frontier   : '%f' set
balance    : '%b' char char
single     : simple | set
simple     : char | class
set        : '[' '^'? set-item+ ']'
set-item   : range | simple
range      : char '-' char
char       : dot | <character excluding ".()%[]"> | '%' <non-alphanumeric character>
dot        : '.'
class      : '%a' | '%c' | ...

*/

var decodeRune = _decodeByte
var lastDecodeRune = _lastDecodeByte

// var decodeRune = _decodeRune
// var lastDecodeRune = _lastDecodeRune

var (
	errInvalidEscape    = errors.New("invalid escape")
	errInvalidBalance   = errors.New("invalid balance")
	errInvalidFrontier  = errors.New("invalid frontier")
	errMissingBracket   = errors.New("missing closing ]")
	errUnexpectedParen  = errors.New("unexpected )")
	errMalformedPattern = errors.New("malformed pattern (ends with '%')")
)

type matchType int

const (
	matchPrefix matchType = 1 << iota
	matchSuffix
)

const (
	eos = -1
	sos = -2
)

type Capture struct {
	Begin   int
	End     int
	IsEmpty bool
}

func (cap Capture) Value(input string) object.Value {
	if cap.IsEmpty {
		return object.Integer(cap.Begin + 1)
	}
	return object.String(input[cap.Begin:cap.End])
}

// IR for input pattern
type Pattern struct {
	typ    matchType
	prefix string
	code   []instruction
	sets   []*set
	nsaved int
}

func (p *Pattern) FindIndex(input string, off int) []Capture {
	if len(input) < off {
		return nil
	}
	m := &machine{
		saved: make([]Capture, p.nsaved),
	}
	if m.match(p, input, off) {
		return m.saved
	}
	return nil
}

func (p *Pattern) FindAllIndex(input string, n int) [][]Capture {
	if n == 0 {
		return nil
	}

	var rets [][]Capture

	if n > 0 {
		rets = make([][]Capture, 0, n)
	}

	m := &machine{
		saved: make([]Capture, p.nsaved),
	}

	prev := -1

	off := 0
	for {
		if !m.match(p, input, off) {
			break
		}

		loc := m.saved[0]

		off = loc.End
		if off == loc.Begin {
			off++
			if loc.Begin == prev {
				continue
			}
		}

		rets = append(rets, dup(m.saved))

		if len(rets) == n {
			break
		}

		prev = loc.End
	}

	return rets
}

func (p *Pattern) ReplaceAllFunc(input string, repl func([]Capture) (string, error), n int) (string, int, error) {
	allCaptures := p.FindAllIndex(input, n)
	if allCaptures == nil {
		return input, 0, nil
	}

	var buf bytes.Buffer

	last := 0
	for _, caps := range allCaptures {
		loc := caps[0]

		buf.WriteString(input[last:loc.Begin])

		alt, err := repl(caps)
		if err != nil {
			return "", -1, err
		}

		buf.WriteString(alt)

		last = loc.End
	}

	buf.WriteString(input[last:])

	return buf.String(), len(allCaptures), nil
}

func dup(caps []Capture) []Capture {
	return append([]Capture{}, caps...)
}

func FindIndex(input, pat string, off int) ([]Capture, error) {
	if len(input) < off {
		return nil, nil
	}
	p, err := Compile(pat)
	if err != nil {
		return nil, err
	}
	return p.FindIndex(input, off), nil
}

func FindAllIndex(input, pat string, n int) ([][]Capture, error) {
	p, err := Compile(pat)
	if err != nil {
		return nil, err
	}
	return p.FindAllIndex(input, n), nil
}

func ReplaceAllFunc(input, pat string, repl func([]Capture) (string, error), n int) (string, int, error) {
	p, err := Compile(pat)
	if err != nil {
		return "", -1, err
	}
	return p.ReplaceAllFunc(input, repl, n)
}
