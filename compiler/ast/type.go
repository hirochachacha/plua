package ast

type Type int

func (typ Type) String() string {
	return astNames[typ]
}

const (
	COMMENT Type = iota
	COMMENT_GROUP

	PARAM_LIST

	BAD_EXPR
	NAME
	VARARG
	BASIC_LIT
	FUNC_LIT
	TABLE_LIT
	PAREN_EXPR
	SELECTOR_EXPR
	INDEX_EXPR
	CALL_EXPR
	UNARY_EXPR
	BINARY_EXPR
	KEY_VALUE_EXPR

	BAD_STMT
	EMPTY_STMT
	LOCAL_ASSIGN_STMT
	LOCAL_FUNC_STMT
	FUNC_STMT
	LABEL_STMT
	EXPR_STMT
	ASSIGN_STMT
	GOTO_STMT
	BREAK_STMT
	IF_STMT
	DO_STMT
	WHILE_STMT
	REPEAT_STMT
	RETURN_STMT
	FOR_STMT
	FOR_EACH_STMT

	FILE
	BLOCK
	FUNC_BODY
)

var astNames = [...]string{
	COMMENT:       "Comment",
	COMMENT_GROUP: "CommentGroup",

	PARAM_LIST: "ParamList",

	BAD_EXPR:       "BadExpr",
	NAME:           "Name",
	VARARG:         "Vararg",
	BASIC_LIT:      "BasicLit",
	FUNC_LIT:       "FuncLit",
	TABLE_LIT:      "TableLit",
	PAREN_EXPR:     "ParenExpr",
	SELECTOR_EXPR:  "SelectorExpr",
	INDEX_EXPR:     "IndexExpr",
	CALL_EXPR:      "CallExpr",
	UNARY_EXPR:     "UnaryExpr",
	BINARY_EXPR:    "BinaryExpr",
	KEY_VALUE_EXPR: "KeyValueExpr",

	BAD_STMT:          "BadStmt",
	EMPTY_STMT:        "EmptyStmt",
	LOCAL_ASSIGN_STMT: "LocalAssignStmt",
	LOCAL_FUNC_STMT:   "LocalFuncStmt",
	FUNC_STMT:         "FuncStmt",
	LABEL_STMT:        "LabelStmt",
	EXPR_STMT:         "ExprStmt",
	ASSIGN_STMT:       "AssignStmt",
	GOTO_STMT:         "GotoStmt",
	BREAK_STMT:        "BreakStmt",
	IF_STMT:           "IfStmt",
	DO_STMT:           "DoStmt",
	WHILE_STMT:        "WhileStmt",
	REPEAT_STMT:       "RepeatStmt",
	RETURN_STMT:       "ReturnStmt",
	FOR_STMT:          "ForStmt",
	FOR_EACH_STMT:     "ForEachStmt",

	FILE:      "File",
	BLOCK:     "Block",
	FUNC_BODY: "FuncBody",
}
