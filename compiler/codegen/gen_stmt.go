package codegen

import (
	"fmt"

	"github.com/hirochachacha/plua/compiler/ast"
	"github.com/hirochachacha/plua/compiler/token"
	"github.com/hirochachacha/plua/internal/version"
	"github.com/hirochachacha/plua/object"
	"github.com/hirochachacha/plua/opcode"
)

type immBool int

const (
	immUndefined immBool = iota
	immTrue
	immFalse
)

func (g *generator) genStmt(stmt ast.Stmt) {
	switch stmt := stmt.(type) {
	case *ast.BadStmt:
		panic("bad stmt")
	case *ast.EmptyStmt:
		// do nothing
	case *ast.LocalAssignStmt:
		g.genLocalAssignStmt(stmt)
	case *ast.LocalFuncStmt:
		g.genLocalFuncStmt(stmt)
	case *ast.FuncStmt:
		g.genFuncStmt(stmt)
	case *ast.LabelStmt:
		g.genLabelStmt(stmt)
	case *ast.ExprStmt:
		g.genExprStmt(stmt)
	case *ast.AssignStmt:
		g.genAssignStmt(stmt)
	case *ast.GotoStmt:
		g.genGotoStmt(stmt)
	case *ast.BreakStmt:
		g.genBreakStmt(stmt)
	case *ast.IfStmt:
		g.genIfStmt(stmt)
	case *ast.DoStmt:
		g.genDoStmt(stmt)
	case *ast.WhileStmt:
		g.genWhileStmt(stmt)
	case *ast.RepeatStmt:
		g.genRepeatStmt(stmt)
	case *ast.ReturnStmt:
		g.genReturnStmt(stmt)
	case *ast.ForStmt:
		g.genForStmt(stmt)
	case *ast.ForEachStmt:
		g.genForEachStmt(stmt)
	default:
		panic("unreachable")
	}
}

func (g *generator) genLocalAssignStmt(stmt *ast.LocalAssignStmt) {
	sp := g.sp

	g.locktmp = true

	switch {
	case len(stmt.LHS) > len(stmt.RHS):
		if len(stmt.RHS) == 0 {
			g.pushInst(opcode.AB(opcode.LOADNIL, g.sp, len(stmt.LHS)-1))
		} else {
			for _, e := range stmt.RHS[:len(stmt.RHS)-1] {
				g.genExpr(e, genMove)
			}

			g.genExprN(stmt.RHS[len(stmt.RHS)-1], len(stmt.LHS)+1-len(stmt.RHS))
		}
	case len(stmt.LHS) < len(stmt.RHS):
		for i := range stmt.LHS {
			g.genExpr(stmt.RHS[i], genMove)
		}

		for i := len(stmt.LHS); i < len(stmt.RHS); i++ {
			if _, ok := g.foldExpr(stmt.RHS[i]); !ok {
				g.genExprN(stmt.RHS[i], 0)
			}
		}
	default:
		for i := range stmt.LHS {
			g.genExpr(stmt.RHS[i], genMove)
		}
	}

	g.locktmp = false

	for i, name := range stmt.LHS {
		g.declareLocalName(name, sp+i)
	}

	g.setSP(sp + len(stmt.LHS))
}

func (g *generator) genLocalFuncStmt(stmt *ast.LocalFuncStmt) {
	name := stmt.Name

	body := stmt.Body

	endLine := stmt.End().Line

	g.declareLocalName(name, g.sp) // declare before genFuncBody (for recursive function)

	p := g.proto(body, false, endLine)

	g.pushInstLine(opcode.ABx(opcode.CLOSURE, g.sp, p), endLine)

	g.LocVars[len(g.LocVars)-1].StartPC++ // adjust StartPC (start from CLOSURE)

	g.nextSP()
}

func (g *generator) genFuncStmt(stmt *ast.FuncStmt) {
	sp := g.sp

	name := stmt.Name
	prefix := stmt.PathList

	body := stmt.Body

	endLine := stmt.End().Line

	if prefix == nil {
		l, ok := g.resolveName(name)

		p := g.proto(body, false, endLine)

		if ok {
			switch l.kind {
			case linkLocal:
				g.pushInstLine(opcode.ABx(opcode.CLOSURE, l.index, p), endLine)
			case linkUpval:
				g.pushInstLine(opcode.ABx(opcode.CLOSURE, g.sp, p), endLine)

				g.pushInst(opcode.AB(opcode.SETUPVAL, g.sp, l.index))
			default:
				panic("unreachable")
			}
		} else {
			g.pushInstLine(opcode.ABx(opcode.CLOSURE, g.sp, p), endLine)

			g.genSetGlobal(name, g.sp)
		}
	} else {
		r := g.genPrefix(prefix)

		rk := g.genName(name, genKey)

		switch stmt.AccessTok {
		case token.COLON:
			p := g.proto(body, true, endLine)

			g.pushInstLine(opcode.ABx(opcode.CLOSURE, g.sp, p), endLine)

			g.pushInst(opcode.ABC(opcode.SETTABLE, r, rk, g.sp))
		case token.PERIOD:
			p := g.proto(body, false, endLine)

			g.pushInstLine(opcode.ABx(opcode.CLOSURE, g.sp, p), endLine)

			g.pushInst(opcode.ABC(opcode.SETTABLE, r, rk, g.sp))
		default:
			panic("unreachable")
		}
	}

	// recover sp
	g.sp = sp
}

func (g *generator) genLabelStmt(stmt *ast.LabelStmt) {
	nameNode := stmt.Name
	name := nameNode.Name
	pos := nameNode.Pos()

	if label, ok := g.labels[name]; ok {
		g.error(pos, fmt.Errorf("label '%s' already defined on %s", name, label.pos))
	}

	g.declareLabelPos(name, pos)

	g.lockpeep = true
}

func (g *generator) genExprStmt(stmt *ast.ExprStmt) {
	sp := g.sp

	g.genExprN(stmt.X, 0)

	// recover sp
	g.sp = sp
}

func (g *generator) genAssignStmt(stmt *ast.AssignStmt) {
	sp := g.sp

	g.locktmp = true

	switch {
	case len(stmt.LHS) > len(stmt.RHS):
		if len(stmt.RHS) == 0 {
			g.pushInst(opcode.AB(opcode.LOADNIL, sp, len(stmt.LHS)-1))
		} else {
			for _, e := range stmt.RHS[:len(stmt.RHS)-1] {
				g.genExpr(e, genMove)
			}

			g.genExprN(stmt.RHS[len(stmt.RHS)-1], len(stmt.LHS)+1-len(stmt.RHS))
		}
	case len(stmt.LHS) < len(stmt.RHS):
		for i := range stmt.LHS {
			g.genExpr(stmt.RHS[i], genMove)
		}

		for i := len(stmt.LHS); i < len(stmt.RHS); i++ {
			if _, ok := g.foldExpr(stmt.RHS[i]); !ok {
				g.genExprN(stmt.RHS[i], 0)
			}
		}
	default:
		for i := range stmt.LHS {
			g.genExpr(stmt.RHS[i], genMove)
		}
	}

	g.locktmp = false

	if len(stmt.LHS) == 1 { // fast path
		g.genAssignSimple(stmt.LHS[0], sp)
	} else {
		g.genAssign(stmt.LHS, sp)
	}

	g.sp = sp
}

func (g *generator) genAssignSimple(lhs ast.Expr, r int) {
	switch lhs := lhs.(type) {
	case *ast.Name:
		if l, ok := g.resolveName(lhs); ok {
			switch l.kind {
			case linkLocal:
				g.pushInst(opcode.AB(opcode.MOVE, l.index, r))
			case linkUpval:
				g.pushInst(opcode.AB(opcode.SETUPVAL, r, l.index))
			default:
				panic("unreachable")
			}
		} else {
			g.genSetGlobal(lhs, r)
		}
	case *ast.SelectorExpr:
		x := g.genExpr(lhs.X, genR|genK)
		y := g.genName(lhs.Sel, genKey)

		g.pushInst(opcode.ABC(opcode.SETTABLE, x, y, r))
	case *ast.IndexExpr:
		x := g.genExpr(lhs.X, genR|genK)
		y := g.genExpr(lhs.Index, genR|genK)

		g.pushInst(opcode.ABC(opcode.SETTABLE, x, y, r))
	default:
		panic("unreachable")
	}
}

func (g *generator) genAssign(LHS []ast.Expr, base int) {
	assigns := make([]opcode.Instruction, len(LHS))

	g.locktmp = true

	var r int
	for i, lhs := range LHS {
		r = base + i

		switch lhs := lhs.(type) {
		case *ast.Name:
			if l, ok := g.resolveName(lhs); ok {
				switch l.kind {
				case linkLocal:
					assigns[i] = opcode.AB(opcode.MOVE, l.index, r)
				case linkUpval:
					assigns[i] = opcode.AB(opcode.SETUPVAL, r, l.index)
				default:
					panic("unreachable")
				}
			} else {
				rk := g.markRK(g.constant(object.String(lhs.Name)))

				env := g.genName(tmpName(version.LUA_ENV), genR|genMove)

				assigns[i] = opcode.ABC(opcode.SETTABLE, env, rk, r)
			}
		case *ast.SelectorExpr:
			x := g.genExpr(lhs.X, genR|genK|genMove)
			y := g.genName(lhs.Sel, genKey)

			assigns[i] = opcode.ABC(opcode.SETTABLE, x, y, r)
		case *ast.IndexExpr:
			x := g.genExpr(lhs.X, genR|genK|genMove)
			y := g.genExpr(lhs.Index, genR|genK|genMove)

			assigns[i] = opcode.ABC(opcode.SETTABLE, x, y, r)
		default:
			panic("unreachable")
		}
	}

	for _, assign := range assigns[:len(assigns)-1] {
		g.pushInst(assign)
	}

	g.locktmp = false

	g.pushInst(assigns[len(assigns)-1])
}

func (g *generator) genGotoStmt(stmt *ast.GotoStmt) {
	g.genSetJumpPoint(stmt.Label.Name, stmt.Label.Pos())
}

func (g *generator) genBreakStmt(stmt *ast.BreakStmt) {
	g.genSetJumpPoint("@break", stmt.Break)
}

func (g *generator) genIfStmt(stmt *ast.IfStmt) {
	if stmt.ElseIfList != nil {
		elseBody := stmt.ElseBody
		root := &ast.IfStmt{
			Cond: stmt.Cond,
			Body: stmt.Body,
		}
		leaf := root
		for _, stmt := range stmt.ElseIfList {
			nextLeaf := &ast.IfStmt{
				Cond: stmt.Cond,
				Body: stmt.Body,
			}
			leaf.ElseBody = &ast.Block{
				List: []ast.Stmt{
					nextLeaf,
				},
			}
			leaf = nextLeaf
		}
		leaf.ElseBody = elseBody
		stmt = root
	}

	g.openScope()

	switch g.genTest(stmt.Cond, false) {
	case immTrue:
		if stmt.Body != nil {
			g.genBlock(stmt.Body)
		}
	case immFalse:
		if stmt.ElseBody != nil {
			g.genBlock(stmt.ElseBody)
		}
	default:
		elseJump := g.genJumpPoint()

		if stmt.Body != nil {
			g.openScope()

			g.genBlock(stmt.Body)

			g.closeScope()
		}

		if stmt.ElseBody != nil {
			doneJump := g.genJumpPoint()

			g.genJumpFrom(elseJump)

			g.openScope()

			g.genBlock(stmt.ElseBody)

			g.closeScope()

			g.genJumpFrom(doneJump)
		} else {
			g.genJumpFrom(elseJump)
		}
	}

	g.closeScope()
}

func (g *generator) genDoStmt(stmt *ast.DoStmt) {
	if stmt.Body != nil {
		g.openScope()

		g.genBlock(stmt.Body)

		g.closeScope()
	}
}

func (g *generator) genWhileStmt(stmt *ast.WhileStmt) {
	g.openScope()

	initLabel := g.newLabel()

	switch g.genTest(stmt.Cond, false) {
	case immTrue:
		if stmt.Body != nil {
			g.genBlock(stmt.Body)
		}

		g.closeScope()

		g.genJumpTo(initLabel)

		g.declareLabel("@break")
	case immFalse:
		g.closeScope()
	default:
		endJump := g.genJumpPoint()

		if stmt.Body != nil {
			g.genBlock(stmt.Body)
		}

		g.closeScope()

		g.genJumpTo(initLabel)

		g.genJumpFrom(endJump)

		g.declareLabel("@break")
	}
}

func (g *generator) genRepeatStmt(stmt *ast.RepeatStmt) {
	g.openScope()

	initLabel := g.newLabel()

	if stmt.Body != nil {
		g.genBlock(stmt.Body)
	}

	switch g.genTest(stmt.Cond, true) {
	case immTrue:
		g.genJumpTo(initLabel)

		g.declareLabel("@break")
	case immFalse:
		g.declareLabel("@break")
	default:
		endJump := g.genJumpPoint()

		g.genJumpTo(initLabel)

		g.genJumpFrom(endJump)

		g.declareLabel("@break")
	}

	g.closeScope()
}

func (g *generator) genReturnStmt(stmt *ast.ReturnStmt) {
	if len(stmt.Results) == 0 {
		g.pushInst(opcode.AB(opcode.RETURN, 0, 1))

		return
	}

	sp := g.sp

	if len(stmt.Results) == 1 {
		if tail, ok := stmt.Results[len(stmt.Results)-1].(*ast.CallExpr); ok {
			g.genCallExprN(tail, -1, true)

			g.pushInst(opcode.AB(opcode.RETURN, sp, 0))

			return
		}
	}

	g.locktmp = true

	var isVar bool
	if len(stmt.Results) != 0 {
		for _, e := range stmt.Results[:len(stmt.Results)-1] {
			g.genExpr(e, genMove)
		}

		isVar = g.genExprN(stmt.Results[len(stmt.Results)-1], -1)
	}

	if isVar {
		g.pushInst(opcode.AB(opcode.RETURN, sp, 0))
	} else {
		g.pushInst(opcode.AB(opcode.RETURN, sp, len(stmt.Results)+1))
	}

	g.locktmp = false
}

func (g *generator) genForStmt(stmt *ast.ForStmt) {
	g.openScope()

	forLine := stmt.For.Line

	sp := g.sp

	g.locktmp = true

	g.genExpr(stmt.Start, genMove)
	g.genExpr(stmt.Finish, genMove)
	if stmt.Step != nil {
		g.genExpr(stmt.Step, genMove)
	} else {
		g.genConst(object.Integer(1), genMove)
	}

	g.locktmp = false

	g.declareLocal("(for index)", sp)
	g.declareLocal("(for limit)", sp+1)
	g.declareLocal("(for step)", sp+2)

	forprep := g.pushTempLine(forLine)

	g.openScope()

	g.declareLocalName(stmt.Name, g.sp)

	g.addSP(1)

	if stmt.Body != nil {
		g.genBlock(stmt.Body)
	}

	g.closeScope()

	g.Code[forprep] = opcode.AsBx(opcode.FORPREP, sp, g.pc()-forprep-1)

	g.pushInstLine(opcode.AsBx(opcode.FORLOOP, sp, forprep-g.pc()), forLine)

	g.declareLabel("@break")

	g.sp = sp

	g.closeScope()
}

func (g *generator) genForEachStmt(stmt *ast.ForEachStmt) {
	g.openScope()

	forLine := stmt.For.Line

	sp := g.sp

	g.locktmp = true

	switch {
	case len(stmt.Exprs) > 3:
		g.genExpr(stmt.Exprs[0], genMove)
		g.genExpr(stmt.Exprs[1], genMove)
		g.genExpr(stmt.Exprs[2], genMove)
		for i := 3; i < len(stmt.Exprs); i++ {
			if _, ok := g.foldExpr(stmt.Exprs[i]); !ok {
				g.genExprN(stmt.Exprs[i], 0)
			}
		}
	case len(stmt.Exprs) < 3:
		switch len(stmt.Exprs) {
		case 1:
			g.genExprN(stmt.Exprs[0], 3)
		case 2:
			g.genExpr(stmt.Exprs[0], genMove)
			g.genExprN(stmt.Exprs[1], 2)
		default:
			panic("unreachable")
		}
	default:
		for _, e := range stmt.Exprs {
			g.genExpr(e, genMove)
		}
	}

	g.locktmp = false

	g.declareLocal("(for generator)", sp)
	g.declareLocal("(for generator)", sp+1)
	g.declareLocal("(for control)", sp+2)

	g.setSP(sp + 3)

	loopJump := g.genJumpPoint()

	g.openScope()

	for i, name := range stmt.Names {
		g.declareLocalName(name, g.sp+i)
	}

	g.addSP(len(stmt.Names))

	init := g.pc()

	if stmt.Body != nil {
		g.genBlock(stmt.Body)
	}

	g.genJumpFrom(loopJump)

	g.closeScope()

	g.pushInstLine(opcode.AC(opcode.TFORCALL, sp, len(stmt.Names)), forLine)

	g.pushInstLine(opcode.AsBx(opcode.TFORLOOP, sp+2, init-g.pc()-1), forLine)

	g.declareLabel("@break")

	g.closeScope()
}
