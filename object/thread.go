package object

type ThreadStatus int

const (
	THREAD_INIT ThreadStatus = iota
	THREAD_SUSPENDED
	THREAD_ERROR
	THREAD_RETURN
	THREAD_RUNNING
)

type Thread interface {
	Value

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

	Select(cases []SelectCase) (chosen int, recv Value, recvOK bool)

	// ↓ thread specific APIs

	Func() Value // return closure or go function

	Load(p *Proto)
	LoadFunc(fn Value)
	Resume(args ...Value) (rets []Value, err error)

	IsYieldable() bool
	IsMainThread() bool

	Status() ThreadStatus

	// ↓ intended to call from vm loop

	GetInfo(level int, what string) *DebugInfo
	GetInfoByFunc(fn Value, what string) *DebugInfo

	GetLocal(d *DebugInfo, n int) (name string, val Value)
	SetLocal(d *DebugInfo, n int, val Value) (name string)

	GetHook() (hook Value, mask string, count int)
	SetHook(hook Value, mask string, count int)

	Call(fn Value, args ...Value) ([]Value, bool)
	PCall(fn Value, errh Value, args ...Value) ([]Value, bool)

	Yield(args ...Value) (rets []Value)
	Propagate(err *Error)
}
