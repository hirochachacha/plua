package object

type Process interface {
	// returns new Process which environment is inherited from parent
	Fork() Process

	Exec(p *Proto, args ...Value) (rets []Value, err error)

	NewTableSize(asize, msize int) Table
	NewTableArray(a []Value) Table
	NewThread() Thread
	NewGoThread() Thread
	NewUserdata(x interface{}) Userdata
	NewClosure(p *Proto) Closure
	NewChannel(capacity int) Channel

	Registry() Table
	Globals() Table
	Loaded() Table
	Preload() Table

	GetMetatable(val Value) Table
	SetMetatable(val Value, mt Table)
}
