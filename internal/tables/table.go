package tables

import (
	"fmt"

	"github.com/hirochachacha/plua/internal/limits"
	"github.com/hirochachacha/plua/object"
)

type table struct {
	alen int
	a    []object.Value
	m    *luaMap

	mt object.Table
}

func NewTableSize(asize, msize int) object.Table {
	return &table{
		alen: 0,
		a:    make([]object.Value, 0, asize),
		m:    newMapSize(msize),
	}
}

func NewTableArray(a []object.Value) object.Table {
	return &table{
		alen: len(a),
		a:    a,
		m:    newMapSize(0),
	}
}

func (t *table) Type() object.Type {
	return object.TTABLE
}

func (t *table) String() string {
	return fmt.Sprintf("table: %p", t)
}

func (t *table) Len() int {
	return t.alen
}

func (t *table) ikey(key object.Value) (object.Integer, bool) {
	if ikey, ok := key.(object.Integer); ok {
		return ikey, !(int64(ikey) > limits.MaxInt || int64(ikey) < limits.MinInt)
	}

	if nkey, ok := key.(object.Number); ok {
		if ikey, ok := object.ToInteger(nkey); ok {
			return ikey, !(int64(ikey) > limits.MaxInt || int64(ikey) < limits.MinInt)
		}
	}

	return 0, false
}

func (t *table) Get(key object.Value) object.Value {
	if ikey, ok := t.ikey(key); ok {
		return t.IGet(int(ikey))
	}

	return t.m.Get(key)
}

func (t *table) Set(key, val object.Value) {
	if ikey, ok := t.ikey(key); ok {
		t.ISet(int(ikey), val)
	} else {
		if val == nil {
			t.m.Delete(key)
		} else {
			t.m.Set(key, val)
		}
	}
}

func (t *table) Del(key object.Value) {
	if ikey, ok := t.ikey(key); ok {
		t.IDel(int(ikey))
	} else {
		t.m.Delete(key)
	}
}

func (t *table) IGet(i int) object.Value {
	if 0 < i && i <= len(t.a) {
		return t.a[i-1]
	}

	return t.m.Get(object.Integer(i))
}

func (t *table) ISet(i int, val object.Value) {
	if val == nil {
		switch {
		case 0 < i && i <= len(t.a):
			if t.alen > i-1 {
				t.alen = i - 1
			}
			t.a[i-1] = nil
		case i == len(t.a)+1:
			// do nothing
		default:
			t.m.Delete(object.Integer(i))
		}
	} else {
		switch {
		case 0 < i && i <= len(t.a):
			if t.a[i-1] == nil && t.alen == i {
				for j := i; j < len(t.a); j++ {
					t.alen++
					if t.a[j] == nil {
						break
					}
				}
			}

			t.a[i-1] = val
		case i == len(t.a)+1:
			t.a = append(t.a, val)

			// migration from map to array
			for {
				i++
				val := t.m.Get(object.Integer(i))
				if val == nil {
					break
				}
				t.a = append(t.a, val)
				t.m.Delete(object.Integer(i))
			}
			t.alen = len(t.a)
		default:
			t.m.Set(object.Integer(i), val)
		}
	}
}

func (t *table) IDel(i int) {
	switch {
	case 0 < i && i <= len(t.a):
		t.a[i-1] = nil

		copy(t.a[i-1:], t.a[i:])

		t.alen--
	case i == len(t.a)+1:
		// do nothing
	default:
		t.m.Delete(object.Integer(i))
	}
}

func (t *table) Next(key object.Value) (nkey, nval object.Value, ok bool) {
	if key == nil {
		if t.alen > 0 {
			return object.Integer(1), t.a[0], true
		}

		return t.m.Next(nil)
	}

	if ikey, ok := t.ikey(key); ok {
		if ikey < object.Integer(t.alen) {
			return ikey + 1, t.a[int(ikey)], true
		}
		if ikey == object.Integer(t.alen) {
			return t.m.Next(nil)
		}
	}

	return t.m.Next(key)
}

func (t *table) INext(i int) (ni int, nval object.Value, ok bool) {
	if t.alen <= i {
		return -1, nil, true
	}

	return i + 1, t.a[i], true
}

func (t *table) Sort(less func(x, y object.Value) bool) {
	ts := &tableSorter{a: t.a[:t.alen], less: less}

	ts.Sort()
}

func (t *table) SetList(base int, src []object.Value) {
	if len(src) < len(t.a)-base {
		copy(t.a[base:], src)
	} else {
		t.a = append(t.a[:base], src...)
	}

	if t.alen < base+len(src) {
		t.alen = base + len(src)
	}
}

func (t *table) SetMetatable(mt object.Table) {
	t.mt = mt
}

func (t *table) Metatable() object.Table {
	return t.mt
}
