package reflect

import (
	"reflect"

	"github.com/hirochachacha/plua/internal/tables"
	"github.com/hirochachacha/plua/object"
)

// uintptr also supportted
func buildUintMT() {
	mt := tables.NewTableSize(0, 20)

	mt.Set(object.TM_METATABLE, object.True)
	mt.Set(object.TM_TOSTRING, object.GoFunction(tostring))

	mt.Set(object.TM_INDEX, object.GoFunction(index))

	mt.Set(object.TM_EQ, cmp(func(x, y reflect.Value) bool { return x.Uint() == y.Uint() }))
	mt.Set(object.TM_LT, cmp(func(x, y reflect.Value) bool { return x.Uint() < y.Uint() }))
	mt.Set(object.TM_LE, cmp(func(x, y reflect.Value) bool { return x.Uint() <= y.Uint() }))

	mt.Set(object.TM_UNM, unary(func(x reflect.Value) reflect.Value { return reflect.ValueOf(-x.Uint()) }, mt))
	mt.Set(object.TM_BNOT, unary(func(x reflect.Value) reflect.Value { return reflect.ValueOf(^x.Uint()) }, mt))

	mt.Set(object.TM_ADD, binary(func(x, y reflect.Value) (reflect.Value, *object.RuntimeError) {
		return reflect.ValueOf(x.Uint() + y.Uint()), nil
	}, mt))
	mt.Set(object.TM_SUB, binary(func(x, y reflect.Value) (reflect.Value, *object.RuntimeError) {
		return reflect.ValueOf(x.Uint() - y.Uint()), nil
	}, mt))
	mt.Set(object.TM_MUL, binary(func(x, y reflect.Value) (reflect.Value, *object.RuntimeError) {
		return reflect.ValueOf(x.Uint() * y.Uint()), nil
	}, mt))
	mt.Set(object.TM_MOD, binary(func(x, y reflect.Value) (reflect.Value, *object.RuntimeError) {
		z, err := umod(x.Uint(), y.Uint())
		return reflect.ValueOf(z), err
	}, mt))
	mt.Set(object.TM_POW, binary(func(x, y reflect.Value) (reflect.Value, *object.RuntimeError) {
		return reflect.ValueOf(upow(x.Uint(), y.Uint())), nil
	}, mt))
	mt.Set(object.TM_DIV, binary(func(x, y reflect.Value) (reflect.Value, *object.RuntimeError) {
		z, err := udiv(x.Uint(), y.Uint())
		return reflect.ValueOf(z), err
	}, mt))
	mt.Set(object.TM_IDIV, binary(func(x, y reflect.Value) (reflect.Value, *object.RuntimeError) {
		z, err := udiv(x.Uint(), y.Uint())
		return reflect.ValueOf(z), err
	}, mt))
	mt.Set(object.TM_BAND, binary(func(x, y reflect.Value) (reflect.Value, *object.RuntimeError) {
		return reflect.ValueOf(x.Uint() & y.Uint()), nil
	}, mt))
	mt.Set(object.TM_BOR, binary(func(x, y reflect.Value) (reflect.Value, *object.RuntimeError) {
		return reflect.ValueOf(x.Uint() | y.Uint()), nil
	}, mt))
	mt.Set(object.TM_BXOR, binary(func(x, y reflect.Value) (reflect.Value, *object.RuntimeError) {
		return reflect.ValueOf(x.Uint() ^ y.Uint()), nil
	}, mt))
	mt.Set(object.TM_SHL, binary(func(x, y reflect.Value) (reflect.Value, *object.RuntimeError) {
		return reflect.ValueOf(x.Uint() << y.Uint()), nil
	}, mt))
	mt.Set(object.TM_SHR, binary(func(x, y reflect.Value) (reflect.Value, *object.RuntimeError) {
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
