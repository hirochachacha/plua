package pattern

import (
	"errors"
)

var (
	errInvalidRange     = errors.New("invalid character range")
	errInvalidEscape    = errors.New("invalid escape")
	errInvalidBalance   = errors.New("invalid balance")
	errInvalidFrontier  = errors.New("invalid frontier")
	errInvalidCapture   = errors.New("invalid capture")
	errMissingBracket   = errors.New("missing closing ]")
	errUnexpectedParen  = errors.New("unexpected )")
	errUnexpectedRange  = errors.New("unexpected -")
	errMalformedPattern = errors.New("malformed pattern (ends with '%')")
)

type matchType uint8

const (
	matchBeginning matchType = 1 << iota
	matchEnd
	matchFullLiteral
	matchPrefixLiteral
)

type compiler struct {
	pat     input
	pos     int
	r       rune
	typ     matchType
	poffset int
	prefix  string
	sets    []*rangeTable
	nparens int
}

func newCompiler(pat input) *compiler {
	return &compiler{
		pat: pat,
	}
}

func (c *compiler) next() {
	r, i := c.pat.stepRune(c.pos)
	c.pos += i
	c.r = r
}

func (c *compiler) peek() rune {
	r, _ := c.pat.stepRune(c.pos)

	return r
}

// return escaped, isSingle
func (c *compiler) escape(r rune) (rune, bool) {
	switch r {
	case 'a':
		return opLetter, true
	case 'A':
		return opNotLetter, true
	case 'c':
		return opControl, true
	case 'C':
		return opNotControl, true
	case 'd':
		return opDigit, true
	case 'D':
		return opNotDigit, true
	case 'g':
		return opGraphic, true
	case 'G':
		return opNotGraphic, true
	case 'l':
		return opLower, true
	case 'L':
		return opNotLower, true
	case 'p':
		return opPunct, true
	case 'P':
		return opNotPunct, true
	case 's':
		return opSpace, true
	case 'S':
		return opNotSpace, true
	case 'u':
		return opUpper, true
	case 'U':
		return opNotUpper, true
	case 'w':
		return opAlphaNum, true
	case 'W':
		return opNotAlphaNum, true
	case 'x':
		return opHexDigit, true
	case 'X':
		return opNotHexDigit, true
	case 'b':
		return r, false
	case 'f':
		return r, false
	}

	if '1' <= r && r <= '9' {
		return r, false
	}

	return r, true
}

// op: rune
func (c *compiler) instChar(r rune) inst {
	return inst{
		op: r,
	}
}

func (c *compiler) instAny() inst {
	return inst{
		op: opAny,
	}
}

// op: rune or opAny
// x: next instruction
func (c *compiler) instSingle(r rune, isEscaped bool) inst {
	if !isEscaped && r == '.' {
		return inst{
			op: opAny,
		}
	}
	return inst{
		op: r,
	}
}

func (c *compiler) instCapture(x int) inst {
	return inst{
		op: opCapture,
		x:  x,
	}
}

func (c *compiler) instBalanceAny(x, y rune) inst {
	return inst{
		op: opBalanceAny,
		x:  int(x),
		y:  int(y),
	}
}

// x : next instruction
func (c *compiler) instBalanceUp(x int) inst {
	return inst{
		op: opBalanceUp,
		x:  x,
	}
}

// x : next instruction
// y : next instruction if balance == 0
func (c *compiler) instBalanceDown(x, y int) inst {
	return inst{
		op: opBalanceDown,
		x:  x,
		y:  y,
	}
}

// x : index of sets
func (c *compiler) instRange(x int) (ins inst, err error) {
	c.next()

	set := &rangeTable{make([][2]rune, 0)}

	op := opRange

	if c.r == '^' {
		op = opNotRange

		c.next()
	}

	// fetch until ']'
	var isSingle bool
L2:
	for {
		switch c.r {
		case endOfText:
			err = errMissingBracket
			return
		case ']':
			break L2
		case '-':
			err = errUnexpectedRange
			return
		case '%':
			c.next()

			if c.r == endOfText {
				err = errInvalidEscape
				return
			}

			c.r, isSingle = c.escape(c.r)
			if !isSingle {
				err = errInvalidEscape
				return
			}
		default:
			if c.peek() == '-' {
				r := c.r

				c.next()
				c.next()

				switch c.r {
				case ']':
					err = errInvalidRange
					return
				case endOfText:
					err = errMissingBracket
					return
				default:
					set.r32 = append(set.r32, [2]rune{r, c.r})
				}
			} else {
				set.r32 = append(set.r32, [2]rune{c.r, c.r})

				c.next()
			}
		}
	}

	ins = inst{
		op: op,
		x:  x,
	}

	c.sets = append(c.sets, set)

	return
}

// x: next instruction, higher precedence than y
// y: next instruction
func (c *compiler) instSplit(x, y int) inst {
	return inst{
		op: opSplit,
		x:  x,
		y:  y,
	}
}

// x: next instruction
func (c *compiler) instJmp(x int) inst {
	return inst{
		op: opJmp,
		x:  x,
	}
}

// x: capture depth
func (c *compiler) instEnterSave(x int) inst {
	return inst{
		op: opEnterSave,
		x:  x,
	}
}

// x: capture depth
func (c *compiler) instExitSave(x int) inst {
	return inst{
		op: opExitSave,
		x:  x,
	}
}

func (c *compiler) instMatch() inst {
	return inst{
		op: opMatch,
	}
}

func (c *compiler) makeRepetition(ins inst, ninsts int) []inst {
	switch c.peek() {
	case '+':
		l1 := ninsts
		l3 := ninsts + 2

		ins1 := ins
		ins2 := c.instSplit(l1, l3)

		c.next()

		return []inst{ins1, ins2}
	case '-':
		l1 := ninsts
		l2 := ninsts + 1
		l4 := ninsts + 3

		ins1 := c.instSplit(l4, l2)
		ins2 := ins
		ins3 := c.instJmp(l1)

		c.next()

		return []inst{ins1, ins2, ins3}
	case '*':
		l1 := ninsts
		l2 := ninsts + 1
		l4 := ninsts + 3

		ins1 := c.instSplit(l2, l4)
		ins2 := ins
		ins3 := c.instJmp(l1)

		c.next()

		return []inst{ins1, ins2, ins3}
	case '?':
		l1 := ninsts + 1
		l2 := ninsts + 2

		ins1 := c.instSplit(l1, l2)
		ins2 := ins

		c.next()

		return []inst{ins1, ins2}
	}

	return []inst{ins}
}

func (c *compiler) makeBalance(x, y rune, ninsts int) []inst {
	l3 := ninsts + 2
	l4 := ninsts + 3
	l6 := ninsts + 5
	l7 := ninsts + 6
	l9 := ninsts + 8
	l11 := ninsts + 10

	ins1 := c.instChar(x)
	ins2 := c.instBalanceUp(l3)
	ins3 := c.instSplit(l4, l6)
	ins4 := c.instBalanceAny(x, y)
	ins5 := c.instJmp(l3)
	ins6 := c.instSplit(l9, l7)
	ins7 := c.instChar(x)
	ins8 := c.instBalanceUp(l3)
	ins9 := c.instChar(y)
	ins10 := c.instBalanceDown(l3, l11)

	c.next()

	return []inst{ins1, ins2, ins3, ins4, ins5, ins6, ins7, ins8, ins9, ins10}
}

func (c *compiler) compile() (insts []inst, err error) {
	insts = make([]inst, 0, c.pat.length()+1)

	isPrefix := true

	isEscaped := false

	c.next()

	if c.r == '^' {
		c.typ |= matchBeginning

		c.next()
	}

	var isSingle bool
	var depth int // paren depth
L:
	for {
		switch c.r {
		case endOfText:
			break L
		case '$':
			if c.peek() == endOfText {
				c.typ |= matchEnd

				break L
			}

			// err = ErrUnexpectedCharAfterEnd
			// return

		// case '+', '-', '*', '?':
		// err = ErrMissingRepeatArgument
		case '(':
			isPrefix = false

			ins := c.instEnterSave(depth)
			insts = append(insts, ins)

			c.next()

			depth++

			continue
		case ')':
			isPrefix = false

			depth--
			if depth < 0 {
				err = errUnexpectedParen
				return
			}
			c.nparens++

			ins := c.instExitSave(depth)
			insts = append(insts, ins)

			c.next()

			continue
		// case ']':
		// err = ErrUnexpectedBracket
		// return
		case '[':
			isPrefix = false

			ins, err := c.instRange(len(c.sets))
			if err != nil {
				return nil, err
			}

			insts = append(insts, c.makeRepetition(ins, len(insts))...)

			c.next()

			continue
		case '%':
			isPrefix = false

			c.next()

			if c.r == endOfText {
				err = errInvalidEscape
				return
			}

			c.r, isSingle = c.escape(c.r)

			if !isSingle {
				switch c.r {
				case 'b':
					c.next()

					if c.r == endOfText {
						err = errInvalidBalance
						return
					}

					x := c.r

					c.next()

					if c.r == endOfText {
						err = errInvalidBalance
						return
					}

					y := c.r

					insts = append(insts, c.makeBalance(x, y, len(insts))...)
				case 'f':
					c.next()

					if c.r != '[' {
						err = errInvalidFrontier
						return
					}

					ins, err := c.instRange(len(c.sets))
					if err != nil {
						return nil, err
					}

					ins.op = opFrontier

					insts = append(insts, ins)
				default:
					x := int(c.r - '0')
					if x > c.nparens {
						err = errInvalidCapture
						return
					}

					ins := c.instCapture(x)
					insts = append(insts, ins)

					c.next()
				}

				c.next()

				continue
			}

			isEscaped = true
		}

		ins := c.instSingle(c.r, isEscaped)
		if ins.op == opAny {
			isPrefix = false
		}

		isEscaped = false

		_insts := c.makeRepetition(ins, len(insts))

		if len(_insts) > 1 {
			isPrefix = false
		}

		if isPrefix {
			c.poffset = c.pos
		} else {
			insts = append(insts, _insts...)
		}

		c.next()
	}

	// if depth != 0 {
	// err = ErrMissingParen
	// return
	// }

	// close open parens
	for i := depth; i > 0; i-- {
		c.nparens++

		ins := c.instExitSave(depth)
		insts = append(insts, ins)
	}

	if isPrefix {
		c.typ |= matchFullLiteral

		if c.typ&matchBeginning != 0 {
			c.prefix = c.pat.slice(1, c.poffset)
		} else {
			c.prefix = c.pat.slice(0, c.poffset)
		}

		return
	} else if c.poffset != 0 {
		c.typ |= matchPrefixLiteral

		if c.typ&matchBeginning != 0 {
			c.prefix = c.pat.slice(1, c.poffset)
		} else {
			c.prefix = c.pat.slice(0, c.poffset)
		}
	}

	insMatch := c.instMatch()

	insts = append(insts, insMatch)

	return
}

/*

pattern : fragments

fragments : fragment | fragments fragment

fragment : repeat | capture

capture : '(' fragments ')'

repeat : single
       | single '*'
       | single '+'
       | single '-'

single : '.' | char | class

class : '%a' | '%c'

*/
