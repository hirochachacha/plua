package codegen

import (
	"fmt"
	"math"

	"github.com/hirochachacha/plua/compiler/ast"
	"github.com/hirochachacha/plua/internal/strconv"
	"github.com/hirochachacha/plua/object"
	"github.com/hirochachacha/plua/opcode"
)

// assertion flag for debugging
const assert = false

const (
	skipPeepholeOptimization = false
	skipDeadCodeElimination  = false
	skipConstantFolding      = false
)

func Generate(f *ast.File) (proto *object.Proto) {
	g := newGenerator(nil)

	g.Source = f.Filename
	g.cfolds = make(map[ast.Expr]object.Value) // cache for constant folding

	g.genFile(f)

	proto = g.Proto

	g.Proto = nil

	return
}

// snapshot of proto generator
type label struct {
	pc int
	sp int
}

func newGenerator(outer *generator) *generator {
	g := &generator{
		Proto:      new(object.Proto),
		outer:      outer,
		jumps:      make(map[*scope]map[string]int, 0),
		rconstants: make(map[object.Value]int, 0),
	}

	if outer != nil {
		g.Source = outer.Source
		g.cfolds = outer.cfolds
	}

	return g
}

type generator struct {
	*object.Proto
	*scope            // block scope
	outer  *generator // function scope

	cfolds map[ast.Expr]object.Value // cache for constant folding

	sp int

	jumps map[*scope]map[string]int // pending jumps

	rconstants map[object.Value]int // reverse map of Proto.constants

	tokLine int

	locktmp  bool // don't remove tmp variable by peep hole optimization
	lockpeep bool // don't do peep hole optimization, because here is jump destination
}

func (g *generator) pc() int {
	return len(g.Code)
}

func (g *generator) nextSP() {
	g.sp++
	if g.sp > g.MaxStackSize {
		g.MaxStackSize = g.sp
	}
}

func (g *generator) addSP(i int) {
	g.sp += i
	if g.sp > g.MaxStackSize {
		g.MaxStackSize = g.sp
	}
}

func (g *generator) setSP(sp int) {
	g.sp = sp
	if g.sp > g.MaxStackSize {
		g.MaxStackSize = g.sp
	}
}

// local jumps

func (g *generator) newLocalLabel() (lid int) {
	lid = g.lid

	g.llabels[lid] = &label{pc: g.pc(), sp: g.sp}

	g.lid++

	g.lockpeep = true

	return
}

func (g *generator) genPendingLocalJump() (lid int) {
	lid = g.lid

	g.llabels[lid] = &label{pc: g.pushTemp(), sp: g.sp}

	g.lid++

	return
}

func (g *generator) setLocalJumpDst(lid int) {
	label := g.llabels[lid]

	reljmp := g.pc() - label.pc - 1
	if assert {
		if reljmp < 0 {
			panic("unexpected")
		}
	}

	g.Code[label.pc] = opcode.AsBx(opcode.JMP, 0, reljmp)

	g.lockpeep = true
}

func (g *generator) genLocalJump(lid int) {
	label := g.llabels[lid]

	reljmp := label.pc - g.pc() - 1
	if assert {
		if reljmp > 0 {
			panic("unexpected")
		}
	}

	g.pushInst(opcode.AsBx(opcode.JMP, label.sp+1, reljmp))
}

// global jumps

func (g *generator) newLabel(name string) {
	g.labels[name] = &label{
		pc: g.pc(),
		sp: g.sp,
	}
}

func (g *generator) genJump(name string) {
	if label := g.resolveLabel(name); label != nil {
		// forward jump
		// if label are already defined

		reljmp := label.pc - g.pc() - 1
		if assert {
			if reljmp > 0 {
				panic("unexpected")
			}
		}

		a := label.sp + 1

		g.pushInst(opcode.AsBx(opcode.JMP, a, reljmp))
	} else {
		// backward jump
		g.genPendingJump(name)
	}
}

func (g *generator) genPendingJump(label string) {
	jumps, ok := g.jumps[g.scope]
	if !ok {
		jumps = make(map[string]int, 0)
		g.jumps[g.scope] = jumps
	}

	jumps[label] = g.pushTemp()
}

// close pending jumps
func (g *generator) closeJumps() {
	for scope, jumps := range g.jumps {
		if len(jumps) == 0 {
			continue
		}

		for name, jmp := range jumps {
			label := scope.resolveLabel(name)
			if label == nil {
				panic("unresolve jump " + name)
			}
			reljmp := label.pc - jmp - 1
			if assert {
				if reljmp < 0 {
					panic("unexpected")
				}
			}

			g.Code[jmp] = opcode.AsBx(opcode.JMP, g.sp+1, reljmp)
		}

		delete(g.jumps, scope)
	}

	g.lockpeep = true
}

func (g *generator) resolve(name string) *link {
	if g == nil {
		return nil
	}

	l := g.resolveLocal(name)
	if l != nil {
		return l
	}

	l = g.outer.resolve(name)
	if l != nil {
		return g.declareUpvalue(name, l)
	}

	return nil
}

func (g *generator) declareLocal(name string, v int) {
	l := &link{
		kind: linkLocal,
		v:    v,
	}

	locVar := object.LocVar{
		Name:    name,
		StartPC: g.pc(),
	}

	g.LocVars = append(g.LocVars, locVar)

	g.scope.declare(name, l)

	if assert {
		if g.nlocals != v {
			panic("unexpected")
		}
	}

	g.nlocals++
}

func (g *generator) declareEnviron() {
	u := object.UpvalueDesc{
		Name:    "_ENV",
		Instack: true,
		Index:   0,
	}

	g.Upvalues = append(g.Upvalues, u)

	link := &link{
		kind: linkUpval,
		v:    0,
	}

	g.scope.declare("_ENV", link)
}

func (g *generator) declareUpvalue(name string, l *link) *link {
	instack := l.kind == linkLocal

	// mark upvalue should be close or not
	if instack {
		scope := g.outer.scope

		for {
			_, ok := scope.symbols[name]
			if ok {
				break
			}

			scope = scope.outer
		}

		if scope.outer != nil {
			scope.doClose = true
		}
	}

	u := object.UpvalueDesc{
		Name:    name,
		Instack: instack,
		Index:   l.v,
	}

	g.Upvalues = append(g.Upvalues, u)

	link := &link{
		kind: linkUpval,
		v:    len(g.Upvalues) - 1,
	}

	g.scope.root().declare(name, link)

	return link
}

type negativeZero struct{}

func (n negativeZero) Type() object.Type {
	return object.Type(-1)
}

func (n negativeZero) String() string {
	return "-0.0"
}

func (g *generator) constant(val object.Value) (k int) {
	key := val

	// a stupid trick for avoiding +0.0 == -0.0
	if n, ok := val.(object.Number); ok && n == 0 {
		u := math.Float64bits(float64(n))
		if int(u>>63) == 1 {
			key = negativeZero{}
		}
	}

	if k, ok := g.rconstants[key]; ok {
		return k
	}

	k = len(g.Constants)

	g.Constants = append(g.Constants, val)

	g.rconstants[key] = k

	return
}

func (g *generator) proto(f *ast.FuncBody, self bool, endLine int) (p int) {
	generator := newGenerator(g)

	generator.genFuncBody(f, self, endLine)

	g.Protos = append(g.Protos, generator.Proto)

	generator.Proto = nil
	generator.outer = nil

	return len(g.Protos) - 1
}

func (g *generator) markRK(k int) (rk int) {
	if k > opcode.MaxRKindex {
		if k > opcode.MaxBx {
			g.pushInst(opcode.ABx(opcode.LOADKX, g.sp, 0))
			g.pushInst(opcode.Ax(opcode.EXTRAARG, k))
		} else {
			g.pushInst(opcode.ABx(opcode.LOADK, g.sp, k))
		}

		rk = g.sp

		g.nextSP()

		return
	}

	return k | opcode.BitRK
}

func (g *generator) newScope() {
	g.scope = &scope{
		symbols: make(map[string]*link, 0),
		labels:  make(map[string]*label, 0),
		llabels: make(map[int]*label, 0),
	}
}

func (g *generator) openScope() {
	g.scope = &scope{
		symbols: make(map[string]*link, 0),
		labels:  make(map[string]*label, 0),
		llabels: make(map[int]*label, 0),
		outer:   g.scope,
		savedSP: g.sp,
		nlocals: g.scope.nlocals,
	}
}

func (g *generator) closeScope() {
	g.sp = g.scope.savedSP

	if g.scope.doClose {
		g.pushInst(opcode.AsBx(opcode.JMP, g.sp+1, 0))
	}

	nlocals := g.scope.nlocals

	g.scope = g.scope.outer

	if g.scope != nil {
		nlocals -= g.scope.nlocals
	}

	if nlocals != 0 {
		pc := g.pc()

		for i := len(g.LocVars) - 1; i >= 0; i-- {
			if g.LocVars[i].EndPC != 0 {
				continue
			}

			nlocals--

			g.LocVars[i].EndPC = pc

			if nlocals == 0 {
				break
			}
		}
	}

	return
}

func (g *generator) pushInst(inst opcode.Instruction) {
	g.pushInstLine(inst, g.tokLine)
}

func (g *generator) pushInstLine(inst opcode.Instruction, line int) {
	if !skipPeepholeOptimization && !g.lockpeep {
		g.peepLine(inst, line)
	} else {
		g.Code = append(g.Code, inst)
		g.LineInfo = append(g.LineInfo, line)
	}
	g.lockpeep = false
}

func (g *generator) pushTemp() (pc int) {
	return g.pushTempLine(g.tokLine)
}

func (g *generator) pushTempLine(line int) (pc int) {
	pc = g.pc()

	g.Code = append(g.Code, opcode.AsBx(opcode.JMP, 0, 0))
	g.LineInfo = append(g.LineInfo, line)

	return
}

func (g *generator) pushReturn() {
	g.Code = append(g.Code, opcode.AB(opcode.RETURN, 0, 1))
	g.LineInfo = append(g.LineInfo, g.LastLineDefined)
}

func unquoteString(s string) string {
	us, err := strconv.Unquote(s)
	if err != nil {
		panic(fmt.Sprintf("failed to unquote %s, err: %v", s, err))
	}
	return us
}

func parseInteger(g string) (ret object.Integer, inf int) {
	i, err := strconv.ParseInt(g)
	if err != nil {
		if err != strconv.ErrRange {
			panic(fmt.Sprintf("strconv.ParseInt(%s) = %v", g, err))
		}

		// infinity
		if i < 0 {
			return 0, -1
		}

		return 0, 1
	}

	return object.Integer(i), 0
}

func parseNumber(g string) object.Number {
	f, err := strconv.ParseFloat(g)
	if err != nil {
		if err != strconv.ErrRange {
			panic(fmt.Sprintf("strconv.ParseFloat(%s) = %v", g, err))
		}
	}

	return object.Number(f)
}
