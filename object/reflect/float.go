package reflect

import (
	"math"
	"reflect"

	"github.com/hirochachacha/plua/internal/tables"
	"github.com/hirochachacha/plua/object"
)

func buildFloatMT() {
	mt := tables.NewTableSize(0, 14)

	mt.Set(object.String("__metatable"), object.True)
	mt.Set(object.String("__tostring"), object.GoFunction(tostring))

	mt.Set(object.String("__index"), object.GoFunction(index))

	mt.Set(object.String("__eq"), cmp(func(x, y reflect.Value) bool { return x.Float() == y.Float() }))
	mt.Set(object.String("__lt"), cmp(func(x, y reflect.Value) bool { return x.Float() < y.Float() }))
	mt.Set(object.String("__le"), cmp(func(x, y reflect.Value) bool { return x.Float() <= y.Float() }))

	mt.Set(object.String("__unm"), unary(func(x reflect.Value) reflect.Value { return reflect.ValueOf(-x.Float()) }, mt))

	mt.Set(object.String("__add"), binary(func(x, y reflect.Value) (reflect.Value, *object.RuntimeError) {
		return reflect.ValueOf(x.Float() + y.Float()), nil
	}, mt))
	mt.Set(object.String("__sub"), binary(func(x, y reflect.Value) (reflect.Value, *object.RuntimeError) {
		return reflect.ValueOf(x.Float() - y.Float()), nil
	}, mt))
	mt.Set(object.String("__mul"), binary(func(x, y reflect.Value) (reflect.Value, *object.RuntimeError) {
		return reflect.ValueOf(x.Float() * y.Float()), nil
	}, mt))
	mt.Set(object.String("__mod"), binary(func(x, y reflect.Value) (reflect.Value, *object.RuntimeError) {
		return reflect.ValueOf(fmod(x.Float(), y.Float())), nil
	}, mt))
	mt.Set(object.String("__pow"), binary(func(x, y reflect.Value) (reflect.Value, *object.RuntimeError) {
		return reflect.ValueOf(math.Pow(x.Float(), y.Float())), nil
	}, mt))
	mt.Set(object.String("__div"), binary(func(x, y reflect.Value) (reflect.Value, *object.RuntimeError) {
		return reflect.ValueOf(x.Float() / y.Float()), nil
	}, mt))
	mt.Set(object.String("__idiv"), binary(func(x, y reflect.Value) (reflect.Value, *object.RuntimeError) {
		return reflect.ValueOf(fidiv(x.Float(), y.Float())), nil
	}, mt))

	floatMT = mt
}

func fmod(x, y float64) float64 {
	rem := math.Mod(x, y)

	if rem < 0 {
		rem += y
	}

	return rem
}

func fidiv(x, y float64) float64 {
	f, _ := math.Modf(x / y)

	return f
}
