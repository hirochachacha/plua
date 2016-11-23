package pattern

import "strings"

type machine struct {
	saved []Capture
}

func (m *machine) match(p *Pattern, input string, base int) bool {
	if len(p.prefix) > len(input)-base {
		return false
	}

	if p.typ&matchPrefix != 0 {
		if strings.HasPrefix(input[base:], p.prefix) {
			return m.matchThere(p, input, base)
		}
		return false
	}

	for off := base; off <= len(input)-len(p.prefix); off++ {
		if strings.HasPrefix(input[off:], p.prefix) {
			if m.matchThere(p, input, off) {
				return true
			}
		}
	}

	return false
}

func (m *machine) matchThere(p *Pattern, input string, off int) bool {
	m.saved[0].Begin = off

	if len(p.code) == 0 {
		m.saved[0].End = off + len(p.prefix)
		return true
	}

	return m.recursiveMatch(p, input, 0, off+len(p.prefix))
}

func (m *machine) recursiveMatch(p *Pattern, input string, pc, sp int) bool {
	for {
		inst := p.code[pc]

		switch inst.op {
		case opMatch:
			if p.typ&matchSuffix != 0 {
				r, _ := decodeRune(input, sp)
				if r == eos {
					m.saved[0].End = sp
					return true
				}
				return false
			}

			m.saved[0].End = sp
			return true
		case opJmp:
			pc = inst.x
		case opSplit:
			if m.recursiveMatch(p, input, inst.x, sp) {
				return true
			}
			pc = inst.y
		case opBeginSave:
			cap := &m.saved[inst.x]
			old := cap.Begin
			cap.Begin = sp
			if m.recursiveMatch(p, input, pc+1, sp) {
				return true
			}
			cap.Begin = old
			return false
		case opEndSave:
			cap := &m.saved[inst.x]
			old := cap.End
			pred := cap.IsEmpty
			cap.End = sp
			cap.IsEmpty = inst.y != 0
			if m.recursiveMatch(p, input, pc+1, sp) {
				return true
			}
			cap.End = old
			cap.IsEmpty = pred
			return false
		case opCapture:
			loc := m.saved[inst.x]
			cap := input[loc.Begin:loc.End]

			if !strings.HasPrefix(input[sp:], cap) {
				return false
			}

			pc++
			sp += len(cap)
		case opSet:
			r, rsize := decodeRune(input, sp)
			if r == eos {
				return false
			}

			if !p.sets[inst.x].match(r) {
				return false
			}

			pc++
			sp += rsize
		case opBalance:
			x := rune(inst.x)
			y := rune(inst.y)

			r, rsize := decodeRune(input, sp)
			if r == eos {
				return false
			}

			if x != r {
				return false
			}

			sp += rsize

			balance := 1

		L:
			for {
				r, rsize := decodeRune(input, sp)

				sp += rsize

				switch r {
				case y:
					balance--
					if balance == 0 {
						break L
					}
				case x:
					balance++
				case eos:
					return false
				}
			}

			pc++
		case opFrontier:
			r, _ := lastDecodeRune(input, sp)
			if r == sos {
				r = 0x00
			}
			if p.sets[inst.x].match(r) {
				return false
			}

			r, _ = decodeRune(input, sp)
			if r == eos {
				r = 0x00
			}
			if !p.sets[inst.x].match(r) {
				return false
			}

			pc++
		default:
			r, rsize := decodeRune(input, sp)
			if r == eos {
				return false
			}

			if !simpleMatch(inst.op, r) {
				return false
			}

			pc++
			sp += rsize
		}
	}
}
