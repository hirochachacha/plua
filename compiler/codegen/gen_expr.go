package codegen

import (
	"github.com/hirochachacha/plua/compiler/ast"
	"github.com/hirochachacha/plua/compiler/token"
	"github.com/hirochachacha/plua/internal/version"
	"github.com/hirochachacha/plua/object"
	"github.com/hirochachacha/plua/opcode"
	"github.com/hirochachacha/plua/position"
)

type genType uint

const (
	genR    genType = 0         // normal register number
	genK    genType = 1 << iota // constant number
	genMove                     // copied register number
	genKey                      // if expr is *ast.Name, then return constant number of it'g name
)

// load or resolve
func (g *generator) genExpr(expr ast.Expr, typ genType) (rk int) {
	switch expr := expr.(type) {
	case *ast.BadExpr:
		panic("bad expr")
	case *ast.Name:
		rk = g.genName(expr, typ)
	case *ast.Vararg:
		rk = g.genVarargN(expr, 1)
	case *ast.BasicLit:
		rk = g.genBasicLit(expr, typ)
	case *ast.FuncLit:
		rk = g.genFuncLit(expr)
	case *ast.TableLit:
		rk = g.genTableLit(expr)
	case *ast.ParenExpr:
		rk = g.genExpr(expr.X, typ)
	case *ast.SelectorExpr:
		rk = g.genSelectorExpr(expr)
	case *ast.IndexExpr:
		rk = g.genIndexExpr(expr)
	case *ast.CallExpr:
		rk = g.genCallExprN(expr, 1, false)
	case *ast.UnaryExpr:
		// const folding
		if val, ok := g.foldUnary(expr); ok {
			rk = g.genConst(val, typ)
		} else {
			rk = g.genUnaryExpr(expr, typ)
		}
	case *ast.BinaryExpr:
		// const folding
		if val, ok := g.foldBinary(expr); ok {
			rk = g.genConst(val, typ)
		} else {
			rk = g.genBinaryExpr(expr, typ)
		}
	case *ast.KeyValueExpr:
		panic("unexpected")
	default:
		panic("unreachable")
	}
	return
}

func (g *generator) genExprN(expr ast.Expr, nrets int) (isvar bool) {
	locktmp := g.locktmp

	g.locktmp = true

	switch expr := expr.(type) {
	case *ast.BadExpr:
		panic("bad expr")
	case *ast.Name:
		g.genName(expr, genMove)
	case *ast.Vararg:
		g.genVarargN(expr, nrets)

		isvar = true
	case *ast.BasicLit:
		g.genBasicLit(expr, genMove)
	case *ast.FuncLit:
		g.genFuncLit(expr)
	case *ast.TableLit:
		g.genTableLit(expr)
	case *ast.ParenExpr:
		g.genExpr(expr.X, genMove)
	case *ast.SelectorExpr:
		g.genSelectorExpr(expr)
	case *ast.IndexExpr:
		g.genIndexExpr(expr)
	case *ast.CallExpr:
		g.genCallExprN(expr, nrets, false)

		isvar = true
	case *ast.UnaryExpr:
		// const folding
		if val, ok := g.foldUnary(expr); ok {
			g.genConst(val, genMove)
		} else {
			g.genUnaryExpr(expr, genMove)
		}
	case *ast.BinaryExpr:
		// const folding
		if val, ok := g.foldBinary(expr); ok {
			g.genConst(val, genMove)
		} else {
			g.genBinaryExpr(expr, genMove)
		}
	case *ast.KeyValueExpr:
		panic("unexpected")
	default:
		panic("unreachable")
	}

	if !isvar {
		if nrets > 1 {
			g.pushInst(opcode.AB(opcode.LOADNIL, g.sp, nrets-2))

			g.addSP(nrets - 1)
		}

		if nrets == 0 {
			g.addSP(-1)
		}
	}

	g.locktmp = locktmp

	return
}

// resolve or move
func (g *generator) genName(expr *ast.Name, typ genType) (rk int) {
	g.tokLine = expr.Pos().Line

	if typ&genKey != 0 {
		return g.markRK(g.constant(object.String(expr.Name)), true)
	}

	l, ok := g.resolveName(expr)
	if !ok {
		return g.genGetGlobal(expr)
	}

	switch l.kind {
	case linkUpval:
		g.pushInst(opcode.AB(opcode.GETUPVAL, g.sp, l.index))

		rk = g.sp

		g.nextSP()
	case linkLocal:
		if typ&genMove == 0 {
			return l.index
		}

		rk = g.sp

		g.pushInst(opcode.AB(opcode.MOVE, g.sp, l.index))

		g.nextSP()
	default:
		panic("unreachable")
	}

	return
}

func (g *generator) genVarargN(expr *ast.Vararg, nrets int) (r int) {
	g.tokLine = expr.Pos().Line

	sp := g.sp

	g.pushInst(opcode.AB(opcode.VARARG, g.sp, nrets+1))

	r = g.sp

	if nrets < 0 {
		// just update maxstacksize
		if g.sp+1 > g.MaxStackSize {
			g.MaxStackSize = g.sp + 1
		}

		return
	}

	g.setSP(sp + nrets)

	return
}

func (g *generator) genBasicLit(expr *ast.BasicLit, typ genType) (rk int) {
	g.tokLine = expr.Pos().Line

	var val object.Value

	tok := expr.Token

	switch tok.Type {
	case token.NIL:
		g.pushInst(opcode.AB(opcode.LOADNIL, g.sp, 0))

		rk = g.sp

		g.nextSP()

		return
	case token.FALSE:
		g.pushInst(opcode.ABC(opcode.LOADBOOL, g.sp, 0, 0))

		rk = g.sp

		g.nextSP()

		return
	case token.TRUE:
		g.pushInst(opcode.ABC(opcode.LOADBOOL, g.sp, 1, 0))

		rk = g.sp

		g.nextSP()

		return
	case token.INT:
		if i, ok := g.parseInteger(tok, false); ok {
			val = i
		} else {
			val = g.parseNumber(tok, false)
		}
	case token.FLOAT:
		val = g.parseNumber(tok, false)
	case token.STRING:
		val = object.String(g.unquoteString(tok))
	default:
		panic("unreachable")
	}

	return g.genConst(val, typ)
}

func (g *generator) genFuncLit(expr *ast.FuncLit) (r int) {
	g.tokLine = expr.Pos().Line

	body := expr.Body

	endLine := expr.End().Line

	p := g.proto(body, false, endLine)

	g.pushInstLine(opcode.ABx(opcode.CLOSURE, g.sp, p), endLine)

	r = g.sp

	g.nextSP()

	return
}

func (g *generator) genTableLit(expr *ast.TableLit) (r int) {
	g.tokLine = expr.Pos().Line

	var a []ast.Expr
	var m []*ast.KeyValueExpr
	for _, e := range expr.Fields {
		if keyval, ok := e.(*ast.KeyValueExpr); ok {
			m = append(m, keyval)
		} else {
			a = append(a, e)
		}
	}

	var skipped bool
	for len(a) > 0 {
		if lit, ok := a[len(a)-1].(*ast.BasicLit); ok && lit.Token.Type == token.NIL {
			a = a[:len(a)-1]

			skipped = true

			continue
		}

		break
	}

	g.pushInst(opcode.ABC(opcode.NEWTABLE, g.sp, opcode.IntToLog(len(a)), opcode.IntToLog(len(m))))

	tp := g.sp

	g.nextSP()

	for _, e := range m {
		sp := g.sp
		if e.Lbrack != position.NoPos {
			x := g.genExpr(e.Key, genR|genK)
			y := g.genExpr(e.Value, genR|genK)

			g.pushInst(opcode.ABC(opcode.SETTABLE, tp, x, y))
		} else {
			x := g.genExpr(e.Key, genR|genK|genKey)
			y := g.genExpr(e.Value, genR|genK)

			g.pushInst(opcode.ABC(opcode.SETTABLE, tp, x, y))
		}
		g.sp = sp
	}

	if len(a) > 0 {
		locktmp := g.locktmp

		g.locktmp = true

		sp := g.sp

		i := 1
		for {
			if len(a) < version.LUA_FPF {
				if len(a) == 0 {
					break
				}

				var isVar bool

				if skipped {
					for _, e := range a {
						g.genExpr(e, genMove)
					}
				} else {
					for _, e := range a[:len(a)-1] {
						g.genExpr(e, genMove)
					}

					isVar = g.genExprN(a[len(a)-1], -1)
				}

				n := len(a)
				if isVar {
					n = 0
				}

				if i > opcode.MaxC {
					g.pushInst(opcode.ABC(opcode.SETLIST, tp, n, 0))
					g.pushInst(opcode.Ax(opcode.EXTRAARG, i))
				} else {
					g.pushInst(opcode.ABC(opcode.SETLIST, tp, n, i))
				}

				// recover sp
				g.sp = sp

				break
			}

			for _, e := range a[:version.LUA_FPF] {
				g.genExpr(e, genMove)
			}

			if i > opcode.MaxC {
				g.pushInst(opcode.ABC(opcode.SETLIST, tp, version.LUA_FPF, 0))
				g.pushInst(opcode.Ax(opcode.EXTRAARG, i))
			} else {
				g.pushInst(opcode.ABC(opcode.SETLIST, tp, version.LUA_FPF, i))
			}

			// recover sp
			g.sp = sp

			a = a[version.LUA_FPF:]
			i++
		}

		g.locktmp = locktmp
	}

	r = tp

	return
}

func (g *generator) genSelectorExpr(expr *ast.SelectorExpr) (r int) {
	g.tokLine = expr.Pos().Line

	sp := g.sp

	x := g.genExpr(expr.X, genR)
	y := g.genName(expr.Sel, genKey)

	g.pushInst(opcode.ABC(opcode.GETTABLE, sp, x, y))

	r = sp

	// recover sp
	g.setSP(sp + 1)

	return
}

func (g *generator) genIndexExpr(expr *ast.IndexExpr) (r int) {
	g.tokLine = expr.Pos().Line

	sp := g.sp

	x := g.genExpr(expr.X, genR)
	y := g.genExpr(expr.Index, genR|genK)

	g.pushInst(opcode.ABC(opcode.GETTABLE, sp, x, y))

	r = sp

	// recover sp
	g.setSP(sp + 1)

	return
}

func (g *generator) genCallExprN(expr *ast.CallExpr, nrets int, isTail bool) (r int) {
	g.tokLine = expr.Pos().Line

	sp := g.sp

	var fn int

	nargs := len(expr.Args)

	locktmp := g.locktmp

	if expr.Colon != position.NoPos {
		self := g.genExpr(expr.X, genR)

		name := g.genName(expr.Name, genKey)

		g.pushInst(opcode.ABC(opcode.SELF, sp, self, name))

		nargs++

		fn = sp

		g.setSP(sp + 2)

		g.locktmp = true
	} else {
		g.locktmp = true

		fn = g.genExpr(expr.X, genMove)
	}

	var isVar bool
	if len(expr.Args) != 0 {
		for _, e := range expr.Args[:len(expr.Args)-1] {
			g.genExpr(e, genMove)
		}

		isVar = g.genExprN(expr.Args[len(expr.Args)-1], -1)
	}

	g.locktmp = locktmp

	if isTail {
		if isVar {
			g.pushInst(opcode.ABC(opcode.TAILCALL, fn, 0, 0))
		} else {
			g.pushInst(opcode.ABC(opcode.TAILCALL, fn, nargs+1, 0))
		}

		r = fn

		// recover sp
		g.setSP(sp)

		return
	}

	if isVar {
		g.pushInst(opcode.ABC(opcode.CALL, fn, 0, nrets+1))
	} else {
		g.pushInst(opcode.ABC(opcode.CALL, fn, nargs+1, nrets+1))
	}

	r = fn

	if nrets < 0 {
		// just update maxstacksize
		if g.sp+1 > g.MaxStackSize {
			g.MaxStackSize = g.sp + 1
		}

		return
	}

	// recover sp
	g.setSP(sp + nrets)

	return
}

func (g *generator) genUnaryExpr(expr *ast.UnaryExpr, typ genType) (r int) {
	g.tokLine = expr.Pos().Line

	if expr.Op == token.UNM {
		if x, ok := expr.X.(*ast.BasicLit); ok {
			tok := x.Token

			switch tok.Type {
			case token.INT:
				var val object.Value

				if i, ok := g.parseInteger(tok, true); ok {
					val = i
				} else {
					val = g.parseNumber(tok, true)
				}

				return g.genConst(val, typ)
			case token.FLOAT:
				val := g.parseNumber(tok, true)

				return g.genConst(val, typ)
			}
		}
	}

	sp := g.sp

	x := g.genExpr(expr.X, genR)

	switch expr.Op {
	case token.UNM:
		g.pushInst(opcode.AB(opcode.UNM, sp, x))
	case token.BNOT:
		g.pushInst(opcode.AB(opcode.BNOT, sp, x))
	case token.NOT:
		g.pushInst(opcode.AB(opcode.NOT, sp, x))
	case token.LEN:
		g.pushInst(opcode.AB(opcode.LEN, sp, x))
	default:
		panic("unreachable")
	}

	r = sp

	// recover sp
	g.setSP(sp + 1)

	return
}

func (g *generator) genBinaryExpr(expr *ast.BinaryExpr, typ genType) (r int) {
	g.tokLine = expr.Pos().Line

	sp := g.sp

	switch expr.Op {
	case token.EQ:
		x := g.genExpr(expr.X, genR|genK)
		y := g.genExpr(expr.Y, genR|genK)

		g.pushInst(opcode.ABC(opcode.EQ, 1, x, y))
		g.pushInst(opcode.AsBx(opcode.JMP, 0, 1))
		g.pushInst(opcode.ABC(opcode.LOADBOOL, sp, 0, 1))
		g.pushInst(opcode.ABC(opcode.LOADBOOL, sp, 1, 0))

		g.lockpeep = true
	case token.NE:
		x := g.genExpr(expr.X, genR|genK)
		y := g.genExpr(expr.Y, genR|genK)

		g.pushInst(opcode.ABC(opcode.EQ, 0, x, y))
		g.pushInst(opcode.AsBx(opcode.JMP, 0, 1))
		g.pushInst(opcode.ABC(opcode.LOADBOOL, sp, 0, 1))
		g.pushInst(opcode.ABC(opcode.LOADBOOL, sp, 1, 0))

		g.lockpeep = true
	case token.LT:
		x := g.genExpr(expr.X, genR|genK)
		y := g.genExpr(expr.Y, genR|genK)

		g.pushInst(opcode.ABC(opcode.LT, 1, x, y))
		g.pushInst(opcode.AsBx(opcode.JMP, 0, 1))
		g.pushInst(opcode.ABC(opcode.LOADBOOL, sp, 0, 1))
		g.pushInst(opcode.ABC(opcode.LOADBOOL, sp, 1, 0))

		g.lockpeep = true
	case token.LE:
		x := g.genExpr(expr.X, genR|genK)
		y := g.genExpr(expr.Y, genR|genK)

		g.pushInst(opcode.ABC(opcode.LE, 1, x, y))
		g.pushInst(opcode.AsBx(opcode.JMP, 0, 1))
		g.pushInst(opcode.ABC(opcode.LOADBOOL, sp, 0, 1))
		g.pushInst(opcode.ABC(opcode.LOADBOOL, sp, 1, 0))

		g.lockpeep = true
	case token.GT:
		x := g.genExpr(expr.X, genR|genK)
		y := g.genExpr(expr.Y, genR|genK)

		g.pushInst(opcode.ABC(opcode.LT, 1, y, x))
		g.pushInst(opcode.AsBx(opcode.JMP, 0, 1))
		g.pushInst(opcode.ABC(opcode.LOADBOOL, sp, 0, 1))
		g.pushInst(opcode.ABC(opcode.LOADBOOL, sp, 1, 0))

		g.lockpeep = true
	case token.GE:
		x := g.genExpr(expr.X, genR|genK)
		y := g.genExpr(expr.Y, genR|genK)

		g.pushInst(opcode.ABC(opcode.LE, 1, y, x))
		g.pushInst(opcode.AsBx(opcode.JMP, 0, 1))
		g.pushInst(opcode.ABC(opcode.LOADBOOL, sp, 0, 1))
		g.pushInst(opcode.ABC(opcode.LOADBOOL, sp, 1, 0))

		g.lockpeep = true
	case token.ADD:
		x := g.genExpr(expr.X, genR|genK)
		y := g.genExpr(expr.Y, genR|genK)

		g.pushInst(opcode.ABC(opcode.ADD, sp, x, y))
	case token.SUB:
		x := g.genExpr(expr.X, genR|genK)
		y := g.genExpr(expr.Y, genR|genK)

		g.pushInst(opcode.ABC(opcode.SUB, sp, x, y))
	case token.MUL:
		x := g.genExpr(expr.X, genR|genK)
		y := g.genExpr(expr.Y, genR|genK)

		g.pushInst(opcode.ABC(opcode.MUL, sp, x, y))
	case token.MOD:
		x := g.genExpr(expr.X, genR|genK)
		y := g.genExpr(expr.Y, genR|genK)

		g.pushInst(opcode.ABC(opcode.MOD, sp, x, y))
	case token.POW:
		x := g.genExpr(expr.X, genR|genK)
		y := g.genExpr(expr.Y, genR|genK)

		g.pushInst(opcode.ABC(opcode.POW, sp, x, y))
	case token.DIV:
		x := g.genExpr(expr.X, genR|genK)
		y := g.genExpr(expr.Y, genR|genK)

		g.pushInst(opcode.ABC(opcode.DIV, sp, x, y))
	case token.IDIV:
		x := g.genExpr(expr.X, genR|genK)
		y := g.genExpr(expr.Y, genR|genK)

		g.pushInst(opcode.ABC(opcode.IDIV, sp, x, y))
	case token.BAND:
		x := g.genExpr(expr.X, genR|genK)
		y := g.genExpr(expr.Y, genR|genK)

		g.pushInst(opcode.ABC(opcode.BAND, sp, x, y))
	case token.BOR:
		x := g.genExpr(expr.X, genR|genK)
		y := g.genExpr(expr.Y, genR|genK)

		g.pushInst(opcode.ABC(opcode.BOR, sp, x, y))
	case token.BXOR:
		x := g.genExpr(expr.X, genR|genK)
		y := g.genExpr(expr.Y, genR|genK)

		g.pushInst(opcode.ABC(opcode.BXOR, sp, x, y))
	case token.SHL:
		x := g.genExpr(expr.X, genR|genK)
		y := g.genExpr(expr.Y, genR|genK)

		g.pushInst(opcode.ABC(opcode.SHL, sp, x, y))
	case token.SHR:
		x := g.genExpr(expr.X, genR|genK)
		y := g.genExpr(expr.Y, genR|genK)

		g.pushInst(opcode.ABC(opcode.SHR, sp, x, y))
	case token.CONCAT:
		locktmp := g.locktmp

		g.locktmp = true

		x := g.genExpr(expr.X, genMove)

		Y := expr.Y
		for {
			ct, ok := Y.(*ast.BinaryExpr)
			if !ok || ct.Op != token.CONCAT {
				break
			}

			g.genExpr(ct.X, genMove)

			Y = ct.Y
		}

		y := g.genExpr(Y, genMove)

		g.locktmp = locktmp

		g.pushInst(opcode.ABC(opcode.CONCAT, sp, x, y))
	case token.AND:
		if _, ok := g.foldExpr(expr.X); ok {
			x := g.genExpr(expr.Y, genR)

			if typ&genMove == 0 {
				return x
			}

			g.pushInst(opcode.AB(opcode.MOVE, sp, x))
		} else {
			x := g.genExpr(expr.X, genR)

			g.pushInst(opcode.ABC(opcode.TESTSET, sp, x, 0))

			endJump := g.genJumpPoint()

			y := g.genExpr(expr.Y, genR)

			g.pushInst(opcode.AB(opcode.MOVE, sp, y))

			g.genJumpFrom(endJump)
		}
	case token.OR:
		if _, ok := g.foldExpr(expr.X); ok {
			y := g.genExpr(expr.Y, genR)

			if typ&genMove == 0 {
				return y
			}

			g.pushInst(opcode.AB(opcode.MOVE, sp, y))
		} else {
			x := g.genExpr(expr.X, genR)

			g.pushInst(opcode.ABC(opcode.TESTSET, sp, x, 1))

			endJump := g.genJumpPoint()

			y := g.genExpr(expr.Y, genR)

			g.pushInst(opcode.AB(opcode.MOVE, sp, y))

			g.genJumpFrom(endJump)
		}
	default:
		panic("unreachable")
	}

	r = sp

	// recover sp
	g.setSP(sp + 1)

	return
}
