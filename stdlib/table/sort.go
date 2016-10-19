package table

import (
	"github.com/hirochachacha/plua/internal/arith"
	"github.com/hirochachacha/plua/object"
	"github.com/hirochachacha/plua/object/fnutil"
)

func sort(th object.Thread, args ...object.Value) ([]object.Value, *object.RuntimeError) {
	ap := fnutil.NewArgParser(th, args)

	t, err := ap.ToTable(0)
	if err != nil {
		return nil, err
	}

	var cmp object.Value

	if _, ok := ap.Get(1); ok {
		cmp, err = ap.ToFunction(1)
		if err != nil {
			return nil, err
		}
	}

	err = doSort(th, t, cmp)
	if err != nil {
		return nil, err
	}

	return nil, nil
}

func doSort(th object.Thread, t object.Table, cmp object.Value) *object.RuntimeError {
	tlen, err := callLen(th, t)
	if err != nil {
		return err
	}
	err = quickSort(th, t, cmp, 1, tlen)
	if err != nil {
		return err
	}
	return nil
}

func quickSort(th object.Thread, t object.Table, cmp object.Value, lo, hi int) *object.RuntimeError {
	if lo < hi {
		p, err := partition(th, t, cmp, lo, hi)
		if err != nil {
			return err
		}
		err = quickSort(th, t, cmp, lo, p)
		if err != nil {
			return err
		}
		err = quickSort(th, t, cmp, p+1, hi)
		if err != nil {
			return err
		}
	}
	return nil
}

func partition(th object.Thread, t object.Table, cmp object.Value, lo, hi int) (int, *object.RuntimeError) {
	pivot, err := arith.CallGettable(th, t, object.Integer(lo))
	if err != nil {
		return 0, err
	}

	i := lo - 1
	j := hi + 1

	var xi, xj object.Value

	for {
		for {
			i++
			xi, err = arith.CallGettable(th, t, object.Integer(i))
			if err != nil {
				return 0, err
			}
			b, err := docmp(th, cmp, xi, pivot)
			if err != nil {
				return 0, err
			}
			if !b {
				break
			}
		}
		for {
			j--
			xj, err = arith.CallGettable(th, t, object.Integer(j))
			if err != nil {
				return 0, err
			}
			b, err := docmp(th, cmp, pivot, xj)
			if err != nil {
				return 0, err
			}
			if !b {
				break
			}
		}
		if i >= j {
			return j, nil
		}
		err = doswap(th, t, i, j)
		if err != nil {
			return 0, err
		}
	}
}

func doswap(th object.Thread, t object.Table, i, j int) *object.RuntimeError {
	x, err := arith.CallGettable(th, t, object.Integer(i))
	if err != nil {
		return err
	}
	y, err := arith.CallGettable(th, t, object.Integer(j))
	if err != nil {
		return err
	}
	err = arith.CallSettable(th, t, object.Integer(j), x)
	if err != nil {
		return err
	}
	err = arith.CallSettable(th, t, object.Integer(i), y)
	if err != nil {
		return err
	}
	return nil
}

func docmp(th object.Thread, cmp, x, y object.Value) (bool, *object.RuntimeError) {
	if cmp == nil {
		b, err := arith.CallLessThan(th, false, x, y)
		if err != nil {
			return false, err
		}

		return b, nil
	}

	rets, err := th.Call(cmp, nil, x, y)
	if err != nil {
		return false, err
	}
	if len(rets) == 0 {
		return false, nil
	}
	return object.ToGoBool(rets[0]), nil
}
