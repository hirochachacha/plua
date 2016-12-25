package reflect

import (
	"fmt"
	"reflect"
	"unicode"
	"unicode/utf8"

	"github.com/hirochachacha/plua/object"
	"github.com/hirochachacha/plua/object/fnutil"
)

type toValuefn func(ap *fnutil.ArgParser, n int) (reflect.Value, *object.RuntimeError)

func tostring(toValue toValuefn) object.GoFunction {
	return func(th object.Thread, args ...object.Value) ([]object.Value, *object.RuntimeError) {
		ap := fnutil.NewArgParser(th, args)

		self, err := toValue(ap, 0)
		if err != nil {
			return nil, err
		}

		return []object.Value{object.String(fmt.Sprintf("%v", self))}, nil
	}
}

func index(toValue toValuefn) object.GoFunction {
	return func(th object.Thread, args ...object.Value) ([]object.Value, *object.RuntimeError) {
		ap := fnutil.NewArgParser(th, args)

		self, err := toValue(ap, 0)
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
				return nil, nil
			}
		}

		return []object.Value{valueOfReflect(method, false)}, nil
	}
}

func cmp(op func(x, y reflect.Value) bool, toValue toValuefn) object.GoFunction {
	return func(th object.Thread, args ...object.Value) ([]object.Value, *object.RuntimeError) {
		ap := fnutil.NewArgParser(th, args)

		x, err := toValue(ap, 0)
		if err != nil {
			return nil, err
		}

		y, err := toValue(ap, 1)
		if err != nil {
			return nil, err
		}

		return []object.Value{object.Boolean(op(x, y))}, nil
	}
}

func unary(op func(x reflect.Value) reflect.Value, toValue toValuefn, mt object.Table) object.GoFunction {
	return func(th object.Thread, args ...object.Value) ([]object.Value, *object.RuntimeError) {
		ap := fnutil.NewArgParser(th, args)

		x, err := toValue(ap, 0)
		if err != nil {
			return nil, err
		}

		val := op(x)

		ud := &object.Userdata{Value: val.Convert(x.Type()), Metatable: mt}

		return []object.Value{ud}, nil
	}
}

func binary(op func(x, y reflect.Value) (reflect.Value, *object.RuntimeError), toValue toValuefn, mt object.Table) object.GoFunction {
	return func(th object.Thread, args ...object.Value) ([]object.Value, *object.RuntimeError) {
		ap := fnutil.NewArgParser(th, args)

		x, err := toValue(ap, 0)
		if err != nil {
			return nil, err
		}

		y, err := toValue(ap, 1)
		if err != nil {
			return nil, err
		}

		val, err := op(x, y)
		if err != nil {
			return nil, err
		}

		ud := &object.Userdata{Value: val.Convert(x.Type()), Metatable: mt}

		return []object.Value{ud}, nil
	}
}

func isPublic(name string) bool {
	r, _ := utf8.DecodeRuneInString(name)

	return unicode.IsUpper(r)
}

func toValue(ap *fnutil.ArgParser, n int, tname string) (reflect.Value, *object.RuntimeError) {
	ud, err := ap.ToFullUserdata(n)
	if err != nil {
		return reflect.Value{}, ap.TypeError(n, tname)
	}

	if ud.Metatable == nil || ud.Metatable.Get(object.TM_NAME) != object.String(tname) {
		return reflect.Value{}, ap.TypeError(n, tname)
	}

	val, ok := ud.Value.(reflect.Value)
	if !ok {
		return reflect.Value{}, ap.TypeError(n, tname)
	}

	return val, nil
}
