package reflect

import (
	"math"
	"reflect"

	"github.com/hirochachacha/plua/internal/tables"
	"github.com/hirochachacha/plua/object"
	"github.com/hirochachacha/plua/object/fnutil"
)

func buildFloatMT() {
	mt := tables.NewTableSize(0, 14)

	mt.Set(object.TM_METATABLE, object.True)
	mt.Set(object.TM_NAME, object.String("FLOAT*"))
	mt.Set(object.TM_TOSTRING, tostring(toFloat))
	mt.Set(object.TM_INDEX, index(toFloat))

	mt.Set(object.TM_EQ, cmp(func(x, y reflect.Value) bool { return x.Float() == y.Float() }, toFloat))
	mt.Set(object.TM_LT, cmp(func(x, y reflect.Value) bool { return x.Float() < y.Float() }, toFloat))
	mt.Set(object.TM_LE, cmp(func(x, y reflect.Value) bool { return x.Float() <= y.Float() }, toFloat))

	mt.Set(object.TM_UNM, unary(func(x reflect.Value) reflect.Value { return reflect.ValueOf(-x.Float()) }, toFloat, mt))

	mt.Set(object.TM_ADD, binary(func(x, y reflect.Value) (reflect.Value, *object.RuntimeError) {
		return reflect.ValueOf(x.Float() + y.Float()), nil
	}, toFloat, mt))
	mt.Set(object.TM_SUB, binary(func(x, y reflect.Value) (reflect.Value, *object.RuntimeError) {
		return reflect.ValueOf(x.Float() - y.Float()), nil
	}, toFloat, mt))
	mt.Set(object.TM_MUL, binary(func(x, y reflect.Value) (reflect.Value, *object.RuntimeError) {
		return reflect.ValueOf(x.Float() * y.Float()), nil
	}, toFloat, mt))
	mt.Set(object.TM_MOD, binary(func(x, y reflect.Value) (reflect.Value, *object.RuntimeError) {
		return reflect.ValueOf(fmod(x.Float(), y.Float())), nil
	}, toFloat, mt))
	mt.Set(object.TM_POW, binary(func(x, y reflect.Value) (reflect.Value, *object.RuntimeError) {
		return reflect.ValueOf(math.Pow(x.Float(), y.Float())), nil
	}, toFloat, mt))
	mt.Set(object.TM_DIV, binary(func(x, y reflect.Value) (reflect.Value, *object.RuntimeError) {
		return reflect.ValueOf(x.Float() / y.Float()), nil
	}, toFloat, mt))
	mt.Set(object.TM_IDIV, binary(func(x, y reflect.Value) (reflect.Value, *object.RuntimeError) {
		return reflect.ValueOf(fidiv(x.Float(), y.Float())), nil
	}, toFloat, mt))

	floatMT = mt
}

func toFloat(ap *fnutil.ArgParser, n int) (reflect.Value, *object.RuntimeError) {
	if f, err := ap.ToGoFloat64(n); err == nil {
		return reflect.ValueOf(f), nil
	}
	val, err := toValue(ap, n, "FLOAT*")
	if err != nil {
		return reflect.Value{}, err
	}
	switch val.Kind() {
	case reflect.Float32, reflect.Float64:
	default:
		return reflect.Value{}, ap.TypeError(n, "FLOAT*")
	}
	return val, nil
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
