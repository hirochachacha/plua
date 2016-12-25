package reflect

import (
	"fmt"
	"reflect"

	"github.com/hirochachacha/plua/internal/tables"
	"github.com/hirochachacha/plua/object"
	"github.com/hirochachacha/plua/object/fnutil"
)

func buildMapMT() {
	mt := tables.NewTableSize(0, 7)

	mt.Set(object.TM_METATABLE, object.True)
	mt.Set(object.TM_NAME, object.String("MAP*"))
	mt.Set(object.TM_TOSTRING, object.GoFunction(mtostring))

	mt.Set(object.TM_INDEX, object.GoFunction(mindex))
	mt.Set(object.TM_NEWINDEX, object.GoFunction(mnewindex))
	mt.Set(object.TM_LEN, object.GoFunction(mlength))
	mt.Set(object.TM_PAIRS, object.GoFunction(mpairs))

	mt.Set(object.TM_EQ, cmp(func(x, y reflect.Value) bool { return x.Pointer() == y.Pointer() }, toMap))

	mapMT = mt
}

func toMap(ap *fnutil.ArgParser, n int) (reflect.Value, *object.RuntimeError) {
	val, err := toValue(ap, n, "MAP*")
	if err != nil {
		return reflect.Value{}, err
	}
	if val.Kind() != reflect.Map {
		return reflect.Value{}, ap.TypeError(n, "MAP*")
	}
	return val, nil
}

func mtostring(th object.Thread, args ...object.Value) ([]object.Value, *object.RuntimeError) {
	ap := fnutil.NewArgParser(th, args)

	m, err := toMap(ap, 0)
	if err != nil {
		return nil, err
	}

	return []object.Value{object.String(fmt.Sprintf("go map (0x%x)", m.Pointer()))}, nil
}

func mlength(th object.Thread, args ...object.Value) ([]object.Value, *object.RuntimeError) {
	ap := fnutil.NewArgParser(th, args)

	m, err := toMap(ap, 0)
	if err != nil {
		return nil, err
	}

	return []object.Value{object.Integer(m.Len())}, nil
}

func mindex(th object.Thread, args ...object.Value) ([]object.Value, *object.RuntimeError) {
	ap := fnutil.NewArgParser(th, args)

	m, err := toMap(ap, 0)
	if err != nil {
		return nil, err
	}

	key, err := ap.ToValue(1)
	if err != nil {
		return nil, err
	}

	ktyp := m.Type().Key()

	if rkey := toReflectValue(ktyp, key); rkey.IsValid() {
		rval := m.MapIndex(rkey)

		return []object.Value{valueOfReflect(rval, false)}, nil
	}

	return nil, object.NewRuntimeError(fmt.Sprintf("cannot use %v (type %s) as type %s in map index", key, reflect.TypeOf(key), ktyp))
}

func mnewindex(th object.Thread, args ...object.Value) ([]object.Value, *object.RuntimeError) {
	ap := fnutil.NewArgParser(th, args)

	m, err := toMap(ap, 0)
	if err != nil {
		return nil, err
	}

	key, err := ap.ToValue(1)
	if err != nil {
		return nil, err
	}

	val, err := ap.ToValue(2)
	if err != nil {
		return nil, err
	}

	styp := m.Type()
	ktyp := styp.Key()
	vtyp := styp.Elem()

	if rkey := toReflectValue(ktyp, key); rkey.IsValid() {
		if rval := toReflectValue(vtyp, val); rval.IsValid() {
			m.SetMapIndex(rkey, rval)

			return nil, nil
		}

		return nil, object.NewRuntimeError(fmt.Sprintf("cannot use %v (type %s) as type %s in map assignment", val, reflect.TypeOf(val), vtyp))
	}

	return nil, object.NewRuntimeError(fmt.Sprintf("cannot use %v (type %s) as type %s in map index", key, reflect.TypeOf(key), ktyp))
}

func mpairs(th object.Thread, args ...object.Value) ([]object.Value, *object.RuntimeError) {
	ap := fnutil.NewArgParser(th, args)

	m, err := toMap(ap, 0)
	if err != nil {
		return nil, err
	}

	keys := m.MapKeys()
	length := len(keys)

	i := 0

	next := func(th object.Thread, args ...object.Value) ([]object.Value, *object.RuntimeError) {
		if i == length {
			return nil, nil
		}

		key := keys[i]
		rval := m.MapIndex(key)

		i++

		return []object.Value{valueOfReflect(key, false), valueOfReflect(rval, false)}, nil
	}

	return []object.Value{object.GoFunction(next)}, nil
}
