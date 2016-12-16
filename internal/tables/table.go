package tables

import (
	"fmt"

	"github.com/hirochachacha/plua/internal/limits"
	"github.com/hirochachacha/plua/object"
)

type table struct {
	a []object.Value
	m *luaMap

	mt object.Table
}

func NewTableSize(asize, msize int) object.Table {
	return &table{
		a: make([]object.Value, 0, asize),
		m: newMapSize(msize),
	}
}

func NewTableArray(a []object.Value) object.Table {
	return &table{
		a: a,
		m: newMapSize(0),
	}
}

func (t *table) Type() object.Type {
	return object.TTABLE
}

func (t *table) String() string {
	return fmt.Sprintf("table: %p", t)
}

func (t *table) Len() int {
	return len(t.a)
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

func (t *table) Next(key object.Value) (nkey, nval object.Value, ok bool) {
	return t.next(normKey(key))
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
	case 0 < i && i < len(t.a):
		t.a[i-1] = nil
	case 0 < i && i == len(t.a):
		t.a = t.a[:len(t.a)-1]
		for len(t.a) > 0 && t.a[len(t.a)-1] == nil {
			t.a = t.a[:len(t.a)-1]
		}
	case i == len(t.a)+1:
		// do nothing
	default:
		t.m.Delete(ikey)
	}
}

func (t *table) next(key object.Value) (nkey, nval object.Value, ok bool) {
	if key == nil {
		for i := 0; i < len(t.a); i++ {
			v := t.a[i]
			if v != nil {
				return object.Integer(i + 1), v, true
			}
		}
		return t.m.Next(nil)
	}

	if ikey, ok := t.ikey(key); ok {
		if i := int(ikey); i >= 0 {
			for ; i < len(t.a); i++ {
				v := t.a[i]
				if v != nil {
					return object.Integer(i + 1), t.a[i], true
				}
			}
			if i == len(t.a) {
				return t.m.Next(nil)
			}
		}
	}

	return t.m.Next(key)
}

func (t *table) SetList(base int, src []object.Value) {
	if len(src) < len(t.a)-base {
		copy(t.a[base:], src)
	} else {
		t.a = append(t.a[:base], src...)
	}
}

func (t *table) SetMetatable(mt object.Table) {
	t.mt = mt
}

func (t *table) Metatable() object.Table {
	return t.mt
}
