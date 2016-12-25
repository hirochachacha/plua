package runtime

import (
	"fmt"

	"github.com/hirochachacha/plua/object"
)

type threadType uint

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

	hookMask  maskType
	hookFunc  object.Value
	instCount int
	hookCount int
	lastLine  int

	depth int
}

func (th *thread) Type() object.Type {
	return object.TTHREAD
}

func (th *thread) String() string {
	return fmt.Sprintf("thread: %p", th)
}

func (th *thread) Yield(args ...object.Value) (rets []object.Value, err *object.RuntimeError) {
	switch th.typ {
	case threadMain:
		return nil, object.NewRuntimeError("attempt to yield a main thread")
	case threadCo:
		if th.status != object.THREAD_RUNNING {
			return nil, object.NewRuntimeError("attempt to yield from outside a coroutine")
		}

		th.status = object.THREAD_SUSPENDED

		th.yield <- args

		rets = <-th.resume

		th.status = object.THREAD_RUNNING

		return rets, nil
	case threadGo:
		return nil, object.NewRuntimeError("attempt to yield a goroutine")
	default:
		panic("unreachable")
	}
}

func (th *thread) Resume(args ...object.Value) (rets []object.Value, err *object.RuntimeError) {
	switch th.typ {
	case threadMain, threadCo:
		if !(th.status == object.THREAD_INIT || th.status == object.THREAD_SUSPENDED) {
			return nil, object.NewRuntimeError("cannot resume dead coroutine")
		}

		th.resume <- args

		rets, ok := <-th.yield
		if !ok {
			ctx := th.context
			if ctx.isRoot() {
				err := ctx.err
				if err != nil {
					return nil, err
				}
			}
		}
		return rets, nil
	case threadGo:
		if th.status != object.THREAD_INIT {
			return nil, object.NewRuntimeError("goroutine is already resumed")
		}

		th.resume <- args
	default:
		panic("unreachable")
	}

	return nil, nil
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

func (th *thread) GetLocal(level, n int) (name string, val object.Value) {
	return th.getLocal(level, n)
}

func (th *thread) SetLocal(level, n int, val object.Value) (name string) {
	return th.setLocal(level, n, val)
}

func (th *thread) GetLocalName(fn object.Value, n int) (name string) {
	mustFunction(fn)

	if cl, ok := fn.(object.Closure); ok {
		return getLocalName(cl.Prototype(), 0, n)
	}

	return ""
}

func (th *thread) GetHook() (hook object.Value, mask string, count int) {
	if th.hookMask&maskCall != 0 {
		mask += "c"
	}

	if th.hookMask&maskReturn != 0 {
		mask += "r"
	}

	if th.hookMask&maskLine != 0 {
		mask += "l"
	}

	return th.hookFunc, mask, th.hookCount
}

func (th *thread) SetHook(hook object.Value, mask string, count int) {
	mustFunctionOrNil(hook)

	if hook == nil {
		th.hookFunc = hook
		th.hookCount = 0
		th.hookMask = 0
	} else {
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
}

func (th *thread) LoadFunc(fn object.Value) {
	mustFunction(fn)

	th.loadfn(fn)
}

func (th *thread) Call(fn object.Value, args ...object.Value) ([]object.Value, *object.RuntimeError) {
	return th.docall(fn, args...)
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

	newth.pushContext(stackSize, false)

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

	th.pushContext(basicStackSize, false)

	go th.execute()

	return th
}
