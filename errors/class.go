package errors

var (
	CompileError = newClass("CompileError")
	DumpError    = newClass("DumpError")
	UndumpError  = newClass("UndumpError")
	SyntaxError  = newClass("SyntaxError")
	RuntimeError = newClass("RuntimeError")
)
