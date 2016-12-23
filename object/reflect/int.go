package reflect

import (
	"reflect"

	"github.com/hirochachacha/plua/internal/limits"
	"github.com/hirochachacha/plua/internal/tables"
	"github.com/hirochachacha/plua/object"
)

func buildIntMT() {
	mt := tables.NewTableSize(0, 20)

	mt.Set(object.TM_METATABLE, object.True)
	mt.Set(object.TM_TOSTRING, object.GoFunction(tostring))

	mt.Set(object.TM_INDEX, object.GoFunction(index))

	mt.Set(object.TM_EQ, cmp(func(x, y reflect.Value) bool { return x.Int() == y.Int() }))
	mt.Set(object.TM_LT, cmp(func(x, y reflect.Value) bool { return x.Int() < y.Int() }))
	mt.Set(object.TM_LE, cmp(func(x, y reflect.Value) bool { return x.Int() <= y.Int() }))

	mt.Set(object.TM_UNM, unary(func(x reflect.Value) reflect.Value { return reflect.ValueOf(-x.Int()) }, mt))
	mt.Set(object.TM_BNOT, unary(func(x reflect.Value) reflect.Value { return reflect.ValueOf(^x.Int()) }, mt))

	mt.Set(object.TM_ADD, binary(func(x, y reflect.Value) (reflect.Value, *object.RuntimeError) {
		return reflect.ValueOf(x.Int() + y.Int()), nil
	}, mt))
	mt.Set(object.TM_SUB, binary(func(x, y reflect.Value) (reflect.Value, *object.RuntimeError) {
		return reflect.ValueOf(x.Int() - y.Int()), nil
	}, mt))
	mt.Set(object.TM_MUL, binary(func(x, y reflect.Value) (reflect.Value, *object.RuntimeError) {
		return reflect.ValueOf(x.Int() * y.Int()), nil
	}, mt))
	mt.Set(object.TM_MOD, binary(func(x, y reflect.Value) (reflect.Value, *object.RuntimeError) {
		z, err := imod(x.Int(), y.Int())
		return reflect.ValueOf(z), err
	}, mt))
	mt.Set(object.TM_POW, binary(func(x, y reflect.Value) (reflect.Value, *object.RuntimeError) {
		return reflect.ValueOf(ipow(x.Int(), y.Int())), nil
	}, mt))
	mt.Set(object.TM_DIV, binary(func(x, y reflect.Value) (reflect.Value, *object.RuntimeError) {
		z, err := idiv(x.Int(), y.Int())
		return reflect.ValueOf(z), err
	}, mt))
	mt.Set(object.TM_IDIV, binary(func(x, y reflect.Value) (reflect.Value, *object.RuntimeError) {
		z, err := idiv(x.Int(), y.Int())
		return reflect.ValueOf(z), err
	}, mt))
	mt.Set(object.TM_BAND, binary(func(x, y reflect.Value) (reflect.Value, *object.RuntimeError) {
		return reflect.ValueOf(x.Int() & y.Int()), nil
	}, mt))
	mt.Set(object.TM_BOR, binary(func(x, y reflect.Value) (reflect.Value, *object.RuntimeError) {
		return reflect.ValueOf(x.Int() | y.Int()), nil
	}, mt))
	mt.Set(object.TM_BXOR, binary(func(x, y reflect.Value) (reflect.Value, *object.RuntimeError) {
		return reflect.ValueOf(x.Int() ^ y.Int()), nil
	}, mt))
	mt.Set(object.TM_SHL, binary(func(x, y reflect.Value) (reflect.Value, *object.RuntimeError) {
		return reflect.ValueOf(ishl(x.Int(), y.Int())), nil
	}, mt))
	mt.Set(object.TM_SHR, binary(func(x, y reflect.Value) (reflect.Value, *object.RuntimeError) {
		return reflect.ValueOf(ishr(x.Int(), y.Int())), nil
	}, mt))

	intMT = mt
}

func ipow(x, y int64) int64 {
	prod := int64(1)
	for y != 0 {
		if y&1 != 0 {
			prod *= x
		}
		y >>= 1
		x *= x
	}
	return prod
}

func ishl(x, y int64) int64 {
	if y > 0 {
		return x << uint64(y)
	}
	return x >> uint64(-y)
}

func ishr(x, y int64) int64 {
	if y > 0 {
		return x >> uint64(y)
	}
	return x << uint64(-y)
}

func imod(x, y int64) (int64, *object.RuntimeError) {
	if y == 0 {
		return 0, &object.RuntimeError{
			RawValue: object.String("integer divide by zero"),
		}
	}

	if x == limits.MinInt64 && y == -1 {
		return 0, nil
	}

	rem := x % y

	if rem < 0 {
		rem += y
	}

	return rem, nil
}

func idiv(x, y int64) (int64, *object.RuntimeError) {
	if y == 0 {
		return 0, &object.RuntimeError{
			RawValue: object.String("integer divide by zero"),
		}
	}

	return x / y, nil
}
