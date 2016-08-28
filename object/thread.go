package object

type ThreadStatus int

const (
	THREAD_INIT ThreadStatus = iota
	THREAD_SUSPENDED
	THREAD_ERROR
	THREAD_RETURN
	THREAD_RUNNING
)

type iThread interface {
	NewTableSize(asize, msize int) *Table
	NewTableArray(a []Value) *Table
	NewThread() *Thread
	NewGoThread() *Thread
	NewUserdata(x interface{}) *Userdata
	NewClosure(p *Proto) *Closure
	NewChannel(capacity int) *Channel

	Registry() *Table
	Globals() *Table
	Loaded() *Table
	Preload() *Table

	GetMetatable(val Value) *Table
	SetMetatable(val Value, mt *Table)

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

	// from aux

	// Requiref(openf GoFunction, modname string) bool

	// ToString(val Value) String

	// NewMetatableNameSize(tname string, alen, mlen int) *Table
	// GetMetatableName(tname string) *Table
	// GetMetaField(val Value, field string) Value
	// CallMetaField(val Value, field string) (rets []Value, done bool)

	// ArgError(fname string, n int, extramsg string)
	// TypeError(fname string, arg Value, n int, tname string)
	// OptionError(fname string, n int, opt string)

	// NewArgChecker(fname string, args []Value) *ArgChecker

	// Traceback(msg string, level int) string
}

func (th *Thread) NewTableSize(asize, msize int) *Table {
	return th.Impl.NewTableSize(asize, msize)
}

func (th *Thread) NewTableArray(a []Value) *Table {
	return th.Impl.NewTableArray(a)
}

func (th *Thread) NewThread() *Thread {
	return th.Impl.NewThread()
}

func (th *Thread) NewGoThread() *Thread {
	return th.Impl.NewGoThread()
}

func (th *Thread) NewUserdata(x interface{}) *Userdata {
	return th.Impl.NewUserdata(x)
}

func (th *Thread) NewClosure(p *Proto) *Closure {
	return th.Impl.NewClosure(p)
}

func (th *Thread) NewChannel(capacity int) *Channel {
	return th.Impl.NewChannel(capacity)
}

func (th *Thread) Registry() *Table {
	return th.Impl.Registry()
}

func (th *Thread) Globals() *Table {
	return th.Impl.Globals()
}

func (th *Thread) Loaded() *Table {
	return th.Impl.Loaded()
}

func (th *Thread) Preload() *Table {
	return th.Impl.Preload()
}

func (th *Thread) GetMetatable(val Value) *Table {
	return th.Impl.GetMetatable(val)
}

func (th *Thread) SetMetatable(val Value, mt *Table) {
	th.Impl.SetMetatable(val, mt)
}

func (th *Thread) Select(cases []SelectCase) (chosen int, recv Value, recvOK bool) {
	return th.Impl.Select(cases)
}

func (th *Thread) Func() Value {
	return th.Impl.Func()
}

func (th *Thread) Yield(args ...Value) []Value {
	return th.Impl.Yield(args...)
}

func (th *Thread) Resume(args ...Value) ([]Value, error) {
	return th.Impl.Resume(args...)
}

func (th *Thread) IsYieldable() bool {
	return th.Impl.IsYieldable()
}

func (th *Thread) IsMainThread() bool {
	return th.Impl.IsMainThread()
}

func (th *Thread) Status() ThreadStatus {
	return th.Impl.Status()
}

func (th *Thread) GetInfo(level int, what string) *DebugInfo {
	return th.Impl.GetInfo(level, what)
}

func (th *Thread) GetInfoByFunc(fn Value, what string) *DebugInfo {
	return th.Impl.GetInfoByFunc(fn, what)
}

func (th *Thread) GetLocal(d *DebugInfo, n int) (name string, val Value) {
	return th.Impl.GetLocal(d, n)
}

func (th *Thread) SetLocal(d *DebugInfo, n int, val Value) (name string) {
	return th.Impl.SetLocal(d, n, val)
}

func (th *Thread) GetHook() (hook Value, mask string, count int) {
	return th.Impl.GetHook()
}

func (th *Thread) SetHook(hook Value, mask string, count int) {
	th.Impl.SetHook(hook, mask, count)
}

func (th *Thread) Load(p *Proto) {
	th.Impl.Load(p)
}

func (th *Thread) LoadFunc(fn Value) {
	th.Impl.LoadFunc(fn)
}

func (th *Thread) Call(fn Value, args ...Value) ([]Value, bool) {
	return th.Impl.Call(fn, args...)
}

func (th *Thread) PCall(fn Value, errh Value, args ...Value) ([]Value, bool) {
	return th.Impl.PCall(fn, errh, args...)
}

func (th *Thread) Propagate(err *Error) {
	th.Impl.Propagate(err)
}
