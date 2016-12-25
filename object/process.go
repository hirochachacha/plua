package object

type Process interface {
	// returns new Process which environment is inherited from parent
	Fork() Process

	Exec(p *Proto, args ...Value) (rets []Value, err error)

	NewTableSize(asize, msize int) Table
	NewClosure(p *Proto) Closure

	Globals() Table
	Loaded() Table
	Preload() Table

	GetMetatable(val Value) Table
	SetMetatable(val Value, mt Table)

	// aux API

	Require(name string, open GoFunction) (Value, bool)
}
