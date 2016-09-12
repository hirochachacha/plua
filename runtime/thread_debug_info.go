package runtime

import (
	"strings"

	"github.com/hirochachacha/plua/internal/version"
	"github.com/hirochachacha/plua/object"
	"github.com/hirochachacha/plua/opcode"
)

func (th *thread) getInfo(level int, what string) *object.DebugInfo {
	if level < 0 {
		return nil
	}

	ctx := th.context

	ci := ctx.ci

	for i := level; i > 0; i-- {
		ci = ci.prev

		if ci == nil {
			return nil
		}
	}

	d := &object.DebugInfo{Func: ctx.fn(), CallInfo: ci}

	cl := ci.closure

	for _, r := range what {
		switch r {
		case 'S':
			setFuncInfo(d, cl)
		case 'l':
			d.CurrentLine = getCurrentLine(ci)
		case 'u':
			setUpInfo(d, cl)
		case 't':
			d.IsTailCall = ci.isTailCall
		case 'n':
			if ctx.hookMask != 0 {
				d.Name = "?"
				d.NameWhat = "hook"
			} else {
				if !ci.isTailCall {
					prev := ci.prev
					if prev != nil && !prev.isGoFunction() {
						setFuncName(d, prev)
					}
				}
			}
		case 'L':
			lines := th.NewTableSize(0, len(cl.LineInfo))
			for _, line := range cl.LineInfo {
				lines.Set(object.Integer(line), object.True)
			}
			d.Lines = lines
		}
	}

	return d
}

func (th *thread) getInfoByFunc(fn object.Value, what string) *object.DebugInfo {
	var cl *closure

	switch fn := fn.(type) {
	case object.Closure:
		cl = fn.(*closure)
	case object.GoFunction:
	default:
		panic("should be function")
	}

	d := &object.DebugInfo{Func: fn}

	for _, r := range what {
		switch r {
		case 'S':
			setFuncInfo(d, cl)
		case 'l':
			d.CurrentLine = -1
		case 'u':
			setUpInfo(d, cl)
		case 'L':
			if cl != nil {
				lines := th.NewTableSize(0, len(cl.LineInfo))
				for _, line := range cl.LineInfo {
					lines.Set(object.Integer(line), object.True)
				}
				d.Lines = lines
			}
		}
	}

	return d
}

func (th *thread) setLocal(d *object.DebugInfo, n int, val object.Value) (name string) {
	name, _ = th.getLocal(d, n)
	if name == "" {
		return
	}

	ctx := th.context

	if _, ok := d.Func.(object.Closure); ok {
		if d.CallInfo != nil {
			ci := d.CallInfo.(*callInfo)

			ctx.stack[ci.base+n] = val
		}
	}

	return
}

func (th *thread) getLocal(d *object.DebugInfo, n int) (name string, val object.Value) {
	if d == nil {
		return
	}

	ctx := th.context

	if fn, ok := d.Func.(object.Closure); ok {
		var pc int
		if d.CallInfo != nil {
			ci := d.CallInfo.(*callInfo)

			if n < 0 {
				return findvarargs(ci, -n)
			}

			pc = ci.pc

			val = ctx.stack[ci.base+n]
		} else if n < 0 {
			return
		}

		name = getLocalName(fn.Prototype(), pc, n)
	}

	return
}

func findvarargs(ci *callInfo, n int) (name string, val object.Value) {
	if n <= len(ci.varargs) {
		name, val = "(*vararg)", ci.varargs[n-1]
	}

	return
}

func getLocalName(p *object.Proto, pc, n int) (name string) {
	for _, locvar := range p.LocVars {
		if pc < locvar.StartPC {
			continue
		}

		if pc >= locvar.EndPC {
			break
		}

		if n == 0 {
			return locvar.Name
		}

		n--
	}

	return ""
}

func getUpvalName(p *object.Proto, n int) (name string) {
	name = p.Upvalues[n].Name
	if len(name) == 0 {
		name = "?"
	}

	return
}

func getRKName(p *object.Proto, pc, rk int) (name string) {
	if rk&opcode.BitRK != 0 {
		if s, ok := p.Constants[rk & ^opcode.BitRK].(object.String); ok {
			return string(s)
		}
	} else {
		name, nameWhat := getObjectName(p, pc, rk)
		if nameWhat == "constant" {
			return name
		}
	}

	return "?"
}

func setFuncInfo(d *object.DebugInfo, cl *closure) {
	if cl == nil {
		d.Source = "=[Go]"
		d.ShortSource = "[Go]"
		d.LineDefined = -1
		d.LastLineDefined = -1
		d.What = "Go"
	} else {
		if len(cl.Source) == 0 {
			d.Source = "=?"
			d.ShortSource = "?"
		} else {
			d.Source = cl.Source
			d.ShortSource = shorten(cl.Source)
		}
		d.LineDefined = cl.LineDefined
		d.LastLineDefined = cl.LastLineDefined
		if d.LineDefined == 0 {
			d.What = "main"
		} else {
			d.What = "Lua"
		}
	}
}

func getCurrentLine(ci *callInfo) int {
	if ci == nil || ci.isGoFunction() {
		return -1
	}
	if len(ci.LineInfo) == 0 {
		return -1
	}
	if ci.pc == 0 { // see execute0, go stack overflow
		return ci.LineInfo[0]
	}
	return ci.LineInfo[ci.pc-1]
}

func getObjectName(p *object.Proto, pc, n int) (name, nameWhat string) {
	name = getLocalName(p, pc, n)
	if len(name) != 0 {
		nameWhat = "local"

		return
	}

	pc = getRelativePC(p, pc, n)

	if pc != -1 { /* could find instruction? */
		inst := p.Code[pc]

		switch inst.OpCode() {
		case opcode.MOVE:
			b := inst.B()
			if b < inst.A() {
				return getObjectName(p, pc, b)
			}
		case opcode.GETTABUP:
			t := inst.B()
			key := inst.C()
			tn := getUpvalName(p, t)
			name = getRKName(p, pc, key)
			if tn == version.LUA_ENV {
				nameWhat = "global"
			} else {
				nameWhat = "field"
			}
		case opcode.GETTABLE:
			t := inst.B()
			key := inst.C()
			tn := getLocalName(p, pc, t)
			name = getRKName(p, pc, key)
			if tn == version.LUA_ENV {
				nameWhat = "global"
			} else {
				nameWhat = "field"
			}
		case opcode.GETUPVAL:
			name = getUpvalName(p, inst.B())
			nameWhat = "upvalue"
		case opcode.LOADK:
			bx := inst.Bx()
			if s, ok := p.Constants[bx].(object.String); ok {
				name = string(s)
			}
			nameWhat = "constant"
		case opcode.LOADKX:
			ax := p.Code[pc+1].Ax()
			if s, ok := p.Constants[ax].(object.String); ok {
				name = string(s)
			}
			nameWhat = "constant"
		case opcode.SELF:
			key := inst.C()
			name = getRKName(p, pc, key)
			nameWhat = "method"
		}
	}

	return
}

func getRelativePC(p *object.Proto, lastpc, n int) (relpc int) {
	var jmpdest int

	relpc = -1

	for pc := 0; pc < lastpc; pc++ {
		inst := p.Code[pc]

		a := inst.A()

		switch op := inst.OpCode(); op {
		case opcode.LOADNIL:
			b := inst.B()
			if a <= n && n <= a+b {
				if pc < jmpdest {
					relpc = -1
				} else {
					relpc = pc
				}
			}
		case opcode.TFORCALL:
			if n >= a+2 {
				if pc < jmpdest {
					relpc = -1
				} else {
					relpc = pc
				}
			}
		case opcode.CALL, opcode.TAILCALL:
			if n >= a {
				if pc < jmpdest {
					relpc = -1
				} else {
					relpc = pc
				}
			}
		case opcode.JMP:
			sbx := inst.SBx()
			dest := pc + 1 + sbx
			if pc < dest && dest <= lastpc {
				if dest > jmpdest {
					jmpdest = dest
				}
			}
		default:
			if op.TestAMode() && n == a {
				if pc < jmpdest {
					relpc = -1
				} else {
					relpc = pc
				}
			}
		}
	}

	return
}

func setUpInfo(d *object.DebugInfo, cl *closure) {
	if cl == nil {
		d.NUpvalues = 0
		d.IsVararg = true
		d.NParams = 0
	} else {
		d.NUpvalues = cl.NUpvalues()
		d.IsVararg = cl.IsVararg
		d.NParams = cl.NParams
	}
}

func setFuncName(d *object.DebugInfo, ci *callInfo) {
	var tag tagType

	inst := ci.Code[ci.pc-1]

	switch inst.OpCode() {
	case opcode.CALL, opcode.TAILCALL:
		d.Name, d.NameWhat = getObjectName(ci.Prototype(), ci.pc-1, inst.A())

		return
	case opcode.TFORCALL:
		d.Name = "for iterator"
		d.NameWhat = "for iterator"

		return
	case opcode.SELF, opcode.GETTABUP, opcode.GETTABLE:
		tag = TM_INDEX
	case opcode.SETTABUP, opcode.SETTABLE:
		tag = TM_NEWINDEX
	case opcode.ADD:
		tag = TM_ADD
	case opcode.SUB:
		tag = TM_SUB
	case opcode.MUL:
		tag = TM_MUL
	case opcode.MOD:
		tag = TM_MOD
	case opcode.POW:
		tag = TM_POW
	case opcode.DIV:
		tag = TM_DIV
	case opcode.IDIV:
		tag = TM_IDIV
	case opcode.BAND:
		tag = TM_BAND
	case opcode.BOR:
		tag = TM_BOR
	case opcode.BXOR:
		tag = TM_BXOR
	case opcode.SHL:
		tag = TM_SHL
	case opcode.SHR:
		tag = TM_SHR
	case opcode.UNM:
		tag = TM_UNM
	case opcode.BNOT:
		tag = TM_BNOT
	case opcode.LEN:
		tag = TM_LEN
	case opcode.CONCAT:
		tag = TM_CONCAT
	case opcode.EQ:
		tag = TM_EQ
	case opcode.LT:
		tag = TM_LT
	case opcode.LE:
		tag = TM_LE
	default:
		return
	}

	d.Name = tag.String()
	d.NameWhat = "metamethod"
}

func shorten(s string) string {
	if len(s) == 0 {
		return ""
	}

	switch s[0] {
	case '=':
		s = s[1:]
		if len(s) > version.LUA_IDSIZE {
			return s[:version.LUA_IDSIZE]
		}
		return s
	case '@':
		s = s[1:]
		if len(s) > version.LUA_IDSIZE {
			return "..." + s[len(s)-version.LUA_IDSIZE+3:len(s)]
		}
		return s
	default:
		i := strings.IndexRune(s, '\n')
		if i == -1 {
			s = "[string \"" + s

			if len(s) > version.LUA_IDSIZE-2 {
				return s[:version.LUA_IDSIZE-5] + "...\"]"
			}
			return s + "\"]"
		}

		s = "[string \"" + s[:i]

		if len(s) > version.LUA_IDSIZE-2 {
			return s[:version.LUA_IDSIZE-5] + "...\"]"
		}

		return s + "...\"]"
	}
}
