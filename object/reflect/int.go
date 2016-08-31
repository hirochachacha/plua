package reflect

import (
	"reflect"

	"github.com/hirochachacha/plua/internal/limits"
	"github.com/hirochachacha/plua/internal/tables"
	"github.com/hirochachacha/plua/object"
)

func buildIntMT() {
	mt := tables.NewTableSize(0, 20)

	mt.Set(object.String("__metatable"), object.True)
	mt.Set(object.String("__tostring"), object.GoFunction(tostring))

	mt.Set(object.String("__index"), object.GoFunction(index))

	mt.Set(object.String("__eq"), cmp(func(x, y reflect.Value) bool { return x.Int() == y.Int() }))
	mt.Set(object.String("__lt"), cmp(func(x, y reflect.Value) bool { return x.Int() < y.Int() }))
	mt.Set(object.String("__le"), cmp(func(x, y reflect.Value) bool { return x.Int() <= y.Int() }))

	mt.Set(object.String("__unm"), unary(func(x reflect.Value) reflect.Value { return reflect.ValueOf(-x.Int()) }, mt))
	mt.Set(object.String("__bnot"), unary(func(x reflect.Value) reflect.Value { return reflect.ValueOf(^x.Int()) }, mt))

	mt.Set(object.String("__add"), binary(func(x, y reflect.Value) (reflect.Value, object.Value) {
		return reflect.ValueOf(x.Int() + y.Int()), nil
	}, mt))
	mt.Set(object.String("__sub"), binary(func(x, y reflect.Value) (reflect.Value, object.Value) {
		return reflect.ValueOf(x.Int() - y.Int()), nil
	}, mt))
	mt.Set(object.String("__mul"), binary(func(x, y reflect.Value) (reflect.Value, object.Value) {
		return reflect.ValueOf(x.Int() * y.Int()), nil
	}, mt))
	mt.Set(object.String("__mod"), binary(func(x, y reflect.Value) (reflect.Value, object.Value) {
		z, err := imod(x.Int(), y.Int())
		return reflect.ValueOf(z), err
	}, mt))
	mt.Set(object.String("__pow"), binary(func(x, y reflect.Value) (reflect.Value, object.Value) {
		return reflect.ValueOf(ipow(x.Int(), y.Int())), nil
	}, mt))
	mt.Set(object.String("__div"), binary(func(x, y reflect.Value) (reflect.Value, object.Value) {
		z, err := idiv(x.Int(), y.Int())
		return reflect.ValueOf(z), err
	}, mt))
	mt.Set(object.String("__idiv"), binary(func(x, y reflect.Value) (reflect.Value, object.Value) {
		z, err := idiv(x.Int(), y.Int())
		return reflect.ValueOf(z), err
	}, mt))
	mt.Set(object.String("__band"), binary(func(x, y reflect.Value) (reflect.Value, object.Value) {
		return reflect.ValueOf(x.Int() & y.Int()), nil
	}, mt))
	mt.Set(object.String("__bor"), binary(func(x, y reflect.Value) (reflect.Value, object.Value) {
		return reflect.ValueOf(x.Int() | y.Int()), nil
	}, mt))
	mt.Set(object.String("__bxor"), binary(func(x, y reflect.Value) (reflect.Value, object.Value) {
		return reflect.ValueOf(x.Int() ^ y.Int()), nil
	}, mt))
	mt.Set(object.String("__shl"), binary(func(x, y reflect.Value) (reflect.Value, object.Value) {
		return reflect.ValueOf(ishl(x.Int(), y.Int())), nil
	}, mt))
	mt.Set(object.String("__shr"), binary(func(x, y reflect.Value) (reflect.Value, object.Value) {
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

func imod(x, y int64) (int64, object.Value) {
	if y == 0 {
		return 0, object.String("integer divide by zero")
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

func idiv(x, y int64) (int64, object.Value) {
	if y == 0 {
		return 0, object.String("integer divide by zero")
	}

	return x / y, nil
}
