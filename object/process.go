package object

type Process interface {
	// returns new Process which environment is inherited from parent
	Fork() Process

	Exec(p *Proto, args ...Value) (rets []Value, err error)

	NewTableSize(asize, msize int) Table
	NewTableArray(a []Value) Table
	NewThread() Thread
	NewGoThread() Thread
	NewClosure(p *Proto) Closure
	NewChannel(capacity int) Channel

	Globals() Table
	Loaded() Table
	Preload() Table

	GetMetatable(val Value) Table
	SetMetatable(val Value, mt Table)

	Select(cases []SelectCase) (chosen int, recv Value, recvOK bool)

	// aux APIs

	Require(name string, open GoFunction) (Value, bool)
}
