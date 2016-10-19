package arith

import (
	"github.com/hirochachacha/plua/internal/errors"
	"github.com/hirochachacha/plua/internal/version"
	"github.com/hirochachacha/plua/object"
)

func CallGettable(th object.Thread, t, key object.Value) (object.Value, *object.RuntimeError) {
	for i := 0; i < version.MAX_TAG_LOOP; i++ {
		var tm object.Value
		if tab, ok := t.(object.Table); ok {
			val := tab.Get(key)
			tm = gettm(tab.Metatable(), object.TM_INDEX)
			if val != nil || tm == nil {
				return val, nil
			}
		} else {
			tm = gettmbyobj(th, t, object.TM_INDEX)
		}

		if tm == nil {
			return nil, errors.IndexError(t)
		}

		if isFunction(tm) {
			return calltm(th, tm, t, key)
		}

		t = tm
	}

	return nil, errors.ErrGetTable
}

func CallSettable(th object.Thread, t, key, val object.Value) *object.RuntimeError {
	for i := 0; i < version.MAX_TAG_LOOP; i++ {
		var tm object.Value
		if tab, ok := t.(object.Table); ok {
			old := tab.Get(key)
			tm = gettm(tab.Metatable(), object.TM_NEWINDEX)
			if old != nil || tm == nil {
				if key == nil {
					return errors.ErrNilIndex
				}

				if key == object.NaN {
					return errors.ErrNaNIndex
				}

				tab.Set(key, val)

				return nil
			}
		} else {
			tm = gettmbyobj(th, t, object.TM_NEWINDEX)
		}

		if tm == nil {
			return errors.IndexError(t)
		}

		if isFunction(tm) {
			_, err := calltm(th, tm, t, key, val)
			return err
		}

		t = tm
	}

	return errors.ErrSetTable
}

func isFunction(val object.Value) bool {
	return object.ToType(val) == object.TFUNCTION
}
