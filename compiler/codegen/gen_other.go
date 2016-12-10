package codegen

import (
	"github.com/hirochachacha/plua/compiler/ast"
	"github.com/hirochachacha/plua/compiler/token"
	"github.com/hirochachacha/plua/internal/version"
	"github.com/hirochachacha/plua/object"
	"github.com/hirochachacha/plua/opcode"
)

// gen with global scope
func (g *generator) genFile(f *ast.File) {
	g.newScope()

	g.LastLineDefined = f.End().Line

	g.IsVararg = true

	g.declareEnviron()

	for _, e := range f.Chunk {
		g.genStmt(e)
	}

	g.pushReturn()

	g.closeScope()

	if g.scope != nil {
		panic("unexpected")
	}

	g.closeJumps()
}

func (g *generator) genFuncBody(f *ast.FuncBody, self bool, endLine int) {
	g.newScope()

	g.LineDefined = f.Pos().Line

	if self {
		g.declareLocal("self", 0)

		for i, param := range f.Params.List {
			g.declareLocalName(param, i+1)
		}

		g.NParams = len(f.Params.List) + 1
	} else {
		for i, param := range f.Params.List {
			g.declareLocalName(param, i)
		}

		g.NParams = len(f.Params.List)
	}

	g.addSP(g.NParams)

	g.IsVararg = f.Params.Ellipsis.IsValid()

	if f.Body != nil {
		g.genBlock(f.Body)
	}

	g.LastLineDefined = endLine

	g.pushReturn()

	g.closeScope()

	g.closeJumps()
}

func (g *generator) genBlock(b *ast.Block) {
	for _, e := range b.List {
		g.genStmt(e)
	}
}

// gen test condtion or immediate bool
func (g *generator) genTest(cond ast.Expr, not bool) immBool {
	switch cond := cond.(type) {
	case *ast.ParenExpr:
		return g.genTest(cond.X, not)

	case *ast.BasicLit:
		if !skipDeadCodeElimination {
			switch cond.Token.Type {
			case token.FALSE, token.NIL:
				if not {
					return immTrue
				}
				return immFalse
			default:
				if not {
					return immFalse
				}
				return immTrue
			}
		}

		x := g.genExpr(cond, genR)
		if not {
			g.pushInst(opcode.AC(opcode.TEST, x, 1))
		} else {
			g.pushInst(opcode.AC(opcode.TEST, x, 0))
		}
	case *ast.UnaryExpr:
		switch cond.Op {
		case token.NOT:
			if !skipDeadCodeElimination {
				if val, ok := g.foldExpr(cond.X); ok {
					if !object.ToGoBool(val) {
						if not {
							return immFalse
						}
						return immTrue
					}

					if not {
						return immTrue
					}
					return immFalse
				}
			}

			return g.genTest(cond.X, !not)
		default:
			x := g.genExpr(cond, genR)
			if not {
				g.pushInst(opcode.AC(opcode.TEST, x, 1))
			} else {
				g.pushInst(opcode.AC(opcode.TEST, x, 0))
			}
		}
	case *ast.BinaryExpr:
		if !skipDeadCodeElimination {
			if val, ok := g.foldBinary(cond); ok {
				if object.ToGoBool(val) {
					if not {
						return immFalse
					}
					return immTrue
				}

				if not {
					return immTrue
				}
				return immFalse
			}
		}

		switch cond.Op {
		case token.EQ:
			x := g.genExpr(cond.X, genR|genK)
			y := g.genExpr(cond.Y, genR|genK)

			if not {
				g.pushInst(opcode.ABC(opcode.EQ, 1, x, y))
			} else {
				g.pushInst(opcode.ABC(opcode.EQ, 0, x, y))
			}
		case token.NE:
			x := g.genExpr(cond.X, genR|genK)
			y := g.genExpr(cond.Y, genR|genK)

			if not {
				g.pushInst(opcode.ABC(opcode.EQ, 0, x, y))
			} else {
				g.pushInst(opcode.ABC(opcode.EQ, 1, x, y))
			}
		case token.LT:
			x := g.genExpr(cond.X, genR|genK)
			y := g.genExpr(cond.Y, genR|genK)

			if not {
				g.pushInst(opcode.ABC(opcode.LT, 1, x, y))
			} else {
				g.pushInst(opcode.ABC(opcode.LT, 0, x, y))
			}
		case token.LE:
			x := g.genExpr(cond.X, genR|genK)
			y := g.genExpr(cond.Y, genR|genK)

			if not {
				g.pushInst(opcode.ABC(opcode.LE, 1, x, y))
			} else {
				g.pushInst(opcode.ABC(opcode.LE, 0, x, y))
			}
		case token.GT:
			x := g.genExpr(cond.X, genR|genK)
			y := g.genExpr(cond.Y, genR|genK)

			if not {
				g.pushInst(opcode.ABC(opcode.LT, 1, y, x))
			} else {
				g.pushInst(opcode.ABC(opcode.LT, 0, y, x))
			}
		case token.GE:
			x := g.genExpr(cond.X, genR|genK)
			y := g.genExpr(cond.Y, genR|genK)

			if not {
				g.pushInst(opcode.ABC(opcode.LE, 1, y, x))
			} else {
				g.pushInst(opcode.ABC(opcode.LE, 0, y, x))
			}
		case token.AND:
			if !skipDeadCodeElimination {
				if val, ok := g.foldExpr(cond.X); ok {
					if !object.ToGoBool(val) {
						if not {
							return immTrue
						}
						return immFalse
					}

					return g.genTest(cond.Y, not)
				}
			}

			x := g.genExpr(cond, genR)
			if not {
				g.pushInst(opcode.AC(opcode.TEST, x, 1))
			} else {
				g.pushInst(opcode.AC(opcode.TEST, x, 0))
			}
		case token.OR:
			if !skipDeadCodeElimination {
				if val, ok := g.foldExpr(cond.X); ok {
					if object.ToGoBool(val) {
						if not {
							return immFalse
						}
						return immTrue
					}

					return g.genTest(cond.Y, not)
				}
			}

			x := g.genExpr(cond, genR)
			if not {
				g.pushInst(opcode.AC(opcode.TEST, x, 1))
			} else {
				g.pushInst(opcode.AC(opcode.TEST, x, 0))
			}
		default:
			x := g.genExpr(cond, genR)
			if not {
				g.pushInst(opcode.AC(opcode.TEST, x, 1))
			} else {
				g.pushInst(opcode.AC(opcode.TEST, x, 0))
			}
		}
	// no optimization
	default:
		x := g.genExpr(cond, genR)
		if not {
			g.pushInst(opcode.AC(opcode.TEST, x, 1))
		} else {
			g.pushInst(opcode.AC(opcode.TEST, x, 0))
		}
	}

	return immUndefined
}

func (g *generator) genConst(val object.Value, typ genType) (rk int) {
	k := g.constant(val)

	if typ&genK != 0 {
		return g.markRK(k)
	}

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

func (g *generator) genSetGlobal(name *ast.Name, r int) {
	env, ok := g.resolve(version.LUA_ENV)
	if !ok {
		panic("unexpected")
	}

	rk := g.markRK(g.constant(object.String(name.Name)))

	switch env.kind {
	case linkLocal:
		g.pushInst(opcode.ABC(opcode.SETTABLE, env.index, rk, r))
	case linkUpval:
		g.pushInst(opcode.ABC(opcode.SETTABUP, env.index, rk, r))
	default:
		panic("unreachable")
	}
}

func (g *generator) genGetGlobal(name *ast.Name) (r int) {
	env, ok := g.resolve(version.LUA_ENV)
	if !ok {
		panic("unexpected")
	}

	rk := g.markRK(g.constant(object.String(name.Name)))

	switch env.kind {
	case linkLocal:
		g.pushInst(opcode.ABC(opcode.GETTABLE, g.sp, env.index, rk))
	case linkUpval:
		g.pushInst(opcode.ABC(opcode.GETTABUP, g.sp, env.index, rk))
	default:
		panic("unreachable")
	}

	r = g.sp

	g.nextSP()

	return
}

func (g *generator) genPrefix(prefix []*ast.Name) (r int) {
	sp := g.sp

	r = g.genName(prefix[0], genR)
	for _, name := range prefix[1:] {
		y := g.genName(name, genKey)

		g.pushInst(opcode.ABC(opcode.GETTABLE, sp, r, y))

		r = sp
	}

	if r == sp {
		g.setSP(sp + 1)
	}

	return
}
