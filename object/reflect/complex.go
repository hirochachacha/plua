package reflect

import (
	"fmt"
	"reflect"

	"github.com/hirochachacha/plua/internal/tables"
	"github.com/hirochachacha/plua/object"
	"github.com/hirochachacha/plua/object/fnutil"
)

func buildComplexMT() {
	mt := tables.NewTableSize(0, 9)

	mt.Set(object.TM_METATABLE, object.True)
	mt.Set(object.TM_NAME, object.String("COMPLEX*"))
	mt.Set(object.TM_TOSTRING, tostring(toComplex))
	mt.Set(object.TM_INDEX, index(toComplex))

	mt.Set(object.TM_EQ, cmp(func(x, y reflect.Value) bool { return x.Complex() == y.Complex() }, toComplex))

	mt.Set(object.TM_UNM, unary(func(x reflect.Value) reflect.Value { return reflect.ValueOf(-x.Complex()) }, toComplex, mt))

	mt.Set(object.TM_ADD, binary(func(x, y reflect.Value) (reflect.Value, *object.RuntimeError) {
		return reflect.ValueOf(x.Complex() + y.Complex()), nil
	}, toComplex, mt))
	mt.Set(object.TM_SUB, binary(func(x, y reflect.Value) (reflect.Value, *object.RuntimeError) {
		return reflect.ValueOf(x.Complex() - y.Complex()), nil
	}, toComplex, mt))
	mt.Set(object.TM_MUL, binary(func(x, y reflect.Value) (reflect.Value, *object.RuntimeError) {
		return reflect.ValueOf(x.Complex() * y.Complex()), nil
	}, toComplex, mt))
	mt.Set(object.TM_DIV, binary(func(x, y reflect.Value) (reflect.Value, *object.RuntimeError) {
		return reflect.ValueOf(x.Complex() / y.Complex()), nil
	}, toComplex, mt))

	complexMT = mt
}

func toComplex(ap *fnutil.ArgParser, n int) (reflect.Value, *object.RuntimeError) {
	val, err := toValue(ap, n, "COMPLEX*")
	if err != nil {
		return reflect.Value{}, err
	}
	switch val.Kind() {
	case reflect.Complex64, reflect.Complex128:
	default:
		return reflect.Value{}, ap.TypeError(n, "COMPLEX*")
	}
	return val, nil
}

func cmtostring(th object.Thread, args ...object.Value) ([]object.Value, *object.RuntimeError) {
	ap := fnutil.NewArgParser(th, args)

	f, err := toComplex(ap, 0)
	if err != nil {
		return nil, err
	}

	return []object.Value{object.String(fmt.Sprintf("%v", f))}, nil
}
