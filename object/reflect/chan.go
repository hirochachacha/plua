package reflect

import (
	"fmt"
	"reflect"

	"github.com/hirochachacha/plua/internal/tables"
	"github.com/hirochachacha/plua/object"
	"github.com/hirochachacha/plua/object/fnutil"
)

func buildChanMT() {
	mt := tables.NewTableSize(0, 6)

	mt.Set(object.TM_METATABLE, object.True)
	mt.Set(object.TM_NAME, object.String("CHAN*"))
	mt.Set(object.TM_TOSTRING, object.GoFunction(ctostring))
	mt.Set(object.TM_INDEX, object.GoFunction(cindex))

	mt.Set(object.TM_LEN, object.GoFunction(clength))
	mt.Set(object.TM_PAIRS, object.GoFunction(cpairs))

	mt.Set(object.TM_EQ, cmp(func(x, y reflect.Value) bool { return x.Pointer() == y.Pointer() }, toChan))

	chanMT = mt
}

func toChan(ap *fnutil.ArgParser, n int) (reflect.Value, *object.RuntimeError) {
	val, err := toValue(ap, n, "CHAN*")
	if err != nil {
		return reflect.Value{}, err
	}
	if val.Kind() != reflect.Chan {
		return reflect.Value{}, ap.TypeError(n, "CHAN*")
	}
	return val, nil
}

func ctostring(th object.Thread, args ...object.Value) ([]object.Value, *object.RuntimeError) {
	ap := fnutil.NewArgParser(th, args)

	ch, err := toChan(ap, 0)
	if err != nil {
		return nil, err
	}

	return []object.Value{object.String(fmt.Sprintf("go channel (0x%x)", ch.Pointer()))}, nil
}

func cindex(th object.Thread, args ...object.Value) ([]object.Value, *object.RuntimeError) {
	ap := fnutil.NewArgParser(th, args)

	ch, err := toChan(ap, 0)
	if err != nil {
		return nil, err
	}

	name, err := ap.ToGoString(1)
	if err != nil {
		return nil, err
	}

	if !isPublic(name) {
		switch name {
		case "send":
			return []object.Value{object.GoFunction(csend)}, nil
		case "recv":
			return []object.Value{object.GoFunction(crecv)}, nil
		case "close":
			return []object.Value{object.GoFunction(cclose)}, nil
		}

		return nil, nil
	}

	method := ch.MethodByName(name)

	if !method.IsValid() {
		if ch.CanAddr() {
			method = ch.Addr().MethodByName(name)
		} else {
			self2 := reflect.New(ch.Type())
			self2.Elem().Set(ch)
			method = self2.MethodByName(name)
		}

		if !method.IsValid() {
			return nil, nil
		}
	}

	return []object.Value{valueOfReflect(method, false)}, nil
}

func clength(th object.Thread, args ...object.Value) ([]object.Value, *object.RuntimeError) {
	ap := fnutil.NewArgParser(th, args)

	ch, err := toChan(ap, 0)
	if err != nil {
		return nil, err
	}

	return []object.Value{object.Integer(ch.Len())}, nil
}

func csend(th object.Thread, args ...object.Value) ([]object.Value, *object.RuntimeError) {
	ap := fnutil.NewArgParser(th, args)

	ch, err := toChan(ap, 0)
	if err != nil {
		return nil, err
	}

	x, err := ap.ToValue(1)
	if err != nil {
		return nil, err
	}

	styp := ch.Type()
	vtyp := styp.Elem()

	if x := toReflectValue(vtyp, x); x.IsValid() {
		ch.Send(x)

		return nil, nil
	}

	return nil, object.NewRuntimeError(fmt.Sprintf("cannot use %v (type %s) as type %s in send", x, reflect.TypeOf(x), vtyp))
}

func crecv(th object.Thread, args ...object.Value) ([]object.Value, *object.RuntimeError) {
	ap := fnutil.NewArgParser(th, args)

	ch, err := toChan(ap, 0)
	if err != nil {
		return nil, err
	}

	rval, ok := ch.Recv()
	if !ok {
		return []object.Value{nil, object.False}, nil
	}

	return []object.Value{valueOfReflect(rval, false), object.True}, nil
}

func cclose(th object.Thread, args ...object.Value) ([]object.Value, *object.RuntimeError) {
	ap := fnutil.NewArgParser(th, args)

	ch, err := toChan(ap, 0)
	if err != nil {
		return nil, err
	}

	ch.Close()

	return nil, nil
}

func cpairs(th object.Thread, args ...object.Value) ([]object.Value, *object.RuntimeError) {
	ap := fnutil.NewArgParser(th, args)

	_, err := toChan(ap, 0)
	if err != nil {
		return nil, err
	}

	return []object.Value{object.GoFunction(cnext), args[0]}, nil
}

func cnext(th object.Thread, args ...object.Value) ([]object.Value, *object.RuntimeError) {
	ap := fnutil.NewArgParser(th, args)

	ch, err := toChan(ap, 0)
	if err != nil {
		return nil, err
	}

	rval, ok := ch.Recv()
	if !ok {
		return nil, nil
	}

	return []object.Value{nil, valueOfReflect(rval, false)}, nil
}
