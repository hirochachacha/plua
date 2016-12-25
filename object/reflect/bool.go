package reflect

import (
	"reflect"

	"github.com/hirochachacha/plua/internal/tables"
	"github.com/hirochachacha/plua/object"
	"github.com/hirochachacha/plua/object/fnutil"
)

func buildBoolMT() {
	mt := tables.NewTableSize(0, 4)

	mt.Set(object.TM_METATABLE, object.True)
	mt.Set(object.TM_NAME, object.String("BOOL*"))
	mt.Set(object.TM_TOSTRING, tostring(toBool))
	mt.Set(object.TM_INDEX, index(toBool))

	mt.Set(object.TM_EQ, cmp(func(x, y reflect.Value) bool { return x.Bool() == y.Bool() }, toBool))

	boolMT = mt
}

func toBool(ap *fnutil.ArgParser, n int) (reflect.Value, *object.RuntimeError) {
	if b, err := ap.ToGoBool(n); err == nil {
		return reflect.ValueOf(b), nil
	}
	val, err := toValue(ap, n, "BOOL*")
	if err != nil {
		return reflect.Value{}, err
	}
	if val.Kind() != reflect.Bool {
		return reflect.Value{}, ap.TypeError(n, "BOOL*")
	}
	return val, nil
}
