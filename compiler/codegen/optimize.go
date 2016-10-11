package codegen

import (
	"github.com/hirochachacha/plua/compiler/ast"
	"github.com/hirochachacha/plua/compiler/token"
	"github.com/hirochachacha/plua/internal/arith"
	"github.com/hirochachacha/plua/object"
	"github.com/hirochachacha/plua/opcode"
)

// constant folding

func (g *generator) foldExpr(expr ast.Expr) (val object.Value, ok bool) {
	if skipConstantFolding {
		return nil, false
	}

	switch expr := expr.(type) {
	case *ast.ParenExpr:
		return g.foldExpr(expr.X)
	case *ast.BasicLit:
		return g.foldBasic(expr)
	case *ast.UnaryExpr:
		return g.foldUnary(expr)
	case *ast.BinaryExpr:
		return g.foldBinary(expr)
	}

	return nil, false
}

func (g *generator) foldBasic(expr *ast.BasicLit) (val object.Value, ok bool) {
	if skipConstantFolding {
		return nil, false
	}

	g.tokLine = expr.Pos().Line

	if c, ok := g.cfolds[expr]; ok {
		return c, true
	}

	tok := expr.Token

	switch tok.Type {
	case token.NIL:
		val = nil
	case token.FALSE:
		val = object.False
	case token.TRUE:
		val = object.True
	case token.INT:
		i, inf := parseInteger(tok.Lit)
		if inf != 0 {
			if inf > 0 {
				val = object.Infinity
			} else {
				val = object.Infinity
			}
		} else {
			val = i
		}
	case token.FLOAT:
		val = parseNumber(tok.Lit)
	case token.STRING:
		val = object.String(unquoteString(tok.Lit))
	default:
		panic("unreachable")
	}

	g.cfolds[expr] = val

	return val, true
}

func (g *generator) foldUnary(expr *ast.UnaryExpr) (val object.Value, ok bool) {
	if skipConstantFolding {
		return nil, false
	}

	g.tokLine = expr.Pos().Line

	if c, ok := g.cfolds[expr]; ok {
		return c, c != nil
	}

	if expr.Op == token.UNM {
		if x, ok := expr.X.(*ast.BasicLit); ok {
			tok := x.Token

			switch tok.Type {
			case token.INT:
				var val object.Value

				i, inf := parseInteger("-" + tok.Lit)
				if inf != 0 {
					if inf > 0 {
						val = object.Infinity
					} else {
						val = object.Infinity
					}
				} else {
					val = i
				}

				g.cfolds[expr] = val

				return val, true
			case token.FLOAT:
				val := parseNumber("-" + tok.Lit)

				g.cfolds[expr] = val

				return val, true
			}
		}
	}

	switch expr.Op {
	case token.UNM:
		if x, ok := g.foldExpr(expr.X); ok {
			if unm := arith.Unm(x); unm != nil {
				val = unm
			}
		}
	case token.BNOT:
		if x, ok := g.foldExpr(expr.X); ok {
			if bnot := arith.Bnot(x); bnot != nil {
				val = bnot
			}
		}
	case token.NOT:
		if x, ok := g.foldExpr(expr.X); ok {
			if not := arith.Not(x); not != nil {
				val = not
			}
		}
	case token.LEN:
		x := expr.X

		for paren, ok := x.(*ast.ParenExpr); ok; {
			x = paren.X
		}

		switch x := x.(type) {
		case *ast.BasicLit:
			tok := x.Token
			if tok.Type == token.STRING {
				val = object.Integer(len(unquoteString(tok.Lit)))
			}
		case *ast.TableLit:
			var a []ast.Expr
			for _, e := range x.Fields {
				if _, ok := e.(*ast.KeyValueExpr); !ok {
					if _, ok := g.foldExpr(e); !ok {
						g.cfolds[expr] = nil

						return nil, false
					}
					a = append(a, e)
				}
			}

			for len(a) > 0 {
				if lit, ok := a[len(a)-1].(*ast.BasicLit); ok && lit.Token.Type == token.NIL {
					a = a[:len(a)-1]

					continue
				}

				break
			}

			val = object.Integer(len(a))
		}
	default:
		panic("unreachable")
	}

	g.cfolds[expr] = val

	return val, val != nil
}

func (g *generator) foldBinary(expr *ast.BinaryExpr) (val object.Value, ok bool) {
	if skipConstantFolding {
		return nil, false
	}

	g.tokLine = expr.Pos().Line

	if c, ok := g.cfolds[expr]; ok {
		return c, c != nil
	}

	switch expr.Op {
	case token.EQ:
		if x, ok := g.foldExpr(expr.X); ok {
			if y, ok := g.foldExpr(expr.Y); ok {
				if b := arith.Equal(x, y); b != nil {
					val = b
				}
			}
		}
	case token.NE:
		if x, ok := g.foldExpr(expr.X); ok {
			if y, ok := g.foldExpr(expr.Y); ok {
				if b := arith.NotEqual(x, y); b != nil {
					val = b
				}
			}
		}
	case token.LT:
		if x, ok := g.foldExpr(expr.X); ok {
			if y, ok := g.foldExpr(expr.Y); ok {
				if b := arith.LessThan(x, y); b != nil {
					val = b
				}
			}
		}
	case token.LE:
		if x, ok := g.foldExpr(expr.X); ok {
			if y, ok := g.foldExpr(expr.Y); ok {
				if b := arith.LessThanOrEqualTo(x, y); b != nil {
					val = b
				}
			}
		}
	case token.GT:
		if x, ok := g.foldExpr(expr.X); ok {
			if y, ok := g.foldExpr(expr.Y); ok {
				if b := arith.GreaterThan(x, y); b != nil {
					val = b
				}
			}
		}
	case token.GE:
		if x, ok := g.foldExpr(expr.X); ok {
			if y, ok := g.foldExpr(expr.Y); ok {
				if b := arith.GreaterThanOrEqualTo(x, y); b != nil {
					val = b
				}
			}
		}

	case token.ADD:
		if x, ok := g.foldExpr(expr.X); ok {
			if y, ok := g.foldExpr(expr.Y); ok {
				if sum := arith.Add(x, y); sum != nil {
					val = sum
				}
			}
		}
	case token.SUB:
		if x, ok := g.foldExpr(expr.X); ok {
			if y, ok := g.foldExpr(expr.Y); ok {
				if diff := arith.Sub(x, y); diff != nil {
					val = diff
				}
			}
		}
	case token.MUL:
		if x, ok := g.foldExpr(expr.X); ok {
			if y, ok := g.foldExpr(expr.Y); ok {
				if prod := arith.Mul(x, y); prod != nil {
					val = prod
				}
			}
		}
	case token.MOD:
		if x, ok := g.foldExpr(expr.X); ok {
			if y, ok := g.foldExpr(expr.Y); ok {
				if rem, _ := arith.Mod(x, y); rem != nil {
					val = rem
				}
			}
		}
	case token.POW:
		if x, ok := g.foldExpr(expr.X); ok {
			if y, ok := g.foldExpr(expr.Y); ok {
				if prod := arith.Pow(x, y); prod != nil {
					val = prod
				}
			}
		}
	case token.DIV:
		if x, ok := g.foldExpr(expr.X); ok {
			if y, ok := g.foldExpr(expr.Y); ok {
				if quo := arith.Div(x, y); quo != nil {
					val = quo
				}
			}
		}
	case token.IDIV:
		if x, ok := g.foldExpr(expr.X); ok {
			if y, ok := g.foldExpr(expr.Y); ok {
				if quo, _ := arith.Idiv(x, y); quo != nil {
					val = quo
				}
			}
		}
	case token.BAND:
		if x, ok := g.foldExpr(expr.X); ok {
			if y, ok := g.foldExpr(expr.Y); ok {
				if band := arith.Band(x, y); band != nil {
					val = band
				}
			}
		}
	case token.BOR:
		if x, ok := g.foldExpr(expr.X); ok {
			if y, ok := g.foldExpr(expr.Y); ok {
				if bor := arith.Bor(x, y); bor != nil {
					val = bor
				}
			}
		}
	case token.BXOR:
		if x, ok := g.foldExpr(expr.X); ok {
			if y, ok := g.foldExpr(expr.Y); ok {
				if bxor := arith.Bxor(x, y); bxor != nil {
					val = bxor
				}
			}
		}
	case token.SHL:
		if x, ok := g.foldExpr(expr.X); ok {
			if y, ok := g.foldExpr(expr.Y); ok {
				if shl := arith.Shl(x, y); shl != nil {
					val = shl
				}
			}
		}
	case token.SHR:
		if x, ok := g.foldExpr(expr.X); ok {
			if y, ok := g.foldExpr(expr.Y); ok {
				if shr := arith.Shr(x, y); shr != nil {
					val = shr
				}
			}
		}
	case token.CONCAT:
		if x, ok := g.foldExpr(expr.X); ok {
			if y, ok := g.foldExpr(expr.Y); ok {
				if con := arith.Concat(x, y); con != nil {
					val = con
				}
			}
		}
	case token.AND:
		if x, ok := g.foldExpr(expr.X); ok {
			if object.ToBoolean(x) {
				return g.foldExpr(expr.Y)
			}

			return x, true
		}
	case token.OR:
		if x, ok := g.foldExpr(expr.X); ok {
			if object.ToBoolean(x) {
				return x, true
			}

			return g.foldExpr(expr.Y)
		}
	default:
		panic("unreachable")
	}

	g.cfolds[expr] = val

	return val, val != nil
}

// peep hole optimization

// optimize load to same address twice
func (g *generator) peepLoad(i0 opcode.Instruction, i opcode.Instruction) bool {
	a := i.A()

	switch i0.OpCode() {
	case opcode.MOVE, opcode.LOADK, opcode.GETUPVAL, opcode.NEWTABLE, opcode.CLOSURE:

		a0 := i0.A()

		if a0 == a { // local x = any; local x = 1 => local x = 1
			return true
		}
	case opcode.LOADBOOL:
		c0 := i0.C()

		if c0 == 0 {
			a0 := i0.A()

			if a0 == a { // local x = true; local x = 1 => local x = 1
				return true
			}
		}
	case opcode.LOADNIL:
		a0 := i0.A()
		b0 := i0.B()

		if a0 == a && b0 == 0 { // local x = nil; local x = 1 => local x = 1
			return true
		}
	}

	return false
}

func (g *generator) peepLine(i opcode.Instruction, line int) {
	var offset int
	var i0 opcode.Instruction

	for {
		if len(g.Code) > offset {
			i0 = g.Code[len(g.Code)-offset-1]
			op0 := i0.OpCode()

			switch op := i.OpCode(); op {
			case opcode.LOADNIL:
				a := i.A()
				b := i.B()

				switch op0 {
				case opcode.MOVE, opcode.LOADK, opcode.GETUPVAL, opcode.NEWTABLE, opcode.CLOSURE:
					a0 := i0.A()

					if a <= a0 && a0 <= a+b { // local x = any; x = nil => local x = nil
						offset++

						continue
					}
				case opcode.LOADBOOL:
					c0 := i0.C()

					if c0 == 0 {
						a0 := i0.A()

						if a <= a0 && a0 <= a+b { // local x = true; x = nil => local x = nil
							offset++

							continue
						}
					}
				case opcode.LOADNIL:
					a0 := i0.A()
					b0 := i0.B()

					// local x = nil; x = nil => local x = nil
					switch {
					case a0 < a:
						if a <= a0+b0 {
							if a+b > a0+b0 {
								i = opcode.AB(opcode.LOADNIL, a0, a+b-a0)

								offset++

								continue
							}

							g.Code = g.Code[:len(g.Code)-offset]
							g.LineInfo = g.LineInfo[:len(g.LineInfo)-offset]

							return
						}

						if a == a0+b0+1 {
							i = opcode.AB(opcode.LOADNIL, a0, a+b-a0)

							offset++

							continue
						}

						if a == a0+1 {
							i = opcode.AB(opcode.LOADNIL, a0, a+b-a0)

							offset++

							continue
						}
					case a < a0:
						if a0 <= a+b {
							if a+b > a0+b0 {
								offset++

								continue
							}

							i = opcode.AB(opcode.LOADNIL, a, a0+b0-a)

							offset++

							continue
						}

						if a0 == a+b+1 {
							i = opcode.AB(opcode.LOADNIL, a, a0+b0-a)

							offset++

							continue
						}

						if a0 == a+1 {
							i = opcode.AB(opcode.LOADNIL, a, a0+b0-a)

							offset++

							continue
						}
					default: // a == a0
						if b > b0 {
							offset++

							continue
						}

						g.Code = g.Code[:len(g.Code)-offset]
						g.LineInfo = g.LineInfo[:len(g.LineInfo)-offset]

						return
					}
				}
			case opcode.MOVE:
				a := i.A()
				b := i.B()

				if a == b { // local x = local x => none
					g.Code = g.Code[:len(g.Code)-offset]
					g.LineInfo = g.LineInfo[:len(g.LineInfo)-offset]

					return
				}

				if g.peepLoad(i0, i) {
					offset++

					continue
				}

				switch op0 {
				case opcode.MOVE:
					a0 := i0.A()
					b0 := i0.B()

					if a0 == b {
						if b0 == a { // local x = local y; y = x => local x = local y
							g.Code = g.Code[:len(g.Code)-offset]
							g.LineInfo = g.LineInfo[:len(g.LineInfo)-offset]

							return
						}

						if !g.locktmp && a0 >= g.nlocals { // tmp x = local y; local z = tmp x => local z = local y
							b0 := i0.B()

							i = opcode.AB(op0, a, b0)

							offset++

							continue
						}
					}
				case opcode.LOADK:
					a0 := i0.A()

					if !g.locktmp && a0 >= g.nlocals {
						if a0 == b { // tmp x = 1; local z = tmp x => local z = 1
							bx0 := i0.Bx()

							i = opcode.ABx(op0, a, bx0)

							offset++

							continue
						}
					}
				case opcode.LOADBOOL:
					c0 := i0.C()

					if c0 == 0 {
						a0 := i0.A()

						if !g.locktmp && a0 >= g.nlocals {
							if a0 == b { // tmp x = true; local z = tmp x => local z = true
								b0 := i0.B()

								i = opcode.ABC(op0, a, b0, 0)

								offset++

								continue
							}
						}
					}
				case opcode.LOADNIL:
					a0 := i0.A()
					b0 := i0.B()

					if a0 <= b && b <= a0+b0 {
						if a0 <= a && a <= a0+b0 {
							g.Code = g.Code[:len(g.Code)-offset]
							g.LineInfo = g.LineInfo[:len(g.LineInfo)-offset]

							return
						}

						if a == a0-1 {
							i = opcode.AB(op0, a, a0+b0-a)

							offset++

							continue
						}

						if a == a0+b0+1 {
							i = opcode.AB(op0, a0, b0+1)

							offset++

							continue
						}

						if !g.locktmp && a0 >= g.nlocals {
							if b0 == 0 { // tmp x = nil; local z = tmp x => local z = nil
								i = opcode.AB(op0, a, 0)

								offset++

								continue
							}
						}
					}
				case opcode.GETUPVAL:
					a0 := i0.A()

					if !g.locktmp && a0 >= g.nlocals {
						if a0 == b { // tmp x = lexical y; local z = tmp x => local z = lexical y
							b0 := i0.B()

							i = opcode.AB(op0, a, b0)

							offset++

							continue
						}
					}
				case opcode.GETTABLE:
					a0 := i0.A()

					if !g.locktmp && a0 >= g.nlocals {
						if a0 == b { // tmp x = y.foo; local z = tmp x => local z = y.foo
							b0 := i0.B()
							c0 := i0.C()

							i = opcode.ABC(op0, a, b0, c0)

							offset++

							continue
						}
					}
				case opcode.GETTABUP:
					a0 := i0.A()

					if !g.locktmp && a0 >= g.nlocals {
						if a0 == b { // tmp x = global y; local z = tmp x => local z = global y
							b0 := i0.B()
							c0 := i0.C()

							i = opcode.ABC(op0, a, b0, c0)

							offset++

							continue
						}
					}
				case opcode.NEWTABLE:
					a0 := i0.A()

					if !g.locktmp && a0 >= g.nlocals {
						if a0 == b { // tmp x = <table>; local z = tmp x => local z = <table>
							b0 := i0.B()
							c0 := i0.C()

							i = opcode.ABC(op0, a, b0, c0)

							offset++

							continue
						}
					}
				case opcode.CLOSURE:
					a0 := i0.A()

					if !g.locktmp && a0 >= g.nlocals {
						if a0 == b { // tmp x = <func>; local z = tmp x => local z = <func>
							bx0 := i0.Bx()

							i = opcode.ABx(op0, a, bx0)

							offset++

							continue
						}
					}
				case opcode.UNM, opcode.BNOT, opcode.NOT, opcode.LEN:
					a0 := i0.A()

					if !g.locktmp && a0 >= g.nlocals {
						if a0 == b { // tmp x = - local y; local z = tmp x => local z = - local y
							b0 := i0.B()

							i = opcode.AB(op0, a, b0)

							offset++

							continue
						}
					}
				case opcode.ADD, opcode.SUB, opcode.MUL, opcode.MOD, opcode.POW, opcode.DIV,
					opcode.IDIV, opcode.BAND, opcode.BOR, opcode.BXOR, opcode.SHL, opcode.SHR:
					a0 := i0.A()

					if !g.locktmp && a0 >= g.nlocals {
						if a0 == b { // tmp x = tmp y + tmp z; local w = tmp x => local w = tmp y + tmp z
							b0 := i0.B()
							c0 := i0.C()

							i = opcode.ABC(op0, a, b0, c0)

							offset++

							continue
						}
					}
				case opcode.CONCAT:
					a0 := i0.A()

					if !g.locktmp && a0 >= g.nlocals {
						if a0 == b { // tmp x = - local y; local z = tmp x => local z = - local y
							b0 := i0.B()
							c0 := i0.C()

							i = opcode.ABC(op0, a, b0, c0)

							offset++

							continue
						}
					}
				}
			case opcode.LOADBOOL:
				c0 := i0.C()

				if c0 == 0 {
					if g.peepLoad(i0, i) {
						offset++

						continue
					}
				}
			case opcode.LOADK, opcode.NEWTABLE, opcode.CLOSURE:
				if g.peepLoad(i0, i) {
					offset++

					continue
				}
			case opcode.GETUPVAL:
				if g.peepLoad(i0, i) {
					offset++

					continue
				}

				a := i.A()
				b := i.B()

				switch op0 {
				case opcode.SETUPVAL:
					a0 := i0.A()
					b0 := i0.B()

					if a0 == a && b0 == b { // lexical x = local y; local y = lexical x => lexical x = local y
						g.Code = g.Code[:len(g.Code)-offset]
						g.LineInfo = g.LineInfo[:len(g.LineInfo)-offset]

						return
					}
				}
			case opcode.GETTABUP:
				a := i.A()
				c := i.C()

				if a != c {
					if g.peepLoad(i0, i) {
						offset++

						continue
					}
				}
			case opcode.GETTABLE:
				a := i.A()
				b := i.B()
				c := i.C()

				if a != b && a != c {
					if g.peepLoad(i0, i) {
						offset++

						continue
					}
				}

				switch op0 {
				case opcode.GETUPVAL:
					a0 := i0.A()

					if a0 == a {
						b0 := i0.B()

						i = opcode.ABC(opcode.GETTABUP, a, b0, c)

						offset++

						continue
					}

					if !g.locktmp && a0 >= g.nlocals {
						if a0 == b { // tmp x = lexical env; local z = (tmp x).foo => local z = global foo
							b0 := i0.B()

							i = opcode.ABC(opcode.GETTABUP, a, b0, c)

							offset++

							continue
						}
					}
				}
			case opcode.ADD, opcode.SUB, opcode.MUL, opcode.MOD, opcode.POW,
				opcode.DIV, opcode.IDIV, opcode.BAND, opcode.BOR, opcode.BXOR, opcode.SHL, opcode.SHR:
				a := i.A()
				b := i.B()
				c := i.C()

				if a != b && a != c {
					if g.peepLoad(i0, i) {
						offset++

						continue
					}
				}
			case opcode.UNM, opcode.BNOT, opcode.NOT, opcode.LEN:
				a := i.A()
				b := i.B()

				if a != b {
					if g.peepLoad(i0, i) {
						offset++

						continue
					}
				}
			case opcode.SETUPVAL:
				a := i.A()
				b := i.B()

				switch op0 {
				case opcode.MOVE:
					a0 := i0.A()

					if a == a0 {
						if !g.locktmp && a >= g.nlocals {
							b0 := i0.B()

							i = opcode.AB(opcode.SETUPVAL, b0, b)

							offset++

							continue
						}
					}
				case opcode.SETUPVAL:
					b0 := i0.B()

					if b0 == b { // lexical x = local y; lexical x = local z => lexical x = local z
						offset++

						continue
					}
				case opcode.GETUPVAL:
					a0 := i0.A()
					b0 := i0.B()

					if a0 == a && b0 == b { // lexical x = local y; local y = lexical x => lexical x = local y
						g.Code = g.Code[:len(g.Code)-offset]
						g.LineInfo = g.LineInfo[:len(g.LineInfo)-offset]

						return
					}
				}
			case opcode.SETTABLE:
				a := i.A()
				b := i.B()
				c := i.C()

				switch op0 {
				case opcode.MOVE:
					a0 := i0.A()

					if a0 == c {
						a := i.A()
						b := i.B()
						b0 := i0.B()

						i = opcode.ABC(op, a, b, b0)

						offset++

						continue
					}
				case opcode.LOADK:
					a0 := i0.A()
					bx0 := i0.Bx()

					if bx0 <= opcode.MaxBx && a0 == c {
						a := i.A()
						b := i.B()

						i = opcode.ABC(op, a, b, bx0|opcode.BitRK)

						offset++

						continue
					}
				case opcode.GETUPVAL:
					a0 := i0.A()

					if !g.locktmp && a0 >= g.nlocals {
						if a0 == a { // tmp x = lexical env;  (tmp x).foo = local z => global foo = local z
							b0 := i0.B()

							i = opcode.ABC(opcode.SETTABUP, b0, b, c)

							offset++

							continue
						}
					}
				}
			case opcode.SETTABUP:
				c := i.C()

				switch op0 {
				case opcode.MOVE:
					a0 := i0.A()

					if a0 == c {
						a := i.A()
						b := i.B()
						b0 := i0.B()

						i = opcode.ABC(op, a, b, b0)

						offset++

						continue
					}
				case opcode.LOADK:
					a0 := i0.A()
					bx0 := i0.Bx()

					if bx0 <= opcode.MaxBx && a0 == c {
						a := i.A()
						b := i.B()

						i = opcode.ABC(op, a, b, bx0|opcode.BitRK)

						offset++

						continue
					}
				}
			case opcode.JMP:
				a := i.A()
				b := i.B()

				if a == 0 && b == 0 {
					g.Code = g.Code[:len(g.Code)-offset]
					g.LineInfo = g.LineInfo[:len(g.LineInfo)-offset]

					return
				}
			}

			if offset == 0 {
				g.Code = append(g.Code, i)
				g.LineInfo = append(g.LineInfo, line)
			} else {
				g.Code[len(g.Code)-offset] = i
				g.LineInfo[len(g.LineInfo)-offset] = line

				g.Code = g.Code[:len(g.Code)-offset+1]
				g.LineInfo = g.LineInfo[:len(g.LineInfo)-offset+1]
			}

			return

		}

		if offset == 0 {
			g.Code = append(g.Code, i)
			g.LineInfo = append(g.LineInfo, line)
		} else {
			g.Code[len(g.Code)-offset] = i
			g.LineInfo[len(g.LineInfo)-offset] = line

			g.Code = g.Code[:len(g.Code)-offset+1]
			g.LineInfo = g.LineInfo[:len(g.LineInfo)-offset+1]
		}

		return
	}

	panic("unreachable")

	return
}
