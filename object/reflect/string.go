package reflect

import (
	"reflect"

	"github.com/hirochachacha/plua/internal/tables"
	"github.com/hirochachacha/plua/object"
)

func buildStringMT() {
	mt := tables.NewTableSize(0, 7)

	mt.Set(object.String("__metatable"), object.True)
	mt.Set(object.String("__tostring"), object.GoFunction(tostring))

	mt.Set(object.String("__index"), object.GoFunction(index))

	mt.Set(object.String("__eq"), cmp(func(x, y reflect.Value) bool { return x.String() == y.String() }))
	mt.Set(object.String("__lt"), cmp(func(x, y reflect.Value) bool { return x.String() < y.String() }))
	mt.Set(object.String("__le"), cmp(func(x, y reflect.Value) bool { return x.String() <= y.String() }))

	mt.Set(object.String("__concat"), binary(func(x, y reflect.Value) (reflect.Value, *object.RuntimeError) {
		return reflect.ValueOf(x.String() + y.String()), nil
	}, mt))

	stringMT = mt
}
