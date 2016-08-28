package ast

type visit func(node Node) (skip bool)

func Find(node Node, typ Type) (ret Node) {
	Walk(node, func(node Node) bool {
		if node.Type() == typ {
			ret = node

			return true
		}
		return false
	})

	return
}

func Walk(node Node, visit func(node Node) (skip bool)) {
	walk(node, visit)
}

func walkComment(comment *Comment, visit visit) bool {
	if comment == nil {
		return false
	}

	return visit(comment)
}

func walkName(name *Name, visit visit) bool {
	if name == nil {
		return false
	}

	return visit(name)
}

func walkParamList(params *ParamList, visit visit) bool {
	if params == nil {
		return false
	}

	if visit(params) {
		return true
	}

	for _, name := range params.List {
		if walkName(name, visit) {
			return true
		}
	}
	return false
}

func walkBlock(block *Block, visit visit) bool {
	if block == nil {
		return false
	}

	if visit(block) {
		return true
	}

	for _, e := range block.List {
		if walkStmt(e, visit) {
			return true
		}
	}
	return false
}

func walkFuncBody(fbody *FuncBody, visit visit) bool {
	if fbody == nil {
		return false
	}

	if visit(fbody) {
		return true
	}

	if walkParamList(fbody.Params, visit) {
		return true
	}
	return walkBlock(fbody.Body, visit)
}

func walkExpr(expr Expr, visit visit) bool {
	if expr == nil {
		return false
	}

	if visit(expr) {
		return true
	}

	switch expr := expr.(type) {
	case *BadExpr:
	case *Name:
	case *Vararg:
	case *BasicLit:
	case *FuncLit:
		return walkFuncBody(expr.Body, visit)
	case *TableLit:
		for _, e := range expr.Fields {
			if walkExpr(e, visit) {
				return true
			}
		}
	case *ParenExpr:
		return walkExpr(expr.X, visit)
	case *SelectorExpr:
		if walk(expr.X, visit) {
			return true
		}
		return walkName(expr.Sel, visit)
	case *IndexExpr:
		if walkExpr(expr.X, visit) {
			return true
		}
		return walkExpr(expr.Index, visit)
	case *CallExpr:
		if walkExpr(expr.X, visit) {
			return true
		}
		if walkName(expr.Name, visit) {
			return true
		}
		for _, arg := range expr.Args {
			if walkExpr(arg, visit) {
				return true
			}
		}
	case *UnaryExpr:
		return walkExpr(expr.X, visit)
	case *BinaryExpr:
		if walkExpr(expr.X, visit) {
			return true
		}
		return walkExpr(expr.Y, visit)
	case *KeyValueExpr:
		if walkExpr(expr.Key, visit) {
			return true
		}
		return walkExpr(expr.Value, visit)
	default:
		panic("unreachable")
	}

	return false
}

func walkStmt(stmt Stmt, visit visit) bool {
	if stmt == nil {
		return false
	}

	if visit(stmt) {
		return true
	}

	switch stmt := stmt.(type) {
	case *BadStmt:
	case *EmptyStmt:
	case *LocalAssignStmt:
		for _, name := range stmt.LHS {
			if walkName(name, visit) {
				return true
			}
		}
		for _, e := range stmt.RHS {
			if walkExpr(e, visit) {
				return true
			}
		}
	case *LocalFuncStmt:
		if walkName(stmt.Name, visit) {
			return true
		}
		return walkFuncBody(stmt.Body, visit)
	case *FuncStmt:
		for _, name := range stmt.NamePrefix {
			if walkName(name, visit) {
				return true
			}
		}
		if walkName(stmt.Name, visit) {
			return true
		}
		return walkFuncBody(stmt.Body, visit)
	case *LabelStmt:
		return walkName(stmt.Name, visit)
	case *ExprStmt:
		x := stmt.X

		if x == nil {
			return false
		}

		if walkExpr(x.X, visit) {
			return true
		}
		if walkName(x.Name, visit) {
			return true
		}
		for _, arg := range x.Args {
			if walkExpr(arg, visit) {
				return true
			}
		}
	case *AssignStmt:
		for _, e := range stmt.LHS {
			if walkExpr(e, visit) {
				return true
			}
		}
		for _, e := range stmt.RHS {
			if walkExpr(e, visit) {
				return true
			}
		}
	case *GotoStmt:
		return walkName(stmt.Label, visit)
	case *BreakStmt:
	case *IfStmt:
		if walkExpr(stmt.Cond, visit) {
			return true
		}
		if walkBlock(stmt.Body, visit) {
			return true
		}
		return walkBlock(stmt.Else, visit)
	case *DoStmt:
		return walkBlock(stmt.Body, visit)
	case *WhileStmt:
		if walkExpr(stmt.Cond, visit) {
			return true
		}
		return walkBlock(stmt.Body, visit)
	case *RepeatStmt:
		if walkExpr(stmt.Cond, visit) {
			return true
		}
		return walkBlock(stmt.Body, visit)
	case *ReturnStmt:
		for _, ret := range stmt.Results {
			if walkExpr(ret, visit) {
				return true
			}
		}
	case *ForStmt:
		if walkName(stmt.Name, visit) {
			return true
		}
		if walkExpr(stmt.Start, visit) {
			return true
		}
		if walkExpr(stmt.Finish, visit) {
			return true
		}
		if walkExpr(stmt.Step, visit) {
			return true
		}
		return walkBlock(stmt.Body, visit)
	case *ForEachStmt:
		for _, name := range stmt.Names {
			if walkName(name, visit) {
				return true
			}
		}
		for _, e := range stmt.Exprs {
			if walkExpr(e, visit) {
				return true
			}
		}
		return walkBlock(stmt.Body, visit)
	default:
		panic("unreachable")
	}

	return false
}

func walk(node Node, visit visit) (skip bool) {
	if node == nil {
		return false
	}

	if visit(node) {
		return true
	}

	switch node := node.(type) {
	case *Comment:
	case *CommentGroup:
		for _, c := range node.List {
			if walkComment(c, visit) {
				return true
			}
		}

	case *ParamList:
		for _, name := range node.List {
			if walkName(name, visit) {
				return true
			}
		}

	case *BadExpr:
	case *Name:
	case *Vararg:
	case *BasicLit:
	case *FuncLit:
		return walkFuncBody(node.Body, visit)
	case *TableLit:
		for _, e := range node.Fields {
			if walkExpr(e, visit) {
				return true
			}
		}
	case *ParenExpr:
		return walkExpr(node.X, visit)
	case *SelectorExpr:
		if walk(node.X, visit) {
			return true
		}
		return walkName(node.Sel, visit)
	case *IndexExpr:
		if walkExpr(node.X, visit) {
			return true
		}
		return walkExpr(node.Index, visit)
	case *CallExpr:
		if walkExpr(node.X, visit) {
			return true
		}
		if walkName(node.Name, visit) {
			return true
		}
		for _, arg := range node.Args {
			if walkExpr(arg, visit) {
				return true
			}
		}
	case *UnaryExpr:
		return walkExpr(node.X, visit)
	case *BinaryExpr:
		if walkExpr(node.X, visit) {
			return true
		}
		return walkExpr(node.Y, visit)
	case *KeyValueExpr:
		if walkExpr(node.Key, visit) {
			return true
		}
		return walkExpr(node.Value, visit)

	case *BadStmt:
	case *EmptyStmt:
	case *LocalAssignStmt:
		for _, name := range node.LHS {
			if walkName(name, visit) {
				return true
			}
		}
		for _, e := range node.RHS {
			if walkExpr(e, visit) {
				return true
			}
		}
	case *LocalFuncStmt:
		if walkName(node.Name, visit) {
			return true
		}
		return walkFuncBody(node.Body, visit)
	case *FuncStmt:
		for _, name := range node.NamePrefix {
			if walkName(name, visit) {
				return true
			}
		}
		if walkName(node.Name, visit) {
			return true
		}
		return walkFuncBody(node.Body, visit)
	case *LabelStmt:
		return walkName(node.Name, visit)
	case *ExprStmt:
		x := node.X

		if x == nil {
			return false
		}

		if walkExpr(x.X, visit) {
			return true
		}
		if walkName(x.Name, visit) {
			return true
		}
		for _, arg := range x.Args {
			if walkExpr(arg, visit) {
				return true
			}
		}
	case *AssignStmt:
		for _, e := range node.LHS {
			if walkExpr(e, visit) {
				return true
			}
		}
		for _, e := range node.RHS {
			if walkExpr(e, visit) {
				return true
			}
		}
	case *GotoStmt:
		return walkName(node.Label, visit)
	case *BreakStmt:
	case *IfStmt:
		if walkExpr(node.Cond, visit) {
			return true
		}
		if walkBlock(node.Body, visit) {
			return true
		}
		return walkBlock(node.Else, visit)
	case *DoStmt:
		return walkBlock(node.Body, visit)
	case *WhileStmt:
		if walkExpr(node.Cond, visit) {
			return true
		}
		return walkBlock(node.Body, visit)
	case *RepeatStmt:
		if walkExpr(node.Cond, visit) {
			return true
		}
		return walkBlock(node.Body, visit)
	case *ReturnStmt:
		for _, ret := range node.Results {
			if walkExpr(ret, visit) {
				return true
			}
		}
	case *ForStmt:
		if walkName(node.Name, visit) {
			return true
		}
		if walkExpr(node.Start, visit) {
			return true
		}
		if walkExpr(node.Finish, visit) {
			return true
		}
		if walkExpr(node.Step, visit) {
			return true
		}
		return walkBlock(node.Body, visit)
	case *ForEachStmt:
		for _, name := range node.Names {
			if walkName(name, visit) {
				return true
			}
		}
		for _, e := range node.Exprs {
			if walkExpr(e, visit) {
				return true
			}
		}
		return walkBlock(node.Body, visit)

	case *File:
		for _, e := range node.Chunk {
			if walkStmt(e, visit) {
				return true
			}
		}
	case *Block:
		for _, e := range node.List {
			if walkStmt(e, visit) {
				return true
			}
		}
		return false
	case *FuncBody:
		if walkParamList(node.Params, visit) {
			return true
		}
		return walkBlock(node.Body, visit)
	default:
		panic("unreachable")
	}

	return false
}
