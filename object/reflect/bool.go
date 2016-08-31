package reflect

import (
	"reflect"

	"github.com/hirochachacha/plua/internal/tables"
	"github.com/hirochachacha/plua/object"
)

func buildBoolMT() {
	mt := tables.NewTableSize(0, 4)

	mt.Set(object.String("__metatable"), object.True)
	mt.Set(object.String("__tostring"), object.GoFunction(tostring))

	mt.Set(object.String("__index"), object.GoFunction(index))

	mt.Set(object.String("__eq"), cmp(func(x, y reflect.Value) bool { return x.Bool() == y.Bool() }))

	boolMT = mt
}
