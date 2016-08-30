package codegen

import (
	"github.com/hirochachacha/plua"
	"github.com/hirochachacha/plua/compiler/ast"
	"github.com/hirochachacha/plua/compiler/token"
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
		g.declareLocal(name.Name, sp+i)
	}

	g.setSP(sp + len(stmt.LHS))
}

func (g *generator) genLocalFuncStmt(stmt *ast.LocalFuncStmt) {
	name := stmt.Name

	body := stmt.Body

	endLine := stmt.End().Line

	g.declareLocal(name.Name, g.sp) // declare before genFuncBody (for recursive function)

	p := g.proto(body, false, endLine)

	g.pushInstLine(opcode.ABx(opcode.CLOSURE, g.sp, p), endLine)

	g.nextSP()
}

func (g *generator) genFuncStmt(stmt *ast.FuncStmt) {
	sp := g.sp

	name := stmt.Name
	prefix := stmt.NamePrefix

	body := stmt.Body

	endLine := stmt.End().Line

	if prefix == nil {
		l := g.resolve(name.Name)

		p := g.proto(body, false, endLine)

		if l == nil {
			g.pushInstLine(opcode.ABx(opcode.CLOSURE, g.sp, p), endLine)

			g.genSetGlobal(name, g.sp)
		} else {
			switch l.kind {
			case linkLocal:
				g.pushInstLine(opcode.ABx(opcode.CLOSURE, l.v, p), endLine)
			case linkUpval:
				g.pushInstLine(opcode.ABx(opcode.CLOSURE, g.sp, p), endLine)

				g.pushInst(opcode.AB(opcode.SETUPVAL, g.sp, l.v))
			default:
				panic("unreachable")
			}
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
	if _, ok := g.labels[stmt.Name.Name]; ok {
		panic("label already defined")
	}

	g.newLabel(stmt.Name.Name)

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
		l := g.resolve(lhs.Name)

		if l == nil {
			g.genSetGlobal(lhs, r)
		} else {
			switch l.kind {
			case linkLocal:
				g.pushInst(opcode.AB(opcode.MOVE, l.v, r))
			case linkUpval:
				g.pushInst(opcode.AB(opcode.SETUPVAL, r, l.v))
			default:
				panic("unreachable")
			}
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
			l := g.resolve(lhs.Name)

			if l == nil {
				rk := g.markRK(g.constant(object.String(lhs.Name)))

				env := g.genName(&ast.Name{Name: plua.LUA_ENV}, genR|genMove)

				assigns[i] = opcode.ABC(opcode.SETTABLE, env, rk, r)
			} else {
				switch l.kind {
				case linkLocal:
					assigns[i] = opcode.AB(opcode.MOVE, l.v, r)
				case linkUpval:
					assigns[i] = opcode.AB(opcode.SETUPVAL, r, l.v)
				default:
					panic("unreachable")
				}
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
	g.genJump(stmt.Label.Name)
}

func (g *generator) genBreakStmt(stmt *ast.BreakStmt) {
	g.genJump("@break")
}

func (g *generator) genIfStmt(stmt *ast.IfStmt) {
	g.openScope()

	switch g.genTest(stmt.Cond, false) {
	case immTrue:
		if stmt.Body != nil {
			g.genBlock(stmt.Body)
		}
	case immFalse:
		if stmt.Else != nil {
			g.genBlock(stmt.Else)
		}
	default:
		elseJump := g.genPendingLocalJump()

		if stmt.Body != nil {
			g.openScope()

			g.genBlock(stmt.Body)

			g.closeScope()
		}

		if stmt.Else != nil {
			doneJump := g.genPendingLocalJump()

			g.setLocalJumpDst(elseJump)

			g.openScope()

			g.genBlock(stmt.Else)

			g.closeScope()

			g.setLocalJumpDst(doneJump)
		} else {
			g.setLocalJumpDst(elseJump)
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

	initLabel := g.newLocalLabel()

	switch g.genTest(stmt.Cond, false) {
	case immTrue:
		if stmt.Body != nil {
			g.genBlock(stmt.Body)
		}

		g.genLocalJump(initLabel)

		g.newLabel("@break")
	case immFalse:
		// do nothing
	default:
		endJump := g.genPendingLocalJump()

		if stmt.Body != nil {
			g.genBlock(stmt.Body)
		}

		g.genLocalJump(initLabel)

		g.setLocalJumpDst(endJump)

		g.newLabel("@break")
	}

	g.closeScope()
}

func (g *generator) genRepeatStmt(stmt *ast.RepeatStmt) {
	g.openScope()

	initLabel := g.newLocalLabel()

	if stmt.Body != nil {
		g.genBlock(stmt.Body)
	}

	switch g.genTest(stmt.Cond, true) {
	case immTrue:
		g.genLocalJump(initLabel)

		g.newLabel("@break")
	case immFalse:
		g.newLabel("@break")
	default:
		endJump := g.genPendingLocalJump()

		g.genLocalJump(initLabel)

		g.setLocalJumpDst(endJump)

		g.newLabel("@break")
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

	g.openScope()

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

	g.declareLocal(stmt.Name.Name, g.sp)

	g.addSP(1)

	forprep := g.pushTempLine(forLine)

	if stmt.Body != nil {
		g.genBlock(stmt.Body)
	}

	g.closeScope()

	g.Code[forprep] = opcode.AsBx(opcode.FORPREP, sp, g.pc()-forprep-1)

	g.pushInstLine(opcode.AsBx(opcode.FORLOOP, sp, forprep-g.pc()), forLine)

	g.newLabel("@break")

	g.sp = sp

	g.closeScope()
}

func (g *generator) genForEachStmt(stmt *ast.ForEachStmt) {
	g.openScope()

	forLine := stmt.For.Line

	sp := g.sp

	g.openScope()

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

	for i, name := range stmt.Names {
		g.declareLocal(name.Name, g.sp+i)
	}

	g.addSP(len(stmt.Names))

	loopJump := g.genPendingLocalJump()

	init := g.pc()

	if stmt.Body != nil {
		g.genBlock(stmt.Body)
	}

	g.setLocalJumpDst(loopJump)

	g.closeScope()

	g.pushInstLine(opcode.AC(opcode.TFORCALL, sp, len(stmt.Names)), forLine)

	g.pushInstLine(opcode.AsBx(opcode.TFORLOOP, sp+2, init-g.pc()-1), forLine)

	g.newLabel("@break")

	g.closeScope()
}
