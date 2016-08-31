package reflect

import (
	"reflect"

	"github.com/hirochachacha/plua/internal/tables"
	"github.com/hirochachacha/plua/object"
)

func buildComplexMT() {
	mt := tables.NewTableSize(0, 9)

	mt.Set(object.String("__metatable"), object.True)
	mt.Set(object.String("__tostring"), object.GoFunction(tostring))

	mt.Set(object.String("__index"), object.GoFunction(index))

	mt.Set(object.String("__eq"), cmp(func(x, y reflect.Value) bool { return x.Complex() == y.Complex() }))

	mt.Set(object.String("__unm"), unary(func(x reflect.Value) reflect.Value { return reflect.ValueOf(-x.Complex()) }, mt))

	mt.Set(object.String("__add"), binary(func(x, y reflect.Value) (reflect.Value, object.Value) {
		return reflect.ValueOf(x.Complex() + y.Complex()), nil
	}, mt))
	mt.Set(object.String("__sub"), binary(func(x, y reflect.Value) (reflect.Value, object.Value) {
		return reflect.ValueOf(x.Complex() - y.Complex()), nil
	}, mt))
	mt.Set(object.String("__mul"), binary(func(x, y reflect.Value) (reflect.Value, object.Value) {
		return reflect.ValueOf(x.Complex() * y.Complex()), nil
	}, mt))
	mt.Set(object.String("__div"), binary(func(x, y reflect.Value) (reflect.Value, object.Value) {
		return reflect.ValueOf(x.Complex() / y.Complex()), nil
	}, mt))

	complexMT = mt
}
