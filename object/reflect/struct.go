package reflect

import (
	"fmt"
	"reflect"

	"github.com/hirochachacha/plua/internal/tables"
	"github.com/hirochachacha/plua/object"
	"github.com/hirochachacha/plua/object/fnutil"
)

func buildStructMT() {
	mt := tables.NewTableSize(0, 5)

	mt.Set(object.TM_METATABLE, object.True)
	mt.Set(object.TM_NAME, object.String("STRUCT*"))
	mt.Set(object.TM_TOSTRING, object.GoFunction(sttostring))

	mt.Set(object.TM_INDEX, object.GoFunction(stindex))
	mt.Set(object.TM_NEWINDEX, object.GoFunction(stnewindex))

	mt.Set(object.TM_EQ, cmp(func(x, y reflect.Value) bool { return x.Interface() == y.Interface() }, toStruct))

	structMT = mt
}

func toStruct(ap *fnutil.ArgParser, n int) (reflect.Value, *object.RuntimeError) {
	val, err := toValue(ap, n, "STRUCT*")
	if err != nil {
		return reflect.Value{}, err
	}
	if val.Kind() != reflect.Struct {
		return reflect.Value{}, ap.TypeError(n, "STRUCT*")
	}
	return val, nil
}

func sttostring(th object.Thread, args ...object.Value) ([]object.Value, *object.RuntimeError) {
	ap := fnutil.NewArgParser(th, args)

	_, err := toStruct(ap, 0)
	if err != nil {
		return nil, err
	}

	return []object.Value{object.String(fmt.Sprintf("go struct"))}, nil
}

func stindex(th object.Thread, args ...object.Value) ([]object.Value, *object.RuntimeError) {
	ap := fnutil.NewArgParser(th, args)

	s, err := toStruct(ap, 0)
	if err != nil {
		return nil, err
	}

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
			field := s.FieldByName(name)
			if !field.IsValid() {
				return nil, nil
			}
			return []object.Value{valueOfReflect(field, false)}, nil
		}
		return nil, nil
	}

	return []object.Value{valueOfReflect(method, false)}, nil
}

func stnewindex(th object.Thread, args ...object.Value) ([]object.Value, *object.RuntimeError) {
	ap := fnutil.NewArgParser(th, args)

	s, err := toStruct(ap, 0)
	if err != nil {
		return nil, err
	}

	name, err := ap.ToGoString(1)
	if err != nil {
		return nil, err
	}

	val, err := ap.ToValue(2)
	if err != nil {
		return nil, err
	}

	if !isPublic(name) {
		return nil, object.NewRuntimeError(fmt.Sprintf("%s is not public method or field", name))
	}

	field := s.FieldByName(name)
	if field.IsValid() {
		if field.Kind() == reflect.Ptr {
			field = field.Elem()
		}

		if rval := toReflectValue(field.Type(), val); rval.IsValid() {
			field.Set(rval)

			return nil, nil
		}

		return nil, object.NewRuntimeError(fmt.Sprintf("cannot use %v (type %s) as type %s in field assignment", val, reflect.TypeOf(val), field.Type()))
	}

	return nil, object.NewRuntimeError(fmt.Sprintf("type %s has no field %s", s.Type(), name))
}
