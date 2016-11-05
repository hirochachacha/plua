package reflect

import (
	"reflect"

	"github.com/hirochachacha/plua/internal/tables"
	"github.com/hirochachacha/plua/object"
)

func buildStringMT() {
	mt := tables.NewTableSize(0, 7)

	mt.Set(object.TM_METATABLE, object.True)
	mt.Set(object.TM_TOSTRING, object.GoFunction(tostring))

	mt.Set(object.TM_INDEX, object.GoFunction(index))

	mt.Set(object.TM_EQ, cmp(func(x, y reflect.Value) bool { return x.String() == y.String() }))
	mt.Set(object.TM_LT, cmp(func(x, y reflect.Value) bool { return x.String() < y.String() }))
	mt.Set(object.TM_LE, cmp(func(x, y reflect.Value) bool { return x.String() <= y.String() }))

	mt.Set(object.TM_CONCAT, binary(func(x, y reflect.Value) (reflect.Value, *object.RuntimeError) {
		return reflect.ValueOf(x.String() + y.String()), nil
	}, mt))

	stringMT = mt
}
