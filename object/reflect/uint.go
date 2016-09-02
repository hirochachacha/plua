package reflect

import (
	"reflect"

	"github.com/hirochachacha/plua/internal/tables"
	"github.com/hirochachacha/plua/object"
)

// uintptr also supportted
func buildUintMT() {
	mt := tables.NewTableSize(0, 20)

	mt.Set(object.String("__metatable"), object.True)
	mt.Set(object.String("__tostring"), object.GoFunction(tostring))

	mt.Set(object.String("__index"), object.GoFunction(index))

	mt.Set(object.String("__eq"), cmp(func(x, y reflect.Value) bool { return x.Uint() == y.Uint() }))
	mt.Set(object.String("__lt"), cmp(func(x, y reflect.Value) bool { return x.Uint() < y.Uint() }))
	mt.Set(object.String("__le"), cmp(func(x, y reflect.Value) bool { return x.Uint() <= y.Uint() }))

	mt.Set(object.String("__unm"), unary(func(x reflect.Value) reflect.Value { return reflect.ValueOf(-x.Uint()) }, mt))
	mt.Set(object.String("__bnot"), unary(func(x reflect.Value) reflect.Value { return reflect.ValueOf(^x.Uint()) }, mt))

	mt.Set(object.String("__add"), binary(func(x, y reflect.Value) (reflect.Value, *object.RuntimeError) {
		return reflect.ValueOf(x.Uint() + y.Uint()), nil
	}, mt))
	mt.Set(object.String("__sub"), binary(func(x, y reflect.Value) (reflect.Value, *object.RuntimeError) {
		return reflect.ValueOf(x.Uint() - y.Uint()), nil
	}, mt))
	mt.Set(object.String("__mul"), binary(func(x, y reflect.Value) (reflect.Value, *object.RuntimeError) {
		return reflect.ValueOf(x.Uint() * y.Uint()), nil
	}, mt))
	mt.Set(object.String("__mod"), binary(func(x, y reflect.Value) (reflect.Value, *object.RuntimeError) {
		z, err := umod(x.Uint(), y.Uint())
		return reflect.ValueOf(z), err
	}, mt))
	mt.Set(object.String("__pow"), binary(func(x, y reflect.Value) (reflect.Value, *object.RuntimeError) {
		return reflect.ValueOf(upow(x.Uint(), y.Uint())), nil
	}, mt))
	mt.Set(object.String("__div"), binary(func(x, y reflect.Value) (reflect.Value, *object.RuntimeError) {
		z, err := udiv(x.Uint(), y.Uint())
		return reflect.ValueOf(z), err
	}, mt))
	mt.Set(object.String("__idiv"), binary(func(x, y reflect.Value) (reflect.Value, *object.RuntimeError) {
		z, err := udiv(x.Uint(), y.Uint())
		return reflect.ValueOf(z), err
	}, mt))
	mt.Set(object.String("__band"), binary(func(x, y reflect.Value) (reflect.Value, *object.RuntimeError) {
		return reflect.ValueOf(x.Uint() & y.Uint()), nil
	}, mt))
	mt.Set(object.String("__bor"), binary(func(x, y reflect.Value) (reflect.Value, *object.RuntimeError) {
		return reflect.ValueOf(x.Uint() | y.Uint()), nil
	}, mt))
	mt.Set(object.String("__bxor"), binary(func(x, y reflect.Value) (reflect.Value, *object.RuntimeError) {
		return reflect.ValueOf(x.Uint() ^ y.Uint()), nil
	}, mt))
	mt.Set(object.String("__shl"), binary(func(x, y reflect.Value) (reflect.Value, *object.RuntimeError) {
		return reflect.ValueOf(x.Uint() << y.Uint()), nil
	}, mt))
	mt.Set(object.String("__shr"), binary(func(x, y reflect.Value) (reflect.Value, *object.RuntimeError) {
		return reflect.ValueOf(x.Uint() >> y.Uint()), nil
	}, mt))

	uintMT = mt
}

func upow(x, y uint64) uint64 {
	prod := uint64(1)
	for y != 0 {
		if y&1 != 0 {
			prod *= x
		}
		y >>= 1
		x *= x
	}
	return prod
}

func umod(x, y uint64) (uint64, *object.RuntimeError) {
	if y == 0 {
		return 0, &object.RuntimeError{
			Value: object.String("integer divide by zero"),
		}
	}

	rem := x % y

	return rem, nil
}

func udiv(x, y uint64) (uint64, *object.RuntimeError) {
	if y == 0 {
		return 0, &object.RuntimeError{
			Value: object.String("integer divide by zero"),
		}
	}

	return x / y, nil
}
