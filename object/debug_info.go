package object

type DebugInfo struct {
	Name            string
	NameWhat        string
	What            string
	Source          string
	CurrentLine     int
	LineDefined     int
	LastLineDefined int
	NUpvalues       int
	NParams         int
	IsVararg        bool
	IsTailCall      bool
	ShortSource     string
	Lines           *Table
	Func            Value       // go function or closure
	CallInfo        interface{} // implementation detail
}
