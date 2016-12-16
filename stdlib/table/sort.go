package table

import (
	"github.com/hirochachacha/plua/internal/arith"
	"github.com/hirochachacha/plua/internal/limits"
	"github.com/hirochachacha/plua/object"
	"github.com/hirochachacha/plua/object/fnutil"
)

func sort(th object.Thread, args ...object.Value) ([]object.Value, *object.RuntimeError) {
	ap := fnutil.NewArgParser(th, args)

	t, err := ap.ToTable(0)
	if err != nil {
		return nil, err
	}

	var lt object.Value

	if len(args) > 1 {
		lt, err = ap.ToTypes(1, object.TFUNCTION, object.TNIL)
		if err != nil {
			return nil, err
		}
	}

	tlen, err := callLen(th, t)
	if err != nil {
		return nil, err
	}

	if tlen >= int(limits.MaxInt) {
		return nil, ap.ArgError(0, "array too big")
	}

	err = quickSort(th, t, lt, 1, tlen)
	if err != nil {
		return nil, err
	}

	return nil, nil
}

func quickSort(th object.Thread, t object.Table, lt object.Value, lo, hi int) *object.RuntimeError {
	if lo < hi {
		p, err := partition(th, t, lt, lo, hi)
		if err != nil {
			return err
		}
		err = quickSort(th, t, lt, lo, p)
		if err != nil {
			return err
		}
		err = quickSort(th, t, lt, p+1, hi)
		if err != nil {
			return err
		}
	}
	return nil
}

func partition(th object.Thread, t object.Table, lt object.Value, lo, hi int) (int, *object.RuntimeError) {
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
			b, err := callLessThan(th, lt, xi, pivot)
			if err != nil {
				return 0, err
			}
			if !b {
				break
			}
			if i == hi {
				return 0, object.NewRuntimeError("invalid order function for sorting")
			}
		}
		for {
			j--
			xj, err = arith.CallGettable(th, t, object.Integer(j))
			if err != nil {
				return 0, err
			}
			b, err := callLessThan(th, lt, pivot, xj)
			if err != nil {
				return 0, err
			}
			if !b {
				break
			}
			if j < i {
				return 0, object.NewRuntimeError("invalid order function for sorting")
			}
		}
		if i >= j {
			return j, nil
		}
		err = tabSwap(th, t, i, j)
		if err != nil {
			return 0, err
		}
	}
}

func tabSwap(th object.Thread, t object.Table, i, j int) *object.RuntimeError {
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

func callLessThan(th object.Thread, lt, x, y object.Value) (bool, *object.RuntimeError) {
	if lt == nil {
		b, err := arith.CallLessThan(th, false, x, y)
		if err != nil {
			return false, err
		}

		return b, nil
	}

	rets, err := th.Call(lt, nil, x, y)
	if err != nil {
		return false, err
	}
	if len(rets) == 0 {
		return false, nil
	}
	return object.ToGoBool(rets[0]), nil
}
