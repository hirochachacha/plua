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

	// common APIs

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

	CallMetaField(val Value, field string) (res []Value, done bool)
	GetMetaField(val Value, field string) Value
	Require(name string, open GoFunction) (Value, bool)

	// ↓ thread specific APIs

	// ↓ intended to be called from vm loop

	Call(fn, errh Value, args ...Value) ([]Value, *RuntimeError)

	// ↓ for debug support

	Func() Value // returns top level Closure or GoFunction

	GetInfo(level int, what string) *DebugInfo
	GetInfoByFunc(fn Value, what string) *DebugInfo

	GetLocal(d *DebugInfo, n int) (name string, val Value)
	SetLocal(d *DebugInfo, n int, val Value) (name string)

	GetHook() (hook Value, mask string, count int)
	SetHook(hook Value, mask string, count int)

	// ↓ for coroutine & goroutine support

	LoadFunc(fn Value)
	Resume(args ...Value) (rets []Value, err *RuntimeError)

	// ↓ for coroutine support

	Yield(args ...Value) (rets []Value, err *RuntimeError)

	IsYieldable() bool
	IsMainThread() bool

	Status() ThreadStatus
}
