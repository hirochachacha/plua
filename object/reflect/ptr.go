package reflect

import (
	"fmt"
	"reflect"

	"github.com/hirochachacha/plua/internal/tables"
	"github.com/hirochachacha/plua/object"
	"github.com/hirochachacha/plua/object/fnutil"
)

func buildPtrMT() {
	mt := tables.NewTableSize(0, 5)

	mt.Set(object.String("__metatable"), object.True)
	mt.Set(object.String("__tostring"), object.GoFunction(tostring))

	mt.Set(object.String("__index"), object.GoFunction(pindex))
	mt.Set(object.String("__newindex"), object.GoFunction(pnewindex))

	mt.Set(object.String("__eq"), cmp(func(x, y reflect.Value) bool { return x.Pointer() == y.Pointer() }))

	ptrMT = mt
}

func pindex(th object.Thread, args ...object.Value) ([]object.Value, *object.RuntimeError) {
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
			elem := self.Elem()

			method = elem.MethodByName(name)

			if !method.IsValid() {
				if elem.Kind() == reflect.Struct {
					field := elem.FieldByName(name)
					if field.IsValid() {
						return []object.Value{valueOfReflect(field, false)}, nil
					}
				}
				return nil, object.NewRuntimeError(fmt.Sprintf("type %s has no field or method %s", self.Type(), name))
			}
		}

		return []object.Value{valueOfReflect(method, false)}, nil
	}

	return nil, errInvalidUserdata
}

func pnewindex(th object.Thread, args ...object.Value) ([]object.Value, *object.RuntimeError) {
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
		elem := self.Elem()
		if elem.Kind() == reflect.Struct {
			field := elem.FieldByName(name)
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
		}

		return nil, object.NewRuntimeError(fmt.Sprintf("type %s has no field or method %s", self.Type(), name))
	}

	return nil, errInvalidUserdata
}
