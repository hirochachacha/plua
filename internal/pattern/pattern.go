package pattern

import (
	"bytes"
	"strings"
)

type MatchString struct {
	Item     string
	Captures []string
}

type MatchRange struct {
	Item     Range
	Captures []Range
}

func (r *MatchRange) MatchString(s string) *MatchString {
	caps := make([]string, len(r.Captures))
	for i, cr := range r.Captures {
		caps[i] = s[cr.Begin:cr.End]
	}

	ms := &MatchString{
		Item:     string(s[r.Item.Begin:r.Item.End]),
		Captures: caps,
	}

	return ms
}

type Range struct {
	Begin int
	End   int
}

type Pattern struct {
	m machine
}

func Find(pat string, input string) (*MatchRange, error) {
	p, err := Compile(pat)
	if err != nil {
		return nil, err
	}
	return p.Find(input), nil
}

func FindAll(pat, input string) ([]*MatchRange, error) {
	p, err := Compile(pat)
	if err != nil {
		return nil, err
	}
	return p.FindAll(input), nil
}

func Match(pat, input string) (*MatchString, error) {
	p, err := Compile(pat)
	if err != nil {
		return nil, err
	}
	return p.Match(input), nil
}

func Replace(pat, input, repl string, n int) (string, int, error) {
	p, err := Compile(pat)
	if err != nil {
		return "", -1, err
	}
	return p.Replace(input, repl, n)
}

func ReplaceFunc(pat, input string, repl func(ss ...string) string, n int) (string, int, error) {
	p, err := Compile(pat)
	if err != nil {
		return "", -1, err
	}
	return p.ReplaceFunc(input, repl, n)
}

func MatchAll(pat, input string) ([]*MatchString, error) {
	p, err := Compile(pat)
	if err != nil {
		return nil, err
	}
	return p.MatchAll(input), nil
}

func Compile(pat string) (*Pattern, error) {
	c := newCompiler(pat)

	insts, err := c.compile()
	if err != nil {
		return nil, err
	}

	n := len(insts)

	m := machine{
		code:    insts,
		typ:     c.typ,
		literal: c.literal,

		preds: upreds,

		sets: c.sets,

		ncaps: c.nparens + 1,

		current: newQueue(n),
		next:    newQueue(n),
	}

	return &Pattern{m}, nil
}

func (p *Pattern) Find(input string) *MatchRange {
	return p.m.find(input)
}

func (p *Pattern) FindAll(input string) []*MatchRange {
	return p.m.findAll(input)
}

func (p *Pattern) Match(input string) *MatchString {
	r := p.m.find(input)
	if r == nil {
		return nil
	}
	return r.MatchString(input)
}

func (p *Pattern) MatchAll(input string) []*MatchString {
	rs := p.m.findAll(input)
	if rs == nil {
		return nil
	}

	ms := make([]*MatchString, len(rs))
	for i, r := range rs {
		ms[i] = r.MatchString(input)
	}
	return ms
}

func (p *Pattern) Replace(input, repl string, n int) (string, int, error) {
	rs := p.m.findAll(input)
	if rs == nil {
		return input, 0, nil
	}

	isLiteral := strings.IndexRune(repl, '%') == -1

	var buf bytes.Buffer

	var off int

	if isLiteral {
		for j, r := range rs {
			if j == n {
				break
			}

			buf.WriteString(input[off:r.Item.Begin])

			buf.WriteString(repl)

			off = r.Item.End
		}
	} else {
		for j, r := range rs {
			if j == n {
				break
			}

			buf.WriteString(input[off:r.Item.Begin])

			var start int
			var end int
			for {
				end = strings.IndexRune(repl[start:], '%')
				if end == -1 {
					buf.WriteString(repl[start:])
					break
				}
				end += start

				buf.WriteString(repl[start:end])

				if len(repl) == end {
					return "", -1, errMalformedPattern
				}

				d := repl[end+1]
				if !('0' <= d && d <= '9') {
					return "", -1, errInvalidCapture
				}

				i := int(d - '0')
				if i == 0 {
					buf.WriteString(input[r.Item.Begin:r.Item.End])
				} else {
					if i > len(r.Captures) {
						return "", -1, errInvalidCapture
					}

					buf.WriteString(input[r.Captures[i-1].Begin:r.Captures[i-1].End])
				}

				if len(repl) == end+1 {
					break
				}

				start = end + 2
			}

			off = r.Item.End
		}
	}

	buf.WriteString(input[off:])

	return buf.String(), len(rs), nil
}

func (p *Pattern) ReplaceFunc(input string, repl func(ss ...string) string, n int) (string, int, error) {
	rs := p.m.findAll(input)
	if rs == nil {
		return input, 0, nil
	}

	var buf bytes.Buffer

	var off int

	for j, r := range rs {
		if j == n {
			break
		}

		buf.WriteString(input[off:r.Item.Begin])

		m := r.MatchString(input)

		if len(r.Captures) == 0 {
			buf.WriteString(repl(m.Item))
		} else {
			buf.WriteString(repl(m.Captures...))
		}

		off = r.Item.End
	}

	buf.WriteString(input[off:])

	return buf.String(), len(rs), nil
}
