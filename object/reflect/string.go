package reflect

import (
	"reflect"

	"github.com/hirochachacha/plua/internal/tables"
	"github.com/hirochachacha/plua/object"
	"github.com/hirochachacha/plua/object/fnutil"
)

func buildStringMT() {
	mt := tables.NewTableSize(0, 7)

	mt.Set(object.TM_METATABLE, object.True)
	mt.Set(object.TM_NAME, object.String("STRING*"))
	mt.Set(object.TM_TOSTRING, tostring(toString))
	mt.Set(object.TM_INDEX, index(toString))

	mt.Set(object.TM_EQ, cmp(func(x, y reflect.Value) bool { return x.String() == y.String() }, toString))
	mt.Set(object.TM_LT, cmp(func(x, y reflect.Value) bool { return x.String() < y.String() }, toString))
	mt.Set(object.TM_LE, cmp(func(x, y reflect.Value) bool { return x.String() <= y.String() }, toString))

	mt.Set(object.TM_CONCAT, binary(func(x, y reflect.Value) (reflect.Value, *object.RuntimeError) {
		return reflect.ValueOf(x.String() + y.String()), nil
	}, toString, mt))

	stringMT = mt
}

func toString(ap *fnutil.ArgParser, n int) (reflect.Value, *object.RuntimeError) {
	val, err := toValue(ap, n, "STRING*")
	if err != nil {
		return reflect.Value{}, err
	}
	if val.Kind() != reflect.String {
		return reflect.Value{}, ap.TypeError(n, "STRING*")
	}
	return val, nil
}
