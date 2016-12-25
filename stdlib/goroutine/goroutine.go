package goroutine

import (
	goreflect "reflect"

	"github.com/hirochachacha/plua/object"
	"github.com/hirochachacha/plua/object/fnutil"
	"github.com/hirochachacha/plua/object/reflect"
)

func newchannel(th object.Thread, args ...object.Value) ([]object.Value, *object.RuntimeError) {
	ap := fnutil.NewArgParser(th, args)

	cap, err := ap.OptGoInt(0, 0)
	if err != nil {
		return nil, err
	}

	if cap < 0 {
		return nil, ap.ArgError(0, "capacity should not be negative")
	}

	return []object.Value{reflect.ValueOf(make(chan object.Value, cap))}, nil
}

func _select(th object.Thread, args ...object.Value) ([]object.Value, *object.RuntimeError) {
	ap := fnutil.NewArgParser(th, args)

	cases := make([]goreflect.SelectCase, len(args))
	for i := range args {
		ud, err := ap.ToFullUserdata(i)
		if err != nil {
			return nil, ap.TypeError(i, "SELECT_CASE*")
		}

		c, ok := ud.Value.(goreflect.SelectCase)
		if !ok {
			return nil, ap.TypeError(i, "SELECT_CASE*")
		}

		cases[i] = c
	}

	chosen, recv, recvOK := goreflect.Select(cases)

	return []object.Value{object.Integer(chosen + 1), reflect.ValueOf(recv), object.Boolean(recvOK)}, nil
}

func _case(th object.Thread, args ...object.Value) ([]object.Value, *object.RuntimeError) {
	ap := fnutil.NewArgParser(th, args)

	dir, err := ap.ToGoString(0)
	if err != nil {
		return nil, err
	}

	var rdir goreflect.SelectDir

	switch dir {
	case "send":
		rdir = goreflect.SelectSend
	case "recv":
		rdir = goreflect.SelectRecv
	case "default":
		rdir = goreflect.SelectDefault
	}

	ud, err := ap.ToFullUserdata(1)
	if err != nil {
		return nil, ap.TypeError(1, "CHAN*")
	}

	if ud.Metatable == nil || ud.Metatable.Get(object.TM_NAME) != object.String("CHAN*") {
		return nil, ap.TypeError(1, "CHAN*")
	}

	rch, ok := ud.Value.(goreflect.Value)
	if !ok {
		return nil, ap.TypeError(1, "CHAN*")
	}

	var rsend goreflect.Value
	if send, ok := ap.Get(2); ok {
		rsend = goreflect.ValueOf(send)
	}

	c := goreflect.SelectCase{
		Dir:  rdir,
		Chan: rch,
		Send: rsend,
	}

	ud = &object.Userdata{
		Value: c,
	}

	return []object.Value{ud}, nil
}

func wrap(th object.Thread, args ...object.Value) ([]object.Value, *object.RuntimeError) {
	ap := fnutil.NewArgParser(th, args)

	fn, err := ap.ToFunction(0)
	if err != nil {
		return nil, err
	}

	th1 := th.NewGoThread()

	th1.LoadFunc(fn)

	retfn := func(_ object.Thread, args1 ...object.Value) ([]object.Value, *object.RuntimeError) {
		rets, err := th1.Resume(args1...)
		if err != nil {
			return nil, err
		}

		return rets, nil
	}

	return []object.Value{object.GoFunction(retfn)}, nil
}

func Open(th object.Thread, args ...object.Value) ([]object.Value, *object.RuntimeError) {
	m := th.NewTableSize(0, 4)

	m.Set(object.String("newchannel"), object.GoFunction(newchannel))
	m.Set(object.String("select"), object.GoFunction(_select))
	m.Set(object.String("case"), object.GoFunction(_case))
	m.Set(object.String("wrap"), object.GoFunction(wrap))

	return []object.Value{m}, nil
}
