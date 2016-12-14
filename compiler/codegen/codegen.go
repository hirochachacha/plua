package codegen

import (
	"fmt"
	"math"

	"github.com/hirochachacha/plua/compiler/ast"
	"github.com/hirochachacha/plua/compiler/token"
	"github.com/hirochachacha/plua/internal/strconv"
	"github.com/hirochachacha/plua/internal/version"
	"github.com/hirochachacha/plua/object"
	"github.com/hirochachacha/plua/opcode"
	"github.com/hirochachacha/plua/position"
)

const (
	skipPeepholeOptimization = false
	skipDeadCodeElimination  = false
	skipConstantFolding      = false
)

var tmp ast.Name

func tmpName(name string) *ast.Name {
	t := &tmp
	t.Name = name
	return t
}

func Generate(f *ast.File) (proto *object.Proto, err error) {
	g := newGenerator(nil)

	g.Source = f.Filename
	g.cfolds = make(map[ast.Expr]object.Value) // cache for constant folding

	defer func() {
		if r := recover(); r != nil {
			b := r.(bailout)

			err = b.err
		}
	}()

	g.genFile(f)

	proto = g.Proto

	return
}

type label struct {
	pc  int
	sp  int
	pos position.Position
}

type jumpPoint struct {
	pc  int
	pos position.Position
}

type generator struct {
	*object.Proto
	*scope            // block scope
	outer  *generator // function scope

	cfolds map[ast.Expr]object.Value // cache for constant folding

	sp int

	pendingJumps map[*scope]map[string]jumpPoint // pending jumps per scopes

	rconstants map[object.Value]int // reverse map of Proto.constants

	tokLine int

	locktmp  bool // don't remove tmp variable by peep hole optimization
	lockpeep bool // don't do peep hole optimization, because here is jump destination
}

type bailout struct {
	err error
}

func newGenerator(outer *generator) *generator {
	g := &generator{
		Proto:        new(object.Proto),
		outer:        outer,
		pendingJumps: make(map[*scope]map[string]jumpPoint, 0),
		rconstants:   make(map[object.Value]int, 0),
	}

	if outer != nil {
		g.Source = outer.Source
		g.cfolds = outer.cfolds
	}

	return g
}

func (g *generator) error(pos position.Position, err error) {
	pos.SourceName = g.Source

	panic(bailout{
		err: &Error{
			Pos: pos,
			Err: err,
		},
	})
}

func (g *generator) pc() int {
	return len(g.Code)
}

func (g *generator) nextSP() {
	g.addSP(1)
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

func (g *generator) newLabel() label {
	l := label{pc: g.pc(), sp: g.sp}

	g.lockpeep = true

	return l
}

func (g *generator) genJumpPoint() jumpPoint {
	return jumpPoint{pc: g.pushTemp()}
}

// backward jump
func (g *generator) genJumpTo(label label) {
	reljmp := label.pc - g.pc() - 1
	if reljmp > 0 {
		panic("unexpected")
	}

	g.pushInst(opcode.AsBx(opcode.JMP, label.sp+1, reljmp))
}

// forward jump
func (g *generator) genJumpFrom(jmp jumpPoint) {
	reljmp := g.pc() - jmp.pc - 1
	if reljmp < 0 {
		panic("unexpected")
	}

	g.Code[jmp.pc] = opcode.AsBx(opcode.JMP, 0, reljmp)

	g.lockpeep = true
}

// global jumps

func (g *generator) declareLabel(name string) {
	g.declareLabelPos(name, position.NoPos)
}

func (g *generator) declareLabelPos(name string, pos position.Position) {
	g.labels[name] = label{
		pc:  g.pc(),
		sp:  g.sp,
		pos: pos,
	}
}

func (g *generator) genSetJumpPoint(name string, pos position.Position) {
	g.tokLine = pos.Line

	jmp := g.genJumpPoint()

	jmp.pos = pos

	if jmps, ok := g.pendingJumps[g.scope]; ok {
		jmps[name] = jmp
	} else {
		g.pendingJumps[g.scope] = map[string]jumpPoint{
			name: jmp,
		}
	}
}

// close pending jumps
func (g *generator) closeJumps() {
	for scope, pendingJumps := range g.pendingJumps {
		if len(pendingJumps) == 0 {
			continue
		}

		for name, jmp := range pendingJumps {
			label, ok := scope.resolveLabel(name)
			if !ok {
				g.error(jmp.pos, fmt.Errorf("unknown label '%s' for jump", name))
			}

			for _, locVar := range g.LocVars {
				if jmp.pc < locVar.StartPC && locVar.StartPC <= label.pc && label.pc < locVar.EndPC {
					g.error(label.pos, fmt.Errorf("forward jump over local '%s'", locVar.Name))
				}
			}

			reljmp := label.pc - jmp.pc - 1
			if reljmp >= 0 { // forward jump
				g.Code[jmp.pc] = opcode.AsBx(opcode.JMP, 0, reljmp)
			} else { // backward jump
				g.Code[jmp.pc] = opcode.AsBx(opcode.JMP, label.sp+1, reljmp)
			}
		}

		delete(g.pendingJumps, scope)
	}
}

func (g *generator) resolveName(name *ast.Name) (link, bool) {
	if g == nil {
		return link{}, false
	}

	if l, ok := g.resolveLocal(name.Name); ok {
		return l, true
	}

	if up, ok := g.outer.resolveName(name); ok {
		return g.declareUpvalueName(name, up), true
	}

	return link{}, false
}

func (g *generator) resolve(name string) (link, bool) {
	return g.resolveName(tmpName(name))
}

func (g *generator) declareLocalName(name *ast.Name, sp int) {
	nlocals := g.scope.nlocals

	if outer := g.scope.outer; outer != nil {
		nlocals -= outer.nlocals
	}

	if nlocals >= version.MAXVAR {
		g.error(name.Pos(), fmt.Errorf("too many local variables (limit is %d)", version.MAXVAR))
	}

	locVar := object.LocVar{
		Name:    name.Name,
		StartPC: g.pc(),
	}

	g.LocVars = append(g.LocVars, locVar)

	g.scope.declare(name.Name, link{
		kind:  linkLocal,
		index: sp,
	})

	g.nlocals++
}

func (g *generator) declareLocal(name string, sp int) {
	g.declareLocalName(tmpName(name), sp)
}

func (g *generator) declareEnviron() {
	u := object.UpvalueDesc{
		Name:    "_ENV",
		Instack: true,
		Index:   0,
	}

	g.Upvalues = append(g.Upvalues, u)

	g.scope.declare("_ENV", link{
		kind:  linkUpval,
		index: 0,
	})
}

func (g *generator) declareUpvalueName(name *ast.Name, up link) link {
	if len(g.Upvalues) >= version.MAXUPVAL {
		g.error(name.Pos(), fmt.Errorf("too many upvalues (limit is %d)", version.MAXUPVAL))
	}

	instack := up.kind == linkLocal

	// mark upvalue whether it should be closed
	if instack {
		scope := g.outer.scope

		for {
			_, ok := scope.symbols[name.Name]
			if ok {
				break
			}

			scope = scope.outer
		}

		if scope.outer != nil {
			scope.doClose = true
		}
	}

	ud := object.UpvalueDesc{
		Name:    name.Name,
		Instack: instack,
		Index:   up.index,
	}

	g.Upvalues = append(g.Upvalues, ud)

	link := link{
		kind:  linkUpval,
		index: len(g.Upvalues) - 1,
	}

	g.scope.root().declare(name.Name, link)

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

func (g *generator) markRK(k int, next bool) (rk int) {
	if k > opcode.MaxRKindex {
		if k > opcode.MaxBx {
			g.pushInst(opcode.ABx(opcode.LOADKX, g.sp, 0))
			g.pushInst(opcode.Ax(opcode.EXTRAARG, k))
		} else {
			g.pushInst(opcode.ABx(opcode.LOADK, g.sp, k))
		}

		rk = g.sp

		if next {
			g.nextSP()
		}

		return rk
	}

	return k | opcode.BitRK
}

func (g *generator) newScope() {
	g.scope = &scope{
		symbols: make(map[string]link, 0),
		labels:  make(map[string]label, 0),
		outer:   nil,
		savedSP: 0,
		nlocals: 0,
	}
}

func (g *generator) openScope() {
	g.scope = &scope{
		symbols: make(map[string]link, 0),
		labels:  make(map[string]label, 0),
		outer:   g.scope,
		savedSP: g.sp,
		nlocals: g.scope.nlocals,
	}
}

func (g *generator) closeScope() {
	nlocals := g.scope.nlocals

	outer := g.scope.outer

	if outer != nil {
		nlocals -= outer.nlocals
	}

	endPC := g.pc()

	if nlocals != 0 {
		for i := len(g.LocVars) - 1; i >= 0; i-- {
			if g.LocVars[i].EndPC != 0 {
				continue
			}

			nlocals--

			g.LocVars[i].EndPC = endPC

			if nlocals == 0 {
				break
			}
		}
	}

	g.sp = g.scope.savedSP

	if g.scope.doClose {
		g.pushInst(opcode.AsBx(opcode.JMP, g.sp+1, 0))
	}

	g.scope.endPC = endPC
	g.scope = outer

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

func (g *generator) unquoteString(tok token.Token) string {
	us, err := strconv.Unquote(tok.Lit)
	if err != nil {
		g.error(tok.Pos, fmt.Errorf("failed to unquote %s, err: %v", tok.Lit, err))
	}
	return us
}

func (g *generator) parseInteger(tok token.Token, negate bool) (ret object.Integer, ok bool) {
	if negate {
		tok.Lit = "-" + tok.Lit
	}

	i, err := strconv.ParseInt(tok.Lit)
	if err != nil {
		if err != strconv.ErrRange {
			g.error(tok.Pos, fmt.Errorf("failed to parse int %s, err: %v", tok.Lit, err))
		}

		return 0, false
	}

	return object.Integer(i), true
}

func (g *generator) parseNumber(tok token.Token, negate bool) object.Number {
	if negate {
		tok.Lit = "-" + tok.Lit
	}

	f, err := strconv.ParseFloat(tok.Lit)
	if err != nil {
		if err != strconv.ErrRange {
			g.error(tok.Pos, fmt.Errorf("failed to parse float %s, err: %v", tok.Lit, err))
		}
	}

	return object.Number(f)
}
