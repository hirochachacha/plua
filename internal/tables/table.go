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

func (t *table) Get(key object.Value) object.Value {
	return t.get(normKey(key))
}

func (t *table) Set(key, val object.Value) {
	t.set(normKey(key), val)
}

func (t *table) Del(key object.Value) {
	t.del(normKey(key))
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

func (t *table) get(key object.Value) object.Value {
	if ikey, ok := t.ikey(key); ok {
		return t.iget(ikey)
	}
	return t.m.Get(key)
}

func (t *table) iget(ikey object.Integer) object.Value {
	i := int(ikey)
	if 0 < i && i <= len(t.a) {
		return t.a[i-1]
	}
	return t.m.Get(ikey)
}

func (t *table) set(key, val object.Value) {
	if val == nil {
		t.del(key)

		return
	}

	if ikey, ok := t.ikey(key); ok {
		t.iset(ikey, val)
	} else {
		t.m.Set(key, val)
	}
}

func (t *table) iset(ikey object.Integer, val object.Value) {
	i := int(ikey)
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
			ikey++
			val := t.m.Get(ikey)
			if val == nil {
				break
			}
			t.a = append(t.a, val)
			t.m.Delete(ikey)
		}

		t.alen = len(t.a)
	default:
		t.m.Set(ikey, val)
	}
}

func (t *table) del(key object.Value) {
	if ikey, ok := t.ikey(key); ok {
		t.idel(ikey)
	} else {
		t.m.Delete(key)
	}
}

func (t *table) idel(ikey object.Integer) {
	i := int(ikey)
	switch {
	case 0 < i && i <= len(t.a):
		if i <= t.alen {
			t.alen = i - 1
		}
		t.a[i-1] = nil
	case i == len(t.a)+1:
		// do nothing
	default:
		t.m.Delete(ikey)
	}
}

func (t *table) Next(key object.Value) (nkey, nval object.Value, ok bool) {
	key = normKey(key)
	if key == nil {
		if t.alen > 0 {
			return object.Integer(1), t.a[0], true
		}

		return t.m.Next(nil)
	}

	if ikey, ok := t.ikey(key); ok {
		i := int(ikey)
		if 0 <= i && i < t.alen {
			return ikey + 1, t.a[i], true
		}
		if ikey == object.Integer(t.alen) {
			return t.m.Next(nil)
		}
	}

	return t.m.Next(key)
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
