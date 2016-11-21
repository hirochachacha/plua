package pattern

import (
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
	errInvalidCapture   = errors.New("invalid capture")
	errMissingBracket   = errors.New("missing closing ]")
	errUnexpectedParen  = errors.New("unexpected )")
	errMalformedPattern = errors.New("malformed pattern (ends with '%')")
)

type matchType int

const (
	matchPrefix matchType = 1 << iota
	matchSuffix
)

const eos = -1

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
