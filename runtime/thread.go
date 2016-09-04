package runtime

import (
	"fmt"

	"github.com/hirochachacha/plua/object"
)

type threadType int

const (
	threadMain threadType = iota
	threadCo
	threadGo
)

type thread struct {
	*context

	env *environment
	typ threadType

	resume chan []object.Value
	yield  chan []object.Value

	depth int
}

func (th *thread) err() error {
	ctx := th.context

	if ctx.isRoot() {
		return ctx.err()
	}

	return nil
}

func (th *thread) Type() object.Type {
	return object.TTHREAD
}

func (th *thread) String() string {
	return fmt.Sprintf("thread: %p", th)
}

func (th *thread) Yield(args ...object.Value) (rets []object.Value) {
	switch th.typ {
	case threadMain:
		th.throwYieldMainThreadError()
	case threadCo:
		if th.status == object.THREAD_RUNNING {
			th.status = object.THREAD_SUSPENDED

			th.yield <- args

			rets = <-th.resume

			th.status = object.THREAD_RUNNING
		} else {
			th.throwYieldFromOutsideError()
		}
	case threadGo:
		th.throwYieldGoThreadError()
	default:
		panic("unreachable")
	}

	return
}

func (th *thread) Resume(args ...object.Value) (rets []object.Value, err error) {
	switch th.typ {
	case threadMain, threadCo:
		if !(th.status == object.THREAD_INIT || th.status == object.THREAD_SUSPENDED) {
			return nil, errDeadCoroutine
		}

		th.resume <- args

		rets, ok := <-th.yield
		if !ok {
			err = th.err()
		}
		return rets, err
	case threadGo:
		if th.status != object.THREAD_INIT {
			return nil, errGoroutineTwice
		}

		th.resume <- args
	default:
		panic("unreachable")
	}

	return
}

func (th *thread) IsYieldable() bool {
	return th.typ == threadCo && th.status == object.THREAD_RUNNING
}

func (th *thread) IsMainThread() bool {
	return th.typ == threadMain
}

func (th *thread) Status() object.ThreadStatus {
	return th.status
}

func (th *thread) NewThread() object.Thread {
	return th.newThreadWith(threadCo, th.env, 0)
}

func (th *thread) NewGoThread() object.Thread {
	return th.newThreadWith(threadGo, th.env, 0)
}

func (th *thread) NewTableSize(asize, msize int) object.Table {
	return newTableSize(asize, msize)
}

func (th *thread) NewTableArray(a []object.Value) object.Table {
	return newTableArray(a)
}

func (th *thread) NewClosure(p *object.Proto) object.Closure {
	return th.newClosure(p)
}

func (th *thread) NewChannel(capacity int) object.Channel {
	return newChannel(capacity)
}

func (th *thread) Func() object.Value {
	return th.fn()
}

func (th *thread) Globals() object.Table {
	return th.env.globals
}

func (th *thread) Loaded() object.Table {
	return th.env.loaded
}

func (th *thread) Preload() object.Table {
	return th.env.preload
}

func (th *thread) GetMetatable(val object.Value) object.Table {
	return th.env.getMetatable(val)
}

func (th *thread) SetMetatable(val object.Value, mt object.Table) {
	th.env.setMetatable(val, mt)
}

func (th *thread) GetInfo(level int, what string) *object.DebugInfo {
	return th.getInfo(level, what)
}

func (th *thread) GetInfoByFunc(fn object.Value, what string) *object.DebugInfo {
	return th.getInfoByFunc(fn, what)
}

func (th *thread) GetLocal(d *object.DebugInfo, n int) (name string, val object.Value) {
	return th.getLocal(d, n)
}

func (th *thread) SetLocal(d *object.DebugInfo, n int, val object.Value) (name string) {
	return th.setLocal(d, n, val)
}

func (th *thread) GetHook() (hook object.Value, mask string, count int) {
	if th.hookMask&maskCall != 0 {
		mask += "c"
	}

	if th.hookMask&maskLine != 0 {
		mask += "l"
	}

	if th.hookMask&maskReturn != 0 {
		mask += "r"
	}

	return th.hookFunc, mask, th.hookCount
}

func (th *thread) SetHook(hook object.Value, mask string, count int) {
	if !isfunction(hook) {
		panic("unexpected")
	}

	th.hookFunc = hook

	var bitmask maskType
	for _, r := range mask {
		switch r {
		case 'c':
			bitmask |= maskCall
		case 'l':
			bitmask |= maskLine
		case 'r':
			bitmask |= maskReturn
		default:
			panic("unreachable")
		}
	}

	if count > 0 {
		bitmask |= maskCount

		th.hookCount = count
	} else {
		th.hookCount = 0
	}

	th.hookMask = bitmask
}

func (th *thread) LoadFunc(fn object.Value) {
	switch fn := fn.(type) {
	case object.GoFunction:
		th.loadfn(fn)
	case object.Closure:
		th.loadfn(fn)
	default:
		panic("unexpected")
	}
}

func (th *thread) Call(fn object.Value, args ...object.Value) ([]object.Value, bool) {
	return th.docallv(fn, args...)
}

func (th *thread) PCall(fn object.Value, errh object.Value, args ...object.Value) ([]object.Value, bool) {
	return th.dopcallv(fn, errh, args...)
}

func (th *thread) newThreadWith(typ threadType, env *environment, stackSize int) *thread {
	if stackSize < minStackSize {
		stackSize = minStackSize
	}

	newth := &thread{
		typ:    typ,
		env:    env,
		resume: make(chan []object.Value, 0),
		yield:  make(chan []object.Value, 0),
		depth:  th.depth,
	}

	newth.pushContext(stackSize)

	go newth.execute()

	return newth
}

func newMainThread() *thread {
	th := &thread{
		typ:    threadMain,
		env:    newEnvironment(),
		resume: make(chan []object.Value, 0),
		yield:  make(chan []object.Value, 0),
	}

	th.pushContext(basicStackSize)

	go th.execute()

	return th
}
