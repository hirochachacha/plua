package reflect

import (
	"fmt"
	"reflect"

	"github.com/hirochachacha/plua/internal/tables"
	"github.com/hirochachacha/plua/object"
	"github.com/hirochachacha/plua/object/fnutil"
)

func buildFuncMT() {
	mt := tables.NewTableSize(0, 5)

	mt.Set(object.TM_METATABLE, object.True)
	mt.Set(object.TM_TOSTRING, object.GoFunction(tostring))

	mt.Set(object.TM_INDEX, object.GoFunction(index))

	mt.Set(object.String("__call"), object.GoFunction(call))

	mt.Set(object.TM_EQ, cmp(func(x, y reflect.Value) bool { return x.Pointer() == y.Pointer() }))

	funcMT = mt
}

func call(th object.Thread, args ...object.Value) ([]object.Value, *object.RuntimeError) {
	ap := fnutil.NewArgParser(th, args)

	self, err := ap.ToFullUserdata(0)
	if err != nil {
		return nil, err
	}

	if self, ok := self.Value.(reflect.Value); ok {
		styp := self.Type()

		var numin int
		if styp.IsVariadic() {
			numin = styp.NumIn() - 1
			if len(args)-1 > numin {
				numin = len(args) - 1
			}
		} else {
			numin = styp.NumIn()
		}

		rargs := make([]reflect.Value, numin)

		if len(args)-1 >= len(rargs) {
			for i := range rargs {
				if rarg := toReflectValue(styp.In(i), args[1+i]); rarg.IsValid() {
					rargs[i] = rarg
				} else {
					return nil, object.NewRuntimeError(fmt.Sprintf("mismatched types %s and %s", styp.In(i), reflect.TypeOf(args[1+i])))
				}
			}
		} else {
			for i, arg := range args[1:] {
				if rarg := toReflectValue(styp.In(i), arg); rarg.IsValid() {
					rargs[i] = rarg
				} else {
					return nil, object.NewRuntimeError(fmt.Sprintf("mismatched types %s and %s", styp.In(i), reflect.TypeOf(arg)))
				}
			}

			for i := len(args); i < len(rargs); i++ {
				if rarg := toReflectValue(styp.In(i), nil); rarg.IsValid() {
					rargs[i] = rarg
				} else {
					return nil, object.NewRuntimeError(fmt.Sprintf("mismatched types %s and %s", styp.In(i), reflect.TypeOf(nil)))
				}
			}
		}

		rrets := self.Call(rargs)

		rets := make([]object.Value, len(rrets))
		for i, rret := range rrets {
			rets[i] = valueOfReflect(rret, false)
		}

		return rets, nil
	}

	return nil, errInvalidUserdata
}
