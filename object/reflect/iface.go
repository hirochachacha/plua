package reflect

import (
	"fmt"
	"reflect"

	"github.com/hirochachacha/plua/internal/tables"
	"github.com/hirochachacha/plua/object"
	"github.com/hirochachacha/plua/object/fnutil"
)

func buildIfaceMT() {
	mt := tables.NewTableSize(0, 4)

	mt.Set(object.TM_METATABLE, object.True)
	mt.Set(object.TM_NAME, object.String("IFACE*"))
	mt.Set(object.TM_TOSTRING, object.GoFunction(itostring))

	mt.Set(object.TM_INDEX, object.GoFunction(iindex))

	mt.Set(object.TM_EQ, cmp(func(x, y reflect.Value) bool { return x.Interface() == y.Interface() }, toIface))

	ifaceMT = mt
}

func toIface(ap *fnutil.ArgParser, n int) (reflect.Value, *object.RuntimeError) {
	val, err := toValue(ap, n, "IFACE*")
	if err != nil {
		return reflect.Value{}, err
	}
	if val.Kind() != reflect.Interface {
		return reflect.Value{}, ap.TypeError(n, "IFACE*")
	}
	return val, nil
}

func itostring(th object.Thread, args ...object.Value) ([]object.Value, *object.RuntimeError) {
	ap := fnutil.NewArgParser(th, args)

	i, err := toIface(ap, 0)
	if err != nil {
		return nil, err
	}

	return []object.Value{object.String(fmt.Sprintf("go interface (0x%x)", i.Pointer()))}, nil
}

func iindex(th object.Thread, args ...object.Value) ([]object.Value, *object.RuntimeError) {
	ap := fnutil.NewArgParser(th, args)

	i, err := toIface(ap, 0)
	if err != nil {
		return nil, err
	}

	name, err := ap.ToGoString(1)
	if err != nil {
		return nil, err
	}

	if !isPublic(name) {
		return nil, nil
	}

	method := i.MethodByName(name)

	if method.IsValid() {
		return []object.Value{valueOfReflect(method, false)}, nil
	}

	return nil, nil
}
