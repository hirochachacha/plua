package token

type Type uint

func (typ Type) String() string {
	return tokenNames[typ]
}

const (
	ILLEGAL Type = iota

	EOF
	COMMENT
	LABEL    // ::
	ELLIPSIS // ...

	literal_beg

	NAME

	INT
	FLOAT
	STRING

	literal_end
	operator_beg

	ADD    // +
	SUB    // -
	MUL    // *
	MOD    // %
	POW    // ^
	DIV    // /
	IDIV   // //
	BAND   // &
	BOR    // |
	BXOR   // ~
	SHL    // <<
	SHR    // >>
	CONCAT // ..

	// UNM    // -
	// BNOT   // ~
	LEN // #

	ASSIGN // =

	EQ // ==
	NE // ~=
	LT // <
	LE // <=
	GT // >
	GE // >=

	LPAREN    // (
	RPAREN    // )
	LBRACK    // [
	RBRACK    // ]
	LBRACE    // {
	RBRACE    // }
	COLON     // :
	COMMA     // ,
	SEMICOLON // ;
	PERIOD    // .

	operator_end
	keyword_beg

	AND
	BREAK
	DO
	ELSE
	ELSEIF
	END
	FALSE
	FOR
	FUNCTION
	GOTO
	IF
	IN
	LOCAL
	NIL
	NOT
	OR
	REPEAT
	RETURN
	THEN
	TRUE
	UNTIL
	WHILE

	keyword_end
)

// alias
const (
	UNM  = SUB
	BNOT = BXOR
)

var tokenNames = [...]string{
	ILLEGAL: "ILLEGAL",

	EOF:      "EOF",
	COMMENT:  "COMMENT",
	LABEL:    "::",
	ELLIPSIS: "...",

	NAME:   "NAME",
	INT:    "INT",
	FLOAT:  "FLOAT",
	STRING: "STRING",

	ADD:    "+",
	SUB:    "-",
	MUL:    "*",
	MOD:    "%",
	POW:    "^",
	DIV:    "/",
	IDIV:   "//",
	BAND:   "&",
	BOR:    "|",
	BXOR:   "~",
	SHL:    "<<",
	SHR:    ">>",
	CONCAT: "..",

	// UNM:  "-",
	// BNOT: "~",
	LEN: "#",

	ASSIGN: "=",

	EQ: "==",
	NE: "~=",
	LT: "<",
	LE: "<=",
	GT: ">",
	GE: ">=",

	LPAREN:    "(",
	RPAREN:    ")",
	LBRACK:    "[",
	RBRACK:    "]",
	LBRACE:    "{",
	RBRACE:    "}",
	COLON:     ":",
	COMMA:     ",",
	SEMICOLON: ";",
	PERIOD:    ".",

	AND:      "and",
	BREAK:    "break",
	DO:       "do",
	ELSE:     "else",
	ELSEIF:   "elseif",
	END:      "end",
	FALSE:    "false",
	FOR:      "for",
	FUNCTION: "function",
	GOTO:     "goto",
	IF:       "if",
	IN:       "in",
	LOCAL:    "local",
	NIL:      "nil",
	NOT:      "not",
	OR:       "or",
	REPEAT:   "repeat",
	RETURN:   "return",
	THEN:     "then",
	TRUE:     "true",
	UNTIL:    "until",
	WHILE:    "while",
}

var keywords = map[string]Type{
	"and":      AND,
	"break":    BREAK,
	"do":       DO,
	"else":     ELSE,
	"elseif":   ELSEIF,
	"end":      END,
	"false":    FALSE,
	"for":      FOR,
	"function": FUNCTION,
	"goto":     GOTO,
	"if":       IF,
	"in":       IN,
	"local":    LOCAL,
	"nil":      NIL,
	"not":      NOT,
	"or":       OR,
	"repeat":   REPEAT,
	"return":   RETURN,
	"then":     THEN,
	"true":     TRUE,
	"until":    UNTIL,
	"while":    WHILE,
}
