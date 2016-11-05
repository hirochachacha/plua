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
	mt.Set(object.TM_TOSTRING, object.GoFunction(tostring))

	mt.Set(object.TM_INDEX, object.GoFunction(mindex))
	mt.Set(object.TM_NEWINDEX, object.GoFunction(mnewindex))
	mt.Set(object.TM_LEN, object.GoFunction(length))
	mt.Set(object.TM_PAIRS, object.GoFunction(mpairs))

	mt.Set(object.TM_EQ, cmp(func(x, y reflect.Value) bool { return x.Pointer() == y.Pointer() }))

	mapMT = mt
}

func mindex(th object.Thread, args ...object.Value) ([]object.Value, *object.RuntimeError) {
	ap := fnutil.NewArgParser(th, args)

	self, err := ap.ToFullUserdata(0)
	if err != nil {
		return nil, err
	}

	key, err := ap.ToValue(1)
	if err != nil {
		return nil, err
	}

	if self, ok := self.Value.(reflect.Value); ok {
		ktyp := self.Type().Key()

		if rkey := toReflectValue(ktyp, key); rkey.IsValid() {
			rval := self.MapIndex(rkey)

			return []object.Value{valueOfReflect(rval, false)}, nil
		}

		return nil, object.NewRuntimeError(fmt.Sprintf("cannot use %v (type %s) as type %s in map index", key, reflect.TypeOf(key), ktyp))
	}

	return nil, errInvalidUserdata
}

func mnewindex(th object.Thread, args ...object.Value) ([]object.Value, *object.RuntimeError) {
	ap := fnutil.NewArgParser(th, args)

	self, err := ap.ToFullUserdata(0)
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

	if self, ok := self.Value.(reflect.Value); ok {
		styp := self.Type()
		ktyp := styp.Key()
		vtyp := styp.Elem()

		if rkey := toReflectValue(ktyp, key); rkey.IsValid() {
			if rval := toReflectValue(vtyp, val); rval.IsValid() {
				self.SetMapIndex(rkey, rval)

				return nil, nil
			}

			return nil, object.NewRuntimeError(fmt.Sprintf("cannot use %v (type %s) as type %s in map assignment", val, reflect.TypeOf(val), vtyp))
		}

		return nil, object.NewRuntimeError(fmt.Sprintf("cannot use %v (type %s) as type %s in map index", key, reflect.TypeOf(key), ktyp))
	}

	return nil, errInvalidUserdata
}

func mpairs(th object.Thread, args ...object.Value) ([]object.Value, *object.RuntimeError) {
	ap := fnutil.NewArgParser(th, args)

	self, err := ap.ToFullUserdata(0)
	if err != nil {
		return nil, err
	}

	if self, ok := self.Value.(reflect.Value); ok {
		keys := self.MapKeys()
		length := len(keys)

		i := 0

		next := func(th object.Thread, args ...object.Value) ([]object.Value, *object.RuntimeError) {
			if i == length {
				return nil, nil
			}

			key := keys[i]
			rval := self.MapIndex(key)

			i++

			return []object.Value{valueOfReflect(key, false), valueOfReflect(rval, false)}, nil
		}

		return []object.Value{object.GoFunction(next)}, nil
	}

	return nil, errInvalidUserdata
}
