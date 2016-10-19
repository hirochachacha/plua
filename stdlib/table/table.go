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

	tlen, err := callLen(th, t)
	if err != nil {
		return nil, err
	}

	j, err := ap.OptGoInt(3, tlen)
	if err != nil {
		return nil, err
	}

	if i > j {
		return []object.Value{object.String("")}, nil
	}

	var buf bytes.Buffer

	var val object.Value
	var ok bool
	var tmp string

	for {
		val, err = arith.CallGettable(th, t, object.Integer(i))
		if err != nil {
			return nil, err
		}

		tmp, ok = object.ToGoString(val)
		if !ok {
			return nil, object.NewRuntimeError(fmt.Sprintf("invalid value (%s) at index %d in table for 'concat'", object.ToType(val), i))
		}

		buf.WriteString(tmp)

		if i == j {
			break
		}

		buf.WriteString(sep)

		i++
	}

	return []object.Value{object.String(buf.String())}, nil
}

// table.insert (list, [pos,] value)
func insert(th object.Thread, args ...object.Value) ([]object.Value, *object.RuntimeError) {
	ap := fnutil.NewArgParser(th, args)

	t, err := ap.ToTable(0)
	if err != nil {
		return nil, err
	}

	tlen, err := callLen(th, t)
	if err != nil {
		return nil, err
	}

	val, ok := ap.Get(2)
	if !ok {
		val, err := ap.ToValue(1)
		if err != nil {
			return nil, err
		}

		err = arith.CallSettable(th, t, object.Integer(tlen+1), val)
		if err != nil {
			return nil, err
		}

		return nil, nil
	}

	pos, err := ap.OptGoInt(1, tlen+1)
	if err != nil {
		return nil, err
	}

	if pos < 1 || pos > tlen+1 {
		return nil, ap.ArgError(1, "position out of bounds")
	}

	for i := tlen + 1; i > pos; i-- {
		v, err := arith.CallGettable(th, t, object.Integer(i-1))
		if err != nil {
			return nil, err
		}
		err = arith.CallSettable(th, t, object.Integer(i), v)
		if err != nil {
			return nil, err
		}
	}

	err = arith.CallSettable(th, t, object.Integer(pos), val)
	if err != nil {
		return nil, err
	}

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

	for i := 0; i <= e-f; i++ {
		v, err := arith.CallGettable(th, a1, object.Integer(f+i))
		if err != nil {
			return nil, err
		}
		err = arith.CallSettable(th, a2, object.Integer(t+i), v)
		if err != nil {
			return nil, err
		}
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

	tlen, err := callLen(th, t)
	if err != nil {
		return nil, err
	}

	pos, err := ap.OptGoInt(1, tlen)
	if err != nil {
		return nil, err
	}

	if tlen == pos {
		val, err := arith.CallGettable(th, t, object.Integer(pos))
		if err != nil {
			return nil, err
		}
		if val != nil {
			err = arith.CallSettable(th, t, object.Integer(pos), nil)
			if err != nil {
				return nil, err
			}
		}
		return []object.Value{val}, nil
	}

	if pos < 1 || pos > tlen+1 {
		return nil, ap.ArgError(1, "position out of bounds")
	}

	val, err := arith.CallGettable(th, t, object.Integer(pos))
	if err != nil {
		return nil, err
	}

	for i := pos; i < tlen+1; i++ {
		v, err := arith.CallGettable(th, t, object.Integer(i+1))
		if err != nil {
			return nil, err
		}
		err = arith.CallSettable(th, t, object.Integer(i), v)
		if err != nil {
			return nil, err
		}
	}

	return []object.Value{val}, nil
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

	tlen, err := callLen(th, t)
	if err != nil {
		return nil, err
	}

	j, err := ap.OptGoInt(2, tlen)
	if err != nil {
		return nil, err
	}

	if i > j {
		return nil, nil
	}

	var rets []object.Value

	var val object.Value

	for {
		val, err = arith.CallGettable(th, t, object.Integer(i))
		if err != nil {
			return nil, err
		}

		rets = append(rets, val)

		if i == j {
			break
		}

		i++
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
