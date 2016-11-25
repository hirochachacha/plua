package pattern

import "fmt"

type scanState struct {
	pat   string
	off   int
	rsize int
	r     rune
}

func (s *scanState) next() {
	s.r, s.rsize = decodeRune(s.pat, s.off)
	s.off += s.rsize
}

func (s *scanState) scanSet() (*set, error) {
	set := new(set)

	if s.r == '^' {
		set.not = true

		s.next()
	}

L:
	for {
		switch s.r {
		case eos:
			return nil, errMalformedSet
		case '%':
			s.next()

			inst, ok := instClassOrEscChar(s.r)
			if !ok {
				if s.r == eos {
					return nil, errMalformedSet
				}
				return nil, errInvalidEscape
			}

			set.elems = append(set.elems, rune(inst.op))

			s.next()

			continue L
		case ']':
			if len(set.elems) != 0 || len(set.ranges) != 0 {
				break L
			}
		}

		low := s.r

		s.next()

		if s.r == '-' {
			s.next()

			hi := s.r

			switch hi {
			case ']':
				set.elems = append(set.elems, low, '-')

				break L
			case eos:
				return nil, errMalformedSet
			default:
				set.ranges = append(set.ranges, r32{low: low, hi: hi})
			}
		} else {
			set.elems = append(set.elems, low)
		}
	}

	return set, nil
}

func Compile(pat string) (*Pattern, error) {
	if pat == "" {
		return &Pattern{nsaved: 1}, nil
	}

	p := &Pattern{
		code: make([]instruction, 0, len(pat)*2),
	}

	if pat[0] == '^' {
		p.typ |= matchPrefix
		pat = pat[1:]
	}

	s := &scanState{
		pat: pat,
	}

	s.next()

	parenDepth := 0
	lparenNum := 1
	nsaved := 1

	parens := make(map[int]int) // depth: unclosed paren num

L:
	for {
		single := false

		switch s.r {
		case eos:
			break L
		case '$':
			s.next()

			if s.r == eos {
				p.typ |= matchSuffix
			} else {
				p.code = append(p.code, instChar('$'))

				single = true
			}
		case '(':
			p.code = append(p.code, instBeginSave(lparenNum))

			s.next()

			parens[parenDepth] = lparenNum

			parenDepth++
			lparenNum++

			if s.r == ')' { // handle empty paren
				parenDepth--
				if parenDepth < 0 {
					return nil, errInvalidPatternCapture
				}

				p.code = append(p.code, instEndSave(parens[parenDepth], 1))

				s.next()

				nsaved++
			}
		case ')':
			parenDepth--
			if parenDepth < 0 {
				return nil, errInvalidPatternCapture
			}

			p.code = append(p.code, instEndSave(parens[parenDepth], 0))

			s.next()

			nsaved++
		case '[':
			s.next()

			set, err := s.scanSet()
			if err != nil {
				return nil, err
			}

			p.code = append(p.code, instSet(len(p.sets)))

			p.sets = append(p.sets, set)

			s.next()

			single = true
		case '.':
			p.code = append(p.code, instAny())

			s.next()

			single = true
		case '%':
			s.next()

			if inst, ok := instClassOrEscChar(s.r); ok {
				p.code = append(p.code, inst)

				s.next()

				single = true
			} else {
				switch s.r {
				case eos:
					return nil, errMalformedEscape
				case 'f':
					s.next()

					if s.r != '[' {
						return nil, errIncompleteFrontier
					}

					s.next()

					set, err := s.scanSet()
					if err != nil {
						return nil, err
					}

					p.code = append(p.code, instFrontier(len(p.sets)))

					p.sets = append(p.sets, set)

					s.next()
				case 'b':
					s.next()

					if s.r == eos {
						return nil, errMalformedBalance
					}

					x := int(s.r)

					s.next()

					if s.r == eos {
						return nil, errMalformedBalance
					}

					y := int(s.r)

					p.code = append(p.code, instBalance(x, y))

					s.next()
				default:
					if !isDigit(s.r) {
						return nil, errInvalidEscape
					}

					n := int(s.r - '0')
					if n <= 0 || n >= nsaved {
						return nil, fmt.Errorf("invalid capture index %%%d", n)
					}

					p.code = append(p.code, instCapture(n))

					s.next()
				}
			}
		default:
			p.code = append(p.code, instChar(s.r))

			s.next()

			single = true
		}

		if single { // try to parse repetition
			switch s.r {
			case '*':
				// e*	L1: split L2, L4
				// 		L2: codes for e
				// 		L3:	jmp L1
				// 		L4:

				inst := p.code[len(p.code)-1]
				code := p.code[:len(p.code)-1]

				L1 := len(code)
				L2 := len(code) + 1
				L4 := len(code) + 3

				i1 := instSplit(L2, L4)
				i2 := inst
				i3 := instJmp(L1)

				p.code = append(code, i1, i2, i3)

				s.next()
			case '+':
				// e+	L1: codes for e
				// 		L2: split L1, L3
				// 		L3:

				inst := p.code[len(p.code)-1]
				code := p.code[:len(p.code)-1]

				L1 := len(code)
				L3 := len(code) + 2

				i1 := inst
				i2 := instSplit(L1, L3)

				p.code = append(code, i1, i2)

				s.next()
			case '-':
				inst := p.code[len(p.code)-1]
				code := p.code[:len(p.code)-1]

				L1 := len(code)
				L2 := len(code) + 1
				L4 := len(code) + 3

				i1 := instSplit(L4, L2)
				i2 := inst
				i3 := instJmp(L1)

				p.code = append(code, i1, i2, i3)

				s.next()
			case '?':
				// e?	L1: split L2, L3
				// 		L2: codes for e
				// 		L3:

				inst := p.code[len(p.code)-1]
				code := p.code[:len(p.code)-1]

				L2 := len(code) + 1
				L3 := len(code) + 2

				i1 := instSplit(L2, L3)
				i2 := inst

				p.code = append(code, i1, i2)

				s.next()
			}
		}
	}

	if parenDepth != 0 {
		return nil, errUnfinishedCapture
	}

	p.nsaved = nsaved

	if len(p.code) != 0 || p.typ&matchSuffix != 0 {
		p.code = append(p.code, instMatch())
	}

	return p, nil
}
