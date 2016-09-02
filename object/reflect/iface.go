package reflect

import (
	"fmt"
	"reflect"

	"github.com/hirochachacha/plua/internal/tables"
	"github.com/hirochachacha/plua/object"
	"github.com/hirochachacha/plua/object/fnutil"
)

func buildIfaceMT() {
	mt := tables.NewTableSize(0, 4)

	mt.Set(object.String("__metatable"), object.True)
	mt.Set(object.String("__tostring"), object.GoFunction(tostring))

	mt.Set(object.String("__index"), object.GoFunction(iindex))

	mt.Set(object.String("__eq"), cmp(func(x, y reflect.Value) bool { return x.Interface() == y.Interface() }))

	ifaceMT = mt
}

func iindex(th object.Thread, args ...object.Value) ([]object.Value, *object.RuntimeError) {
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

		if method.IsValid() {
			return []object.Value{valueOfReflect(method, false)}, nil
		}

		return nil, object.NewRuntimeError(fmt.Sprintf("type %s has no method %s", self.Type(), name))
	}

	return nil, errInvalidUserdata
}
