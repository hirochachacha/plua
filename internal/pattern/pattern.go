package pattern

import (
	"bytes"
	// "fmt"
	"strings"
)

type Pattern struct {
	m machine
}

func Find(pat string, input []byte) ([][2]int, error) {
	p, err := Compile(pat)
	if err != nil {
		return nil, err
	}
	return p.Find(input), nil
}

func FindString(pat string, input string) ([][2]int, error) {
	p, err := Compile(pat)
	if err != nil {
		return nil, err
	}
	return p.FindString(input), nil
}

func FindAll(pat string, input []byte) ([][][2]int, error) {
	p, err := Compile(pat)
	if err != nil {
		return nil, err
	}
	return p.FindAll(input), nil
}

func FindStringAll(pat, input string) ([][][2]int, error) {
	p, err := Compile(pat)
	if err != nil {
		return nil, err
	}
	return p.FindStringAll(input), nil
}

func Match(pat string, input []byte) ([][]byte, error) {
	p, err := Compile(pat)
	if err != nil {
		return nil, err
	}
	return p.Match(input), nil
}

func MatchString(pat, input string) ([]string, error) {
	p, err := Compile(pat)
	if err != nil {
		return nil, err
	}
	return p.MatchString(input), nil
}

func Replace(pat string, input, repl []byte, n int) ([]byte, error) {
	p, err := Compile(pat)
	if err != nil {
		return nil, err
	}
	return p.Replace(input, repl, n)
}

func ReplaceString(pat, input, repl string, n int) (string, error) {
	p, err := Compile(pat)
	if err != nil {
		return "", err
	}
	return p.ReplaceString(input, repl, n)
}

func ReplaceFunc(pat string, input []byte, repl func([]byte) []byte, n int) ([]byte, error) {
	p, err := Compile(pat)
	if err != nil {
		return nil, err
	}
	return p.ReplaceFunc(input, repl, n)
}

func ReplaceFuncString(pat, input string, repl func(string) string, n int) (string, error) {
	p, err := Compile(pat)
	if err != nil {
		return "", err
	}
	return p.ReplaceFuncString(input, repl, n)
}

func MatchAll(pat string, input []byte) ([][][]byte, error) {
	p, err := Compile(pat)
	if err != nil {
		return nil, err
	}
	return p.MatchAll(input), nil
}

func MatchStringAll(pat, input string) ([][]string, error) {
	p, err := Compile(pat)
	if err != nil {
		return nil, err
	}
	return p.MatchStringAll(input), nil
}

func Compile(pat string) (*Pattern, error) {
	c := newCompiler(&inputString{pat}, false)

	insts, err := c.compile()
	if err != nil {
		return nil, err
	}

	n := len(insts)

	m := machine{
		byteMatch:   false,
		insts:       insts,
		typ:         c.typ,
		prefix:      c.prefix,
		prefixBytes: c.prefixBytes,

		preds: upreds,

		sets: c.sets,

		ncaps: c.nparens + 1,

		current: newQueue(n),
		next:    newQueue(n),
	}

	return &Pattern{m}, nil
}

func CompileBytes(pat []byte, byteMatch bool) (*Pattern, error) {
	c := newCompiler(&inputBytes{pat}, byteMatch)

	insts, err := c.compile()
	if err != nil {
		return nil, err
	}

	n := len(insts)

	var preds predicates
	if byteMatch {
		preds = bpreds
	} else {
		preds = upreds
	}

	m := machine{
		byteMatch:   byteMatch,
		insts:       insts,
		typ:         c.typ,
		prefix:      c.prefix,
		prefixBytes: c.prefixBytes,

		preds: preds,

		sets: c.sets,

		ncaps: c.nparens + 1,

		current: newQueue(n),
		next:    newQueue(n),
	}

	return &Pattern{m}, nil
}

func (p *Pattern) Find(pat []byte) [][2]int {
	return p.m.find(&inputBytes{pat})
}

func (p *Pattern) FindString(pat string) [][2]int {
	return p.m.find(&inputString{pat})
}

func (p *Pattern) FindAll(pat []byte) [][][2]int {
	return p.m.findAll(&inputBytes{pat})
}

func (p *Pattern) FindStringAll(pat string) [][][2]int {
	return p.m.findAll(&inputString{pat})
}

func (p *Pattern) Match(pat []byte) [][]byte {
	indices := p.m.find(&inputBytes{pat})
	if indices == nil {
		return nil
	}
	ss := make([][]byte, len(indices))
	for i, index := range indices {
		ss[i] = pat[index[0]:index[1]]
	}
	return ss
}

func (p *Pattern) MatchString(pat string) []string {
	indices := p.m.find(&inputString{pat})
	if indices == nil {
		return nil
	}
	ss := make([]string, len(indices))
	for i, index := range indices {
		ss[i] = string(pat[index[0]:index[1]])
	}
	return ss
}

func (p *Pattern) MatchAll(pat []byte) [][][]byte {
	indiceses := p.m.findAll(&inputBytes{pat})

	if indiceses == nil {
		return nil
	}

	rets := make([][][]byte, len(indiceses))
	for i, indices := range indiceses {
		ret := make([][]byte, len(indices))
		for i, index := range indices {
			ret[i] = pat[index[0]:index[1]]
		}
		rets[i] = ret
	}

	return rets
}

func (p *Pattern) MatchStringAll(pat string) [][]string {
	indiceses := p.m.findAll(&inputString{pat})

	if indiceses == nil {
		return nil
	}

	rets := make([][]string, len(indiceses))
	for i, indices := range indiceses {
		ret := make([]string, len(indices))
		for i, index := range indices {
			ret[i] = string(pat[index[0]:index[1]])
		}
		rets[i] = ret
	}

	return rets
}

func (p *Pattern) Replace(pat, repl []byte, n int) ([]byte, error) {
	indiceses := p.m.findAll(&inputBytes{pat})

	if indiceses == nil {
		return pat, nil
	}

	isLiteral := bytes.IndexRune(repl, '%') == -1

	var buf bytes.Buffer

	var i int

	if isLiteral {
		for j, indices := range indiceses {
			if j == n {
				break
			}

			buf.Write(pat[i:indices[0][0]])

			buf.Write(repl)

			i = indices[0][1]
		}
	} else {
		for j, indices := range indiceses {
			if j == n {
				break
			}

			buf.Write(pat[i:indices[0][0]])

			var start int
			var end int
			for {
				end = bytes.IndexRune(repl[start:], '%')
				if end == -1 {
					buf.Write(repl[start:])
					break
				}
				end += start

				buf.Write(repl[start:end])

				if len(repl) == end {
					return nil, ErrMalformedPattern
				}

				d := repl[end+1]
				if !('1' <= d && d <= '9') {
					return nil, ErrInvalidCapture
				}

				i := int(d - '0')
				if i >= len(indices) {
					return nil, ErrInvalidCapture
				}

				buf.Write(pat[indices[i][0]:indices[i][1]])

				if len(repl) == end+1 {
					break
				}

				start = end + 2
			}

			i = indices[0][1]
		}
	}

	buf.Write(pat[i:])

	return buf.Bytes(), nil
}

func (p *Pattern) ReplaceString(pat, repl string, n int) (string, error) {
	indiceses := p.m.findAll(&inputString{pat})

	if indiceses == nil {
		return pat, nil
	}

	isLiteral := strings.IndexRune(repl, '%') == -1

	var buf bytes.Buffer

	var i int

	if isLiteral {
		for j, indices := range indiceses {
			if j == n {
				break
			}

			buf.WriteString(pat[i:indices[0][0]])

			buf.WriteString(repl)

			i = indices[0][1]
		}
	} else {
		for j, indices := range indiceses {
			if j == n {
				break
			}

			buf.WriteString(pat[i:indices[0][0]])

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
					return "", ErrMalformedPattern
				}

				d := repl[end+1]
				if !('1' <= d && d <= '9') {
					return "", ErrInvalidCapture
				}

				i := int(d - '0')
				if i >= len(indices) {
					return "", ErrInvalidCapture
				}

				buf.WriteString(pat[indices[i][0]:indices[i][1]])

				if len(repl) == end+1 {
					break
				}

				start = end + 2
			}

			i = indices[0][1]
		}
	}

	buf.WriteString(pat[i:])

	return buf.String(), nil
}

func (p *Pattern) ReplaceFunc(pat []byte, repl func([]byte) []byte, n int) ([]byte, error) {
	indiceses := p.m.findAll(&inputBytes{pat})

	if indiceses == nil {
		return pat, nil
	}

	var buf bytes.Buffer

	var i int

	for j, indices := range indiceses {
		if j == n {
			break
		}

		buf.Write(pat[i:indices[0][0]])

		buf.Write(repl(pat[indices[0][0]:indices[0][1]]))

		i = indices[0][1]
	}

	buf.Write(pat[i:])

	return buf.Bytes(), nil
}

func (p *Pattern) ReplaceFuncString(pat string, repl func(string) string, n int) (string, error) {
	indiceses := p.m.findAll(&inputString{pat})

	if indiceses == nil {
		return pat, nil
	}

	var buf bytes.Buffer

	var i int

	for j, indices := range indiceses {
		if j == n {
			break
		}

		buf.WriteString(pat[i:indices[0][0]])

		buf.WriteString(repl(pat[indices[0][0]:indices[0][1]]))

		i = indices[0][1]
	}

	buf.WriteString(pat[i:])

	return buf.String(), nil
}
