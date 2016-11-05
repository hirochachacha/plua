package reflect

import (
	"fmt"
	"reflect"

	"github.com/hirochachacha/plua/internal/tables"
	"github.com/hirochachacha/plua/object"
	"github.com/hirochachacha/plua/object/fnutil"
)

func buildArrayMT() {
	mt := tables.NewTableSize(0, 7)

	mt.Set(object.TM_METATABLE, object.True)
	mt.Set(object.TM_TOSTRING, object.GoFunction(tostring))

	mt.Set(object.TM_INDEX, object.GoFunction(aindex))
	mt.Set(object.TM_NEWINDEX, object.GoFunction(anewindex))
	mt.Set(object.TM_LEN, object.GoFunction(length))
	mt.Set(object.TM_PAIRS, object.GoFunction(apairs))

	mt.Set(object.TM_EQ, cmp(func(x, y reflect.Value) bool { return aeq(x, y) }))

	arrayMT = mt
}

func buildSliceMT() {
	mt := tables.NewTableSize(0, 7)

	mt.Set(object.TM_METATABLE, object.True)
	mt.Set(object.TM_TOSTRING, object.GoFunction(tostring))

	mt.Set(object.TM_INDEX, object.GoFunction(aindex))
	mt.Set(object.TM_NEWINDEX, object.GoFunction(anewindex))
	mt.Set(object.TM_LEN, object.GoFunction(length))
	mt.Set(object.TM_PAIRS, object.GoFunction(apairs))

	mt.Set(object.TM_EQ, cmp(func(x, y reflect.Value) bool { return x.Pointer() == y.Pointer() }))

	sliceMT = mt
}

func length(th object.Thread, args ...object.Value) ([]object.Value, *object.RuntimeError) {
	ap := fnutil.NewArgParser(th, args)

	self, err := ap.ToFullUserdata(0)
	if err != nil {
		return nil, err
	}

	if self, ok := self.Value.(reflect.Value); ok {
		return []object.Value{object.Integer(self.Len())}, nil
	}

	return nil, errInvalidUserdata
}

func aeq(x, y reflect.Value) bool {
	xlen := x.Len()
	ylen := y.Len()

	if xlen == ylen {
		for i := 0; i < xlen; i++ {
			if x.Index(i) != y.Index(i) {
				return false
			}
		}
		return true
	}

	return false
}

func aindex(th object.Thread, args ...object.Value) ([]object.Value, *object.RuntimeError) {
	ap := fnutil.NewArgParser(th, args)

	self, err := ap.ToFullUserdata(0)
	if err != err {
		return nil, err
	}

	index, err := ap.ToGoInt(1)
	if err != nil {
		return nil, err
	}

	index--

	if self, ok := self.Value.(reflect.Value); ok {
		if 0 <= index && index < self.Len() {
			rval := self.Index(index)

			return []object.Value{valueOfReflect(rval, false)}, nil
		}

		return nil, object.NewRuntimeError(fmt.Sprintf("invalid array index %d (out of bounds for %d-element array)", index, self.Len()))
	}

	return nil, errInvalidUserdata
}

func anewindex(th object.Thread, args ...object.Value) ([]object.Value, *object.RuntimeError) {
	ap := fnutil.NewArgParser(th, args)

	self, err := ap.ToFullUserdata(0)
	if err != nil {
		return nil, err
	}

	index, err := ap.ToGoInt(1)
	if err != nil {
		return nil, err
	}

	val, err := ap.ToValue(2)
	if err != nil {
		return nil, err
	}

	index--

	if self, ok := self.Value.(reflect.Value); ok {
		if 0 <= index && index < self.Len() {
			styp := self.Type()
			vtyp := styp.Elem()

			if rval := toReflectValue(vtyp, val); rval.IsValid() {
				self.Index(index).Set(rval)

				return nil, nil
			}

			return nil, object.NewRuntimeError(fmt.Sprintf("non-%s array index %q", vtyp, reflect.TypeOf(val)))
		}

		return nil, object.NewRuntimeError(fmt.Sprintf("invalid array index %d (out of bounds for %d-element array)", index, self.Len()))
	}

	return nil, errInvalidUserdata
}

func apairs(th object.Thread, args ...object.Value) ([]object.Value, *object.RuntimeError) {
	ap := fnutil.NewArgParser(th, args)

	self, err := ap.ToFullUserdata(0)
	if err != nil {
		return nil, err
	}

	return []object.Value{object.GoFunction(anext), self, object.Integer(0)}, nil
}

func anext(th object.Thread, args ...object.Value) ([]object.Value, *object.RuntimeError) {
	ap := fnutil.NewArgParser(th, args)

	self, err := ap.ToFullUserdata(0)
	if err != nil {
		return nil, err
	}

	index, err := ap.ToGoInt(1)
	if err != nil {
		return nil, err
	}

	if self, ok := self.Value.(reflect.Value); ok {
		if index >= self.Len() {
			return nil, nil
		}

		rval := self.Index(index)

		index++

		return []object.Value{object.Integer(index), valueOfReflect(rval, false)}, nil
	}

	return nil, errInvalidUserdata
}
