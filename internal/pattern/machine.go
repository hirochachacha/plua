package pattern

type thread struct {
	pc  int
	cap [][2]int

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

func (q *queue) sleep(pc int, _cap [][2]int, stack []int, balance, sleep int) {
	j := len(q.dense)
	q.dense = q.dense[:j+1]
	d := q.dense[j]
	if d == nil {
		d = &thread{
			pc:      pc,
			cap:     make([][2]int, 1, cap(_cap)),
			stack:   make([]int, cap(_cap)),
			balance: balance,
			sleep:   sleep,
		}
		for i := range _cap {
			d.cap[i] = _cap[i]
		}
		copy(d.stack, stack)
		q.dense[j] = d
		q.sparse[pc] = -1
	} else {
		d.pc = pc
		for i := range _cap {
			d.cap[i] = _cap[i]
		}
		copy(d.stack, stack)
		d.balance = balance
		d.sleep = sleep
		q.sparse[pc] = -1
	}
}

func (q *queue) push(pc int, _cap [][2]int, stack []int, balance, sleep int) {
	j := len(q.dense)
	q.dense = q.dense[:j+1]
	d := q.dense[j]
	if d == nil {
		d = &thread{
			pc:      pc,
			cap:     make([][2]int, len(_cap), cap(_cap)),
			stack:   make([]int, cap(_cap)),
			balance: balance,
			sleep:   sleep,
		}
		for i := range _cap {
			d.cap[i] = _cap[i]
		}
		copy(d.stack, stack)
		q.dense[j] = d
		q.sparse[pc] = j
	} else {
		if len(d.cap) < len(_cap) {
			if cap(d.cap) >= len(_cap) {
				d.cap = d.cap[:cap(d.cap)]
			} else {
				d.cap = make([][2]int, len(_cap), cap(_cap))
			}
		}
		if len(d.stack) < len(stack) {
			if cap(d.stack) >= len(stack) {
				d.stack = d.stack[:cap(d.stack)]
			} else {
				d.stack = make([]int, cap(_cap))
			}
		}
		d.pc = pc
		copy(d.cap, _cap)
		copy(d.stack, stack)
		d.balance = balance
		d.sleep = sleep
		q.sparse[pc] = j
	}
}

func (q *queue) clear() {
	q.dense = q.dense[:0]
}

type machine struct {
	typ    matchType
	prefix string
	insts  []inst

	preds predicates

	sets []*rangeTable

	ncaps int

	current *queue
	next    *queue
}

func (m *machine) findAll(input input) (rets [][][2]int) {
	switch m.typ {
	case matchBeginning:
		if found := m.matchFrom(input, 0); found != nil {
			return [][][2]int{found}
		}

		return nil
	case matchBeginning | matchFullLiteral:
		if input.hasPrefix(m) {
			return [][][2]int{{{0, len(m.prefix)}}}
		}
		return nil
	case matchBeginning | matchFullLiteral | matchEnd:
		if input.isPrefix(m) {
			return [][][2]int{{{0, len(m.prefix)}}}
		}
		return nil
	case matchBeginning | matchPrefixLiteral:
		if input.hasPrefix(m) {
			if found := m.matchFrom(input, 0); found != nil {
				return [][][2]int{found}
			}
		}
		return nil
	case matchBeginning | matchPrefixLiteral | matchEnd:
		if input.hasPrefix(m) {
			found := m.matchFrom(input, 0)
			if found[0][1] == input.length() {
				return [][][2]int{found}
			}
		}
		return nil
	case matchEnd:
		inputLen := input.length()
		for ioffset := 0; ioffset < inputLen; ioffset++ {
			found := m.matchFrom(input, ioffset)
			if found != nil && found[0][1] == inputLen {
				return [][][2]int{found}
			}
		}
		return nil
	case matchEnd | matchFullLiteral:
		if input.hasSuffix(m) {
			return [][][2]int{{{input.length() - len(m.prefix), input.length()}}}
		}
		return nil
	case matchEnd | matchPrefixLiteral:
		var idx int
		var ioffset int
		inputLen := input.length()

		for {
			idx = input.index(m, ioffset)
			if idx == -1 {
				return nil
			}

			ioffset += idx

			found := m.matchFrom(input, ioffset)
			if found != nil {
				if found[0][1] == inputLen {
					return [][][2]int{found}
				}

				ioffset = found[0][1]
			} else {
				ioffset++
			}

			if ioffset >= inputLen {
				return
			}
		}
	case matchFullLiteral:
		var idx int
		var ioffset int

		for {
			idx = input.index(m, ioffset)
			if idx == -1 {
				return rets
			}

			rets = append(rets, [][2]int{{idx + ioffset, idx + len(m.prefix) + ioffset}})

			ioffset += idx + len(m.prefix) + 1
		}
	case matchPrefixLiteral:
		var idx int
		var ioffset int

		inputLen := input.length()

		for {
			idx = input.index(m, ioffset)
			if idx == -1 {
				return rets
			}

			ioffset += idx

			found := m.matchFrom(input, ioffset)
			if found != nil {
				rets = append(rets, found)

				ioffset = found[0][1]
			} else {
				ioffset++
			}

			if ioffset >= inputLen {
				return
			}
		}
	}

	var ioffset int

	inputLen := input.length()

	for {
		found := m.matchFrom(input, ioffset)
		if found != nil {
			rets = append(rets, found)

			ioffset = found[0][1]
		} else {
			ioffset++
		}

		if ioffset >= inputLen {
			return
		}
	}

	return
}

func (m *machine) find(input input) [][2]int {
	switch m.typ {
	case matchBeginning:
		return m.matchFrom(input, 0)
	case matchBeginning | matchFullLiteral:
		if input.hasPrefix(m) {
			return [][2]int{{0, len(m.prefix)}}
		}
		return nil
	case matchBeginning | matchFullLiteral | matchEnd:
		if input.isPrefix(m) {
			return [][2]int{{0, input.length()}}
		}
		return nil
	case matchBeginning | matchPrefixLiteral:
		if input.hasPrefix(m) {
			return m.matchFrom(input, 0)
		}
		return nil
	case matchBeginning | matchPrefixLiteral | matchEnd:
		if input.hasPrefix(m) {
			found := m.matchFrom(input, 0)
			if found[0][1] == input.length() {
				return found
			}
		}
		return nil
	case matchEnd:
		inputLen := input.length()
		for ioffset := 0; ioffset < inputLen; ioffset++ {
			found := m.matchFrom(input, ioffset)
			if found != nil && found[0][1] == inputLen {
				return found
			}
		}
		return nil
	case matchEnd | matchFullLiteral:
		if input.hasSuffix(m) {
			return [][2]int{{input.length() - len(m.prefix), input.length()}}
		}
		return nil
	case matchEnd | matchPrefixLiteral:
		var idx int
		var ioffset int
		inputLen := input.length()

		for {
			idx = input.index(m, ioffset)
			if idx == -1 {
				return nil
			}

			ioffset += idx

			found := m.matchFrom(input, ioffset)
			if found != nil {
				if found[0][1] == inputLen {
					return found
				}

				ioffset = found[0][1]
			} else {
				ioffset++
			}

			if ioffset >= inputLen {
				return nil
			}
		}
	case matchFullLiteral:
		idx := input.index(m, 0)
		if idx != -1 {
			return [][2]int{{idx, idx + len(m.prefix)}}
		}
		return nil
	case matchPrefixLiteral:
		var idx int
		var ioffset int

		inputLen := input.length()

		for {
			idx = input.index(m, ioffset)
			if idx == -1 {
				return nil
			}

			ioffset += idx

			found := m.matchFrom(input, ioffset)
			if found != nil {
				return found
			}

			ioffset++

			if ioffset >= inputLen {
				return nil
			}
		}
	}

	var ioffset int

	inputLen := input.length()

	for {
		found := m.matchFrom(input, ioffset)
		if found != nil {
			return found
		}

		ioffset++

		if ioffset >= inputLen {
			return nil
		}
	}

	return nil
}

func (m *machine) matchFrom(input input, pos int) [][2]int {
	var matched [][2]int

	ioffset := pos + len(m.prefix)

	var add bool
	var r rune
	var prev rune
	var i int
	var ins inst

	_cap := make([][2]int, 1, m.ncaps)
	_cap[0][0] = pos
	stack := make([]int, m.ncaps)

	m.current.clear()

	m.addThread(m.current, 0, _cap, stack, 0, ioffset)
	for {
		r, i = input.stepRune(ioffset)

		if len(m.current.dense) == 0 {
			break
		}

	L:
		for _, t := range m.current.dense {
			if t.sleep == 1 {
				m.next.push(t.pc+1, t.cap, t.stack, t.balance, t.sleep-1)
				continue
			} else if t.sleep > 0 {
				m.next.sleep(t.pc, t.cap, t.stack, t.balance, t.sleep-1)
				continue
			}

			ins = m.insts[t.pc]

			switch ins.op {
			case opMatch:
				matched = make([][2]int, len(t.cap), cap(t.cap))
				copy(matched, t.cap)
				matched[0][1] = ioffset
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
				begin, end := t.cap[ins.x][0], t.cap[ins.x][1]
				if input.submatch(begin, end, ioffset) {
					m.next.sleep(t.pc, t.cap, t.stack, t.balance, end-begin-1)
				}
			default:
				add, _ = m.preds.is(ins.op, r)
			}

			if add {
				m.addThread(m.next, t.pc+1, t.cap, t.stack, t.balance, ioffset+i)
			}
		}

		if r == endOfText {
			break
		}

		m.current, m.next = m.next, m.current

		m.next.clear()

		ioffset += i

		prev = r
	}

	return matched
}

func (m *machine) addThread(q *queue, pc int, cap [][2]int, stack []int, balance, pos int) {
	var ins inst
	for {
		// prevent infinite loop
		if q.has(pc) {
			return
		}

		ins = m.insts[pc]

		switch ins.op {
		case opJmp:
			pc = ins.x
		case opSplit:
			pc = ins.y
			m.addThread(q, ins.x, cap, stack, balance, pos)
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
			cap = append(cap, [2]int{pos, 0})
			stack[ins.x] = len(cap) - 1
		case opExitSave:
			pc++
			cap[stack[ins.x]][1] = pos
		default:
			q.push(pc, cap, stack, balance, 0)

			return
		}
	}
}
