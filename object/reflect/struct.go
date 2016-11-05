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
	mt.Set(object.TM_TOSTRING, object.GoFunction(tostring))

	mt.Set(object.TM_INDEX, object.GoFunction(sindex))
	mt.Set(object.TM_NEWINDEX, object.GoFunction(snewindex))

	mt.Set(object.TM_EQ, cmp(func(x, y reflect.Value) bool { return x.Interface() == y.Interface() }))

	structMT = mt
}

func sindex(th object.Thread, args ...object.Value) ([]object.Value, *object.RuntimeError) {
	ap := fnutil.NewArgParser(th, args)

	self, err := ap.ToFullUserdata(0)
	if err != nil {
		return nil, err
	}

	name, err := ap.ToGoString(1)
	if err != nil {
		return nil, err
	}

	if !isPublic(name) {
		return nil, object.NewRuntimeError(fmt.Sprintf("%s is not public method or field", name))
	}

	if self, ok := self.Value.(reflect.Value); ok {
		method := self.MethodByName(name)

		if !method.IsValid() {
			if self.CanAddr() {
				method = self.Addr().MethodByName(name)
			} else {
				self2 := reflect.New(self.Type())
				self2.Elem().Set(self)
				method = self2.MethodByName(name)
			}

			if !method.IsValid() {
				field := self.FieldByName(name)
				if !field.IsValid() {
					return nil, object.NewRuntimeError(fmt.Sprintf("type %s has no field or method %s", self.Type(), name))
				}
				return []object.Value{valueOfReflect(field, false)}, nil
			}
			return nil, object.NewRuntimeError(fmt.Sprintf("type %s has no method %s", self.Type(), name))
		}

		return []object.Value{valueOfReflect(method, false)}, nil
	}

	return nil, errInvalidUserdata
}

func snewindex(th object.Thread, args ...object.Value) ([]object.Value, *object.RuntimeError) {
	ap := fnutil.NewArgParser(th, args)

	self, err := ap.ToFullUserdata(0)
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

	if self, ok := self.Value.(reflect.Value); ok {
		field := self.FieldByName(name)
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

		return nil, object.NewRuntimeError(fmt.Sprintf("type %s has no field %s", self.Type(), name))
	}

	return nil, errInvalidUserdata
}
