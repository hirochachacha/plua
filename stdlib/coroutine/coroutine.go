package coroutine

import (
	"github.com/hirochachacha/plua/object"
	"github.com/hirochachacha/plua/object/fnutil"
	// "fmt"
)

func Create(th object.Thread, args ...object.Value) ([]object.Value, *object.RuntimeError) {
	ap := fnutil.NewArgParser(th, args)

	cl, err := ap.ToClosure(0)
	if err != nil {
		return nil, err
	}

	th1 := th.NewThread()

	th1.LoadFunc(cl)

	return []object.Value{th1}, nil
}

func IsYieldable(th object.Thread, args ...object.Value) ([]object.Value, *object.RuntimeError) {
	return []object.Value{object.Boolean(th.IsYieldable())}, nil
}

func Resume(th object.Thread, args ...object.Value) ([]object.Value, *object.RuntimeError) {
	ap := fnutil.NewArgParser(th, args)

	th1, err := ap.ToThread(0)
	if err != nil {
		return nil, err
	}

	rets, err := th1.Resume(args[1:]...)
	if err != nil {
		return []object.Value{object.False, err.Positioned()}, nil
	}

	return append([]object.Value{object.True}, rets...), nil
}

func Running(th object.Thread, args ...object.Value) ([]object.Value, *object.RuntimeError) {
	return []object.Value{th, object.Boolean(th.IsMainThread())}, nil
}

func Status(th object.Thread, args ...object.Value) ([]object.Value, *object.RuntimeError) {
	ap := fnutil.NewArgParser(th, args)

	th1, err := ap.ToThread(0)
	if err != nil {
		return nil, err
	}

	switch th1.Status() {
	case object.THREAD_INIT:
		return []object.Value{object.String("suspended")}, nil
	case object.THREAD_SUSPENDED:
		return []object.Value{object.String("suspended")}, nil
	case object.THREAD_ERROR:
		return []object.Value{object.String("dead")}, nil
	case object.THREAD_RETURN:
		return []object.Value{object.String("dead")}, nil
	case object.THREAD_RUNNING:
		if th == th1 {
			return []object.Value{object.String("running")}, nil
		}

		return []object.Value{object.String("normal")}, nil
	default:
		panic("unreachable")
	}
}

func Wrap(th object.Thread, args ...object.Value) ([]object.Value, *object.RuntimeError) {
	ap := fnutil.NewArgParser(th, args)

	cl, err := ap.ToClosure(0)
	if err != nil {
		return nil, err
	}

	th1 := th.NewThread()

	th1.LoadFunc(cl)

	fn := func(_ object.Thread, args1 ...object.Value) ([]object.Value, *object.RuntimeError) {
		rets, err := th1.Resume(args1...)
		if err != nil {
			return nil, err
		}

		return rets, nil
	}

	return []object.Value{object.GoFunction(fn)}, nil
}

func Yield(th object.Thread, args ...object.Value) ([]object.Value, *object.RuntimeError) {
	return th.Yield(args...)
}

func Open(th object.Thread, args ...object.Value) ([]object.Value, *object.RuntimeError) {
	m := th.NewTableSize(0, 7)

	m.Set(object.String("create"), object.GoFunction(Create))
	m.Set(object.String("isyieldable"), object.GoFunction(IsYieldable))
	m.Set(object.String("resume"), object.GoFunction(Resume))
	m.Set(object.String("running"), object.GoFunction(Running))
	m.Set(object.String("status"), object.GoFunction(Status))
	m.Set(object.String("wrap"), object.GoFunction(Wrap))
	m.Set(object.String("yield"), object.GoFunction(Yield))

	return []object.Value{m}, nil
}
