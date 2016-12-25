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

	mt.Set(object.TM_METATABLE, object.True)
	mt.Set(object.TM_NAME, object.String("PTR*"))
	mt.Set(object.TM_TOSTRING, object.GoFunction(ptostring))

	mt.Set(object.TM_INDEX, object.GoFunction(pindex))
	mt.Set(object.TM_NEWINDEX, object.GoFunction(pnewindex))

	mt.Set(object.TM_EQ, cmp(func(x, y reflect.Value) bool { return x.Pointer() == y.Pointer() }, toPtr))

	ptrMT = mt
}

func toPtr(ap *fnutil.ArgParser, n int) (reflect.Value, *object.RuntimeError) {
	val, err := toValue(ap, n, "PTR*")
	if err != nil {
		return reflect.Value{}, err
	}
	if val.Kind() != reflect.Ptr {
		return reflect.Value{}, ap.TypeError(n, "PTR*")
	}
	return val, nil
}

func ptostring(th object.Thread, args ...object.Value) ([]object.Value, *object.RuntimeError) {
	ap := fnutil.NewArgParser(th, args)

	p, err := toPtr(ap, 0)
	if err != nil {
		return nil, err
	}

	return []object.Value{object.String(fmt.Sprintf("go pointer (0x%x)", p.Pointer()))}, nil
}

func pindex(th object.Thread, args ...object.Value) ([]object.Value, *object.RuntimeError) {
	ap := fnutil.NewArgParser(th, args)

	p, err := toPtr(ap, 0)
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

	method := p.MethodByName(name)

	if !method.IsValid() {
		elem := p.Elem()

		method = elem.MethodByName(name)

		if !method.IsValid() {
			if elem.Kind() == reflect.Struct {
				field := elem.FieldByName(name)
				if field.IsValid() {
					return []object.Value{valueOfReflect(field, false)}, nil
				}
			}
			return nil, nil
		}
	}

	return []object.Value{valueOfReflect(method, false)}, nil
}

func pnewindex(th object.Thread, args ...object.Value) ([]object.Value, *object.RuntimeError) {
	ap := fnutil.NewArgParser(th, args)

	p, err := toPtr(ap, 0)
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

	elem := p.Elem()
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

	return nil, object.NewRuntimeError(fmt.Sprintf("type %s has no field or method %s", p.Type(), name))
}
