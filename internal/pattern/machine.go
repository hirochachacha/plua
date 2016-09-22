package pattern

import "strings"

type thread struct {
	pc   int
	caps []Range

	stack []int // calculaiton stack for cap

	balance int

	sleep int
}

type queue struct {
	sparse []int
	dense  []*thread
}

func newQueue(n int) *queue {
	return &queue{sparse: make([]int, n), dense: make([]*thread, 0, n)}
}

func (q *queue) has(pc int) bool {
	j := q.sparse[pc]
	return 0 < j && j < len(q.dense) && q.dense[j].pc == pc
}

func (q *queue) sleep(pc int, caps []Range, stack []int, balance, sleep int) {
	j := len(q.dense)
	q.dense = q.dense[:j+1]
	d := q.dense[j]
	if d == nil {
		d = &thread{
			pc:      pc,
			caps:    make([]Range, 1, cap(caps)),
			stack:   make([]int, cap(caps)),
			balance: balance,
			sleep:   sleep,
		}
		copy(d.caps, caps)
		copy(d.stack, stack)
		q.dense[j] = d
		q.sparse[pc] = -1
	} else {
		if len(d.caps) < len(caps) {
			if cap(d.caps) >= len(caps) {
				d.caps = d.caps[:len(caps)]
			} else {
				d.caps = make([]Range, len(caps), cap(caps))
			}
		} else {
			d.caps = d.caps[:len(caps)]
		}

		if len(d.stack) < len(stack) {
			if cap(d.stack) >= len(stack) {
				d.stack = d.stack[:len(stack)]
			} else {
				d.stack = make([]int, cap(caps))
			}
		} else {
			d.stack = d.stack[:len(stack)]
		}

		d.pc = pc
		copy(d.caps, caps)
		copy(d.stack, stack)
		d.balance = balance
		d.sleep = sleep
		q.sparse[pc] = -1
	}
}

func (q *queue) push(pc int, caps []Range, stack []int, balance, sleep int) {
	j := len(q.dense)
	q.dense = q.dense[:j+1]
	d := q.dense[j]
	if d == nil {
		d = &thread{
			pc:      pc,
			caps:    make([]Range, len(caps), cap(caps)),
			stack:   make([]int, cap(caps)),
			balance: balance,
			sleep:   sleep,
		}
		copy(d.caps, caps)
		copy(d.stack, stack)
		q.dense[j] = d
		q.sparse[pc] = j
	} else {
		if len(d.caps) < len(caps) {
			if cap(d.caps) >= len(caps) {
				d.caps = d.caps[:len(caps)]
			} else {
				d.caps = make([]Range, len(caps), cap(caps))
			}
		} else {
			d.caps = d.caps[:len(caps)]
		}

		if len(d.stack) < len(stack) {
			if cap(d.stack) >= len(stack) {
				d.stack = d.stack[:len(stack)]
			} else {
				d.stack = make([]int, cap(caps))
			}
		} else {
			d.stack = d.stack[:len(stack)]
		}

		d.pc = pc
		copy(d.caps, caps)
		copy(d.stack, stack)
		d.balance = balance
		d.sleep = sleep
		q.sparse[pc] = j
	}
}

func (q *queue) clear() {
	q.dense = q.dense[:0]
	// q.dense = make([]*thread, 0, len(q.sparse))
	// return &queue{sparse: make([]int, n), dense: make([]*thread, 0, n)}
}

type machine struct {
	typ     matchType
	literal string
	code    []inst
	preds   predicates
	sets    []*rangeTable
	ncaps   int

	current *queue
	next    *queue
}

func (m *machine) findAll(input string) (matches []*MatchRange) {
	switch m.typ {
	case matchBeginning:
		if found := m.matchFrom(input, 0); found != nil {
			return []*MatchRange{found}
		}
		return nil
	case matchBeginning | matchFullLiteral:
		if strings.HasPrefix(input, m.literal) {
			return []*MatchRange{
				&MatchRange{
					Item: Range{0, len(m.literal)},
				},
			}
		}
		return nil
	case matchBeginning | matchFullLiteral | matchEnd:
		if input == m.literal {
			return []*MatchRange{
				&MatchRange{
					Item: Range{0, len(m.literal)},
				},
			}
		}
		return nil
	case matchBeginning | matchPrefixLiteral:
		if strings.HasPrefix(input, m.literal) {
			if found := m.matchFrom(input, 0); found != nil {
				return []*MatchRange{found}
			}
		}
		return nil
	case matchBeginning | matchPrefixLiteral | matchEnd:
		if strings.HasPrefix(input, m.literal) {
			if found := m.matchFrom(input, 0); found != nil {
				if found.Item.End == len(input) {
					return []*MatchRange{found}
				}
			}
		}
		return nil
	case matchEnd:
		for ioff := 0; ioff < len(input); ioff++ {
			if found := m.matchFrom(input, ioff); found != nil {
				if found.Item.End == len(input) {
					return []*MatchRange{found}
				}
			}
		}
		return nil
	case matchEnd | matchFullLiteral:
		if strings.HasSuffix(input, m.literal) {
			return []*MatchRange{
				&MatchRange{
					Item: Range{
						Begin: len(input) - len(m.literal),
						End:   len(input),
					},
				},
			}
		}
		return nil
	case matchEnd | matchPrefixLiteral:
		var idx int
		var ioff int

		for ioff < len(input) {
			idx = strings.Index(input[ioff:], m.literal)
			if idx == -1 {
				return nil
			}

			ioff += idx

			if found := m.matchFrom(input, ioff); found != nil {
				if found.Item.End == len(input) {
					return []*MatchRange{found}
				}

				ioff = found.Item.End
			} else {
				ioff++
			}
		}
		return nil
	case matchFullLiteral:
		var idx int
		var ioff int

		for ioff < len(input) {
			idx = strings.Index(input[ioff:], m.literal)
			if idx == -1 {
				break
			}

			matches = append(matches, &MatchRange{
				Item: Range{
					Begin: idx + ioff,
					End:   idx + len(m.literal) + ioff,
				},
			})

			ioff += idx + len(m.literal) + 1
		}
		return matches
	case matchPrefixLiteral:
		var idx int
		var ioff int

		for ioff < len(input) {
			idx = strings.Index(input[ioff:], m.literal)
			if idx == -1 {
				break
			}

			ioff += idx

			if found := m.matchFrom(input, ioff); found != nil {
				matches = append(matches, found)

				ioff = found.Item.End
			} else {
				ioff++
			}
		}
		return matches
	default:
		var ioff int

		for ioff < len(input) {
			if found := m.matchFrom(input, ioff); found != nil {
				matches = append(matches, found)

				ioff = found.Item.End
			} else {
				ioff++
			}
		}
		return matches
	}
}

func (m *machine) find(input string) *MatchRange {
	switch m.typ {
	case matchBeginning:
		return m.matchFrom(input, 0)
	case matchBeginning | matchFullLiteral:
		if strings.HasPrefix(input, m.literal) {
			return &MatchRange{Item: Range{0, len(m.literal)}}
		}
		return nil
	case matchBeginning | matchFullLiteral | matchEnd:
		if input == m.literal {
			return &MatchRange{Item: Range{0, len(input)}}
		}
		return nil
	case matchBeginning | matchPrefixLiteral:
		if strings.HasPrefix(input, m.literal) {
			return m.matchFrom(input, 0)
		}
		return nil
	case matchBeginning | matchPrefixLiteral | matchEnd:
		if strings.HasPrefix(input, m.literal) {
			if found := m.matchFrom(input, 0); found != nil {
				if found.Item.End == len(input) {
					return found
				}
			}
		}
		return nil
	case matchEnd:
		for ioff := 0; ioff < len(input); ioff++ {
			if found := m.matchFrom(input, ioff); found != nil {
				if found.Item.End == len(input) {
					return found
				}
			}
		}
		return nil
	case matchEnd | matchFullLiteral:
		if strings.HasSuffix(input, m.literal) {
			return &MatchRange{Item: Range{
				Begin: len(input) - len(m.literal),
				End:   len(input),
			}}
		}
		return nil
	case matchEnd | matchPrefixLiteral:
		var idx int
		var ioff int

		for ioff < len(input) {
			idx = strings.Index(input[ioff:], m.literal)
			if idx == -1 {
				return nil
			}

			ioff += idx

			if found := m.matchFrom(input, ioff); found != nil {
				if found.Item.End == len(input) {
					return found
				}

				ioff = found.Item.End
			} else {
				ioff++
			}
		}
		return nil
	case matchFullLiteral:
		idx := strings.Index(input, m.literal)
		if idx != -1 {
			return &MatchRange{Item: Range{idx, idx + len(m.literal)}}
		}
		return nil
	case matchPrefixLiteral:
		var idx int
		var ioff int

		for ioff < len(input) {
			idx = strings.Index(input[ioff:], m.literal)
			if idx == -1 {
				return nil
			}

			ioff += idx

			if found := m.matchFrom(input, ioff); found != nil {
				return found
			}

			ioff++
		}
		return nil
	default:
		var ioff int

		for ioff < len(input) {
			if found := m.matchFrom(input, ioff); found != nil {
				return found
			}

			ioff++
		}
		return nil
	}
}

func (m *machine) matchFrom(input string, off int) (matched *MatchRange) {
	// m.current = newQueue(len(m.current.sparse))
	// m.next = newQueue(len(m.current.sparse))
	m.current.clear()

	caps := make([]Range, 1, m.ncaps)
	stack := make([]int, m.ncaps)

	caps[0].Begin = off

	m.addThread(m.current, 0, caps, stack, 0, off)

	var prev rune

	for {
		if len(m.current.dense) == 0 {
			break
		}

		r, rsize := decodeRune(input, off+len(m.literal))
	L:
		for _, t := range m.current.dense {
			if t.sleep == 1 {
				m.next.push(t.pc+1, t.caps, t.stack, t.balance, t.sleep-1)
				continue
			} else if t.sleep > 0 {
				m.next.sleep(t.pc, t.caps, t.stack, t.balance, t.sleep-1)
				continue
			}

			ins := m.code[t.pc]

			var add bool

			switch ins.op {
			case opMatch:
				matched = &MatchRange{
					Item: Range{
						Begin: t.caps[0].Begin,
						End:   off + len(m.literal),
					},
					Captures: make([]Range, len(t.caps)-1, cap(t.caps)-1),
				}
				copy(matched.Captures, t.caps[1:])
				break L
			case opAny:
				add = true
			case opBalanceAny:
				add = t.balance > 0 && ins.x != int(r) && ins.y != int(r)
			case opRange:
				add = m.sets[ins.x].is(r, m.preds)
			case opNotRange:
				add = !m.sets[ins.x].is(r, m.preds)
			case opFrontier:
				add = !m.sets[ins.x].is(prev, m.preds) && m.sets[ins.x].is(r, m.preds)
			case opCapture:
				begin, end := t.caps[ins.x].Begin, t.caps[ins.x].End
				if input[begin:end] == input[off+len(m.literal):off+len(m.literal)+end-begin] {
					m.next.sleep(t.pc, t.caps, t.stack, t.balance, end-begin-1)
				}
			default:
				add, _ = m.preds.is(ins.op, r)
			}

			if add {
				m.addThread(m.next, t.pc+1, t.caps, t.stack, t.balance, off+rsize)
			}
		}

		if r == eot {
			break
		}

		m.current, m.next = m.next, m.current

		m.next.clear()

		off += rsize

		prev = r
	}

	return matched
}

func (m *machine) addThread(q *queue, pc int, caps []Range, stack []int, balance, off int) {
	var ins inst
	for {
		// prevent infinite loop
		if q.has(pc) {
			return
		}

		ins = m.code[pc]

		switch ins.op {
		case opJmp:
			pc = ins.x
		case opSplit:
			pc = ins.y
			m.addThread(q, ins.x, caps, stack, balance, off)
		case opBalanceUp:
			balance++
			pc = ins.x
		case opBalanceDown:
			balance--
			if balance < 0 {
				return
			}

			if balance == 0 {
				pc = ins.y
			} else {
				pc = ins.x
			}
		case opEnterSave:
			pc++
			caps = append(caps, Range{off + len(m.literal), -1})
			stack[ins.x] = len(caps) - 1
		case opExitSave:
			pc++
			caps[stack[ins.x]].End = off + len(m.literal)
		default:
			q.push(pc, caps, stack, balance, 0)

			return
		}
	}
}
