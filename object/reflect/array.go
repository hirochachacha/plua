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
	mt.Set(object.TM_NAME, object.String("ARRAY*"))
	mt.Set(object.TM_TOSTRING, object.GoFunction(atostring))

	mt.Set(object.TM_INDEX, object.GoFunction(aindex))
	mt.Set(object.TM_LEN, object.GoFunction(alength))
	mt.Set(object.TM_PAIRS, object.GoFunction(apairs))

	mt.Set(object.TM_EQ, cmp(func(x, y reflect.Value) bool { return aeq(x, y) }, toArray))

	arrayMT = mt
}

func toArray(ap *fnutil.ArgParser, n int) (reflect.Value, *object.RuntimeError) {
	val, err := toValue(ap, n, "ARRAY*")
	if err != nil {
		return reflect.Value{}, err
	}
	if val.Kind() != reflect.Array {
		return reflect.Value{}, ap.TypeError(n, "ARRAY*")
	}
	return val, nil
}

func atostring(th object.Thread, args ...object.Value) ([]object.Value, *object.RuntimeError) {
	ap := fnutil.NewArgParser(th, args)

	a, err := toArray(ap, 0)
	if err != nil {
		return nil, err
	}

	return []object.Value{object.String(fmt.Sprintf("go array[%d]", a.Len()))}, nil
}

func alength(th object.Thread, args ...object.Value) ([]object.Value, *object.RuntimeError) {
	ap := fnutil.NewArgParser(th, args)

	a, err := toArray(ap, 0)
	if err != nil {
		return nil, err
	}

	return []object.Value{object.Integer(a.Len())}, nil
}

func aeq(x, y reflect.Value) bool {
	xlen := x.Len()
	ylen := y.Len()

	if xlen == ylen {
		for i := 0; i < xlen; i++ {
			if x.Index(i).Interface() != y.Index(i).Interface() {
				return false
			}
		}
		return true
	}

	return false
}

func aindex(th object.Thread, args ...object.Value) ([]object.Value, *object.RuntimeError) {
	ap := fnutil.NewArgParser(th, args)

	a, err := toArray(ap, 0)
	if err != err {
		return nil, err
	}

	index, err := ap.ToGoInt(1)
	if err != nil {
		name, err := ap.ToGoString(1)
		if err != nil {
			return nil, err
		}

		if !isPublic(name) {
			return nil, nil
		}

		method := a.MethodByName(name)

		if !method.IsValid() {
			if a.CanAddr() {
				method = a.Addr().MethodByName(name)
			} else {
				self2 := reflect.New(a.Type())
				self2.Elem().Set(a)
				method = self2.MethodByName(name)
			}

			if !method.IsValid() {
				return nil, nil
			}
		}

		return []object.Value{valueOfReflect(method, false)}, nil
	}

	index--

	if 0 <= index && index < a.Len() {
		rval := a.Index(index)

		return []object.Value{valueOfReflect(rval, false)}, nil
	}

	return nil, nil
}

func apairs(th object.Thread, args ...object.Value) ([]object.Value, *object.RuntimeError) {
	ap := fnutil.NewArgParser(th, args)

	_, err := toArray(ap, 0)
	if err != nil {
		return nil, err
	}

	return []object.Value{object.GoFunction(anext), args[0], object.Integer(0)}, nil
}

func anext(th object.Thread, args ...object.Value) ([]object.Value, *object.RuntimeError) {
	ap := fnutil.NewArgParser(th, args)

	a, err := toArray(ap, 0)
	if err != nil {
		return nil, err
	}

	index, err := ap.ToGoInt(1)
	if err != nil {
		return nil, err
	}

	if index >= a.Len() {
		return nil, nil
	}

	rval := a.Index(index)

	index++

	return []object.Value{object.Integer(index), valueOfReflect(rval, false)}, nil
}

func buildSliceMT() {
	mt := tables.NewTableSize(0, 7)

	mt.Set(object.TM_METATABLE, object.True)
	mt.Set(object.TM_NAME, object.String("SLICE*"))
	mt.Set(object.TM_TOSTRING, object.GoFunction(stostring))

	mt.Set(object.TM_INDEX, object.GoFunction(sindex))
	mt.Set(object.TM_NEWINDEX, object.GoFunction(snewindex))
	mt.Set(object.TM_LEN, object.GoFunction(slength))
	mt.Set(object.TM_PAIRS, object.GoFunction(spairs))

	mt.Set(object.TM_EQ, cmp(func(x, y reflect.Value) bool { return x.Pointer() == y.Pointer() }, toSlice))

	sliceMT = mt
}

func toSlice(ap *fnutil.ArgParser, n int) (reflect.Value, *object.RuntimeError) {
	val, err := toValue(ap, n, "SLICE*")
	if err != nil {
		return reflect.Value{}, err
	}
	if val.Kind() != reflect.Slice {
		return reflect.Value{}, ap.TypeError(n, "SLICE*")
	}
	return val, nil
}

func stostring(th object.Thread, args ...object.Value) ([]object.Value, *object.RuntimeError) {
	ap := fnutil.NewArgParser(th, args)

	s, err := toSlice(ap, 0)
	if err != nil {
		return nil, err
	}

	return []object.Value{object.String(fmt.Sprintf("go slice (0x%x)", s.Pointer()))}, nil
}

func slength(th object.Thread, args ...object.Value) ([]object.Value, *object.RuntimeError) {
	ap := fnutil.NewArgParser(th, args)

	s, err := toSlice(ap, 0)
	if err != nil {
		return nil, err
	}

	return []object.Value{object.Integer(s.Len())}, nil
}

func sindex(th object.Thread, args ...object.Value) ([]object.Value, *object.RuntimeError) {
	ap := fnutil.NewArgParser(th, args)

	s, err := toSlice(ap, 0)
	if err != err {
		return nil, err
	}

	index, err := ap.ToGoInt(1)
	if err != nil {
		name, err := ap.ToGoString(1)
		if err != nil {
			return nil, err
		}

		if !isPublic(name) {
			return nil, nil
		}

		method := s.MethodByName(name)

		if !method.IsValid() {
			if s.CanAddr() {
				method = s.Addr().MethodByName(name)
			} else {
				self2 := reflect.New(s.Type())
				self2.Elem().Set(s)
				method = self2.MethodByName(name)
			}

			if !method.IsValid() {
				return nil, nil
			}
		}

		return []object.Value{valueOfReflect(method, false)}, nil
	}

	index--

	if 0 <= index && index < s.Len() {
		rval := s.Index(index)

		return []object.Value{valueOfReflect(rval, false)}, nil
	}

	return nil, object.NewRuntimeError(fmt.Sprintf("invalid array index %d (out of bounds for %d-element array)", index, s.Len()))
}

func snewindex(th object.Thread, args ...object.Value) ([]object.Value, *object.RuntimeError) {
	ap := fnutil.NewArgParser(th, args)

	s, err := toSlice(ap, 0)
	if err != err {
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

	if 0 <= index && index < s.Len() {
		styp := s.Type()
		vtyp := styp.Elem()

		if rval := toReflectValue(vtyp, val); rval.IsValid() {
			s.Index(index).Set(rval)

			return nil, nil
		}

		return nil, object.NewRuntimeError(fmt.Sprintf("non-%s array index %q", vtyp, reflect.TypeOf(val)))
	}

	return nil, object.NewRuntimeError(fmt.Sprintf("invalid array index %d (out of bounds for %d-element array)", index, s.Len()))
}

func spairs(th object.Thread, args ...object.Value) ([]object.Value, *object.RuntimeError) {
	ap := fnutil.NewArgParser(th, args)

	_, err := toSlice(ap, 0)
	if err != nil {
		return nil, err
	}

	return []object.Value{object.GoFunction(snext), args[0], object.Integer(0)}, nil
}

func snext(th object.Thread, args ...object.Value) ([]object.Value, *object.RuntimeError) {
	ap := fnutil.NewArgParser(th, args)

	s, err := toSlice(ap, 0)
	if err != nil {
		return nil, err
	}

	index, err := ap.ToGoInt(1)
	if err != nil {
		return nil, err
	}

	if index >= s.Len() {
		return nil, nil
	}

	rval := s.Index(index)

	index++

	return []object.Value{object.Integer(index), valueOfReflect(rval, false)}, nil
}
