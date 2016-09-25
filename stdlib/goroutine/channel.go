package goroutine

import (
	"github.com/hirochachacha/plua/object"
	"github.com/hirochachacha/plua/object/fnutil"
)

func chanRecv(th object.Thread, args ...object.Value) ([]object.Value, *object.RuntimeError) {
	ap := fnutil.NewArgParser(th, args)

	ch, err := ap.ToChannel(0)
	if err != nil {
		return nil, err
	}

	val, ok := ch.Recv()

	return []object.Value{val, object.Boolean(ok)}, nil
}

func shanSend(th object.Thread, args ...object.Value) ([]object.Value, *object.RuntimeError) {
	ap := fnutil.NewArgParser(th, args)

	ch, err := ap.ToChannel(0)
	if err != nil {
		return nil, err
	}

	val, err := ap.ToValue(1)
	if err != nil {
		return nil, err
	}

	ch.Send(val)

	return nil, nil
}

func chanClose(th object.Thread, args ...object.Value) ([]object.Value, *object.RuntimeError) {
	ap := fnutil.NewArgParser(th, args)

	ch, err := ap.ToChannel(0)
	if err != nil {
		return nil, err
	}

	ch.Close()

	return nil, nil
}

func chanPairs(th object.Thread, args ...object.Value) ([]object.Value, *object.RuntimeError) {
	ap := fnutil.NewArgParser(th, args)

	ch, err := ap.ToChannel(0)
	if err != nil {
		return nil, err
	}

	return []object.Value{object.GoFunction(ChanNext), ch}, nil
}

func ChanNext(th object.Thread, args ...object.Value) ([]object.Value, *object.RuntimeError) {
	ap := fnutil.NewArgParser(th, args)

	ch, err := ap.ToChannel(0)
	if err != nil {
		return nil, err
	}

	val, ok := ch.Recv()
	if !ok {
		return nil, nil
	}

	return []object.Value{object.True, val}, nil
}
