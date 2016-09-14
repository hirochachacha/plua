package table

import (
	"bytes"
	"fmt"

	"github.com/hirochachacha/plua/internal/arith"
	"github.com/hirochachacha/plua/object"
	"github.com/hirochachacha/plua/object/fnutil"
)

func concat(th object.Thread, args ...object.Value) ([]object.Value, *object.RuntimeError) {
	ap := fnutil.NewArgParser(th, args)

	t, err := ap.ToTable(0)
	if err != nil {
		return nil, err
	}

	sep, err := ap.OptGoString(1, "")
	if err != nil {
		return nil, err
	}

	i, err := ap.OptGoInt(2, 1)
	if err != nil {
		return nil, err
	}

	j, err := ap.OptGoInt(3, t.Len())
	if err != nil {
		return nil, err
	}

	if i > j {
		return []object.Value{object.String("")}, nil
	}

	var buf bytes.Buffer

	var nkey int
	var val object.Value
	var ok bool
	var tmp string

	key := i - 1
	end := j - 1
	for {
		nkey, val, ok = t.INext(key)
		if !ok {
			return nil, object.NewRuntimeError("invalid key to 'inext'")
		}

		tmp, ok = object.ToGoString(val)
		if !ok {
			return nil, object.NewRuntimeError(fmt.Sprintf("invalid value (%s) at index %d in table for 'concat'", object.ToType(val), key))
		}

		buf.WriteString(tmp)

		if key == end {
			break
		}

		buf.WriteString(sep)

		key = nkey
	}

	return []object.Value{object.String(buf.String())}, nil
}

func insert(th object.Thread, args ...object.Value) ([]object.Value, *object.RuntimeError) {
	ap := fnutil.NewArgParser(th, args)

	t, err := ap.ToTable(0)
	if err != nil {
		return nil, err
	}

	val, ok := ap.Get(2)
	if !ok {
		val, err := ap.ToValue(1)
		if err != nil {
			return nil, err
		}

		t.ISet(t.Len()+1, val)

		return nil, nil
	}

	pos, err := ap.OptGoInt(1, t.Len())
	if err != nil {
		return nil, err
	}

	t.ISet(pos, val)

	return nil, nil
}

// TODO support for not table type
func move(th object.Thread, args ...object.Value) ([]object.Value, *object.RuntimeError) {
	ap := fnutil.NewArgParser(th, args)

	a1, err := ap.ToTable(0)
	if err != nil {
		return nil, err
	}

	f, err := ap.ToGoInt(1)
	if err != nil {
		return nil, err
	}

	e, err := ap.ToGoInt(2)
	if err != nil {
		return nil, err
	}

	t, err := ap.ToGoInt(3)
	if err != nil {
		return nil, err
	}

	a2 := a1
	if _, ok := ap.Get(4); ok {
		a2, err = ap.ToTable(4)
		if err != nil {
			return nil, err
		}
	}

	if f <= 0 {
		return nil, ap.ArgError(1, "initial position must be positive")
	}

	for i := 0; i <= e-f; i++ {
		a2.ISet(t+i, a1.IGet(f+i))
	}

	return nil, nil
}

func pack(th object.Thread, args ...object.Value) ([]object.Value, *object.RuntimeError) {
	t := th.NewTableArray(append([]object.Value{}, args...))

	t.Set(object.String("n"), object.Integer(len(args)))

	return []object.Value{t}, nil
}

func remove(th object.Thread, args ...object.Value) ([]object.Value, *object.RuntimeError) {
	ap := fnutil.NewArgParser(th, args)

	t, err := ap.ToTable(0)
	if err != nil {
		return nil, err
	}

	pos, err := ap.OptGoInt(1, t.Len())
	if err != nil {
		return nil, err
	}

	if val := t.IGet(pos); val != nil {
		t.IDel(pos)

		return []object.Value{val}, nil
	}

	return nil, nil
}

func sort(th object.Thread, args ...object.Value) ([]object.Value, *object.RuntimeError) {
	ap := fnutil.NewArgParser(th, args)

	t, err := ap.ToTable(0)
	if err != nil {
		return nil, err
	}

	var less func(x, y object.Value) bool

	if _, ok := ap.Get(1); ok {
		cmp, err := ap.ToFunction(1)
		if err != nil {
			return nil, err
		}

		less = func(x, y object.Value) bool {
			rets, err := th.Call(cmp, nil, x, y)
			if err != nil {
				return false
			}

			if len(rets) == 0 {
				return false
			}

			return object.ToGoBool(rets[0])
		}
	} else {
		less = func(x, y object.Value) bool {
			if b := arith.LessThan(x, y); b != nil {
				return bool(b.(object.Boolean))
			}

			tm := th.GetMetaField(x, "__lt")
			if tm == nil {
				tm = th.GetMetaField(y, "__lt")
				if tm == nil {
					tm = th.GetMetaField(x, "__le")
					if tm == nil {
						tm = th.GetMetaField(y, "__le")
					}
					x, y = y, x
				}
			}

			rets, err := th.Call(tm, nil, x, y)
			if err != nil {
				return false
			}

			if len(rets) == 0 {
				return false
			}

			return object.ToGoBool(rets[0])
		}
	}

	t.Sort(less)

	return nil, nil
}

func unpack(th object.Thread, args ...object.Value) ([]object.Value, *object.RuntimeError) {
	ap := fnutil.NewArgParser(th, args)

	t, err := ap.ToTable(0)
	if err != nil {
		return nil, err
	}

	i, err := ap.OptGoInt(1, 1)
	if err != nil {
		return nil, err
	}

	j, err := ap.OptGoInt(2, t.Len())
	if err != nil {
		return nil, err
	}

	if i > j {
		return nil, nil
	}

	var rets []object.Value

	var nkey int
	var val object.Value
	var ok bool

	key := i - 1
	end := j - 1
	for {
		nkey, val, ok = t.INext(key)
		if !ok {
			return nil, object.NewRuntimeError("invalid key to 'inext'")
		}

		rets = append(rets, val)

		if key == end {
			break
		}

		key = nkey
	}

	return rets, nil
}

func Open(th object.Thread, args ...object.Value) ([]object.Value, *object.RuntimeError) {
	m := th.NewTableSize(0, 7)

	m.Set(object.String("concat"), object.GoFunction(concat))
	m.Set(object.String("insert"), object.GoFunction(insert))
	m.Set(object.String("move"), object.GoFunction(move))
	m.Set(object.String("pack"), object.GoFunction(pack))
	m.Set(object.String("remove"), object.GoFunction(remove))
	m.Set(object.String("sort"), object.GoFunction(sort))
	m.Set(object.String("unpack"), object.GoFunction(unpack))

	return []object.Value{m}, nil
}
