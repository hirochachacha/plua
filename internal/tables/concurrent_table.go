package tables

import (
	"fmt"
	"sync"

	"github.com/hirochachacha/plua/internal/limits"
	"github.com/hirochachacha/plua/object"
)

type concurrentTable struct {
	a []object.Value
	m *concurrentMap

	sync.Mutex

	mt object.Table
}

func NewConcurrentTableSize(asize, msize int) object.Table {
	return &concurrentTable{
		a: make([]object.Value, asize),
		m: newConcurrentMapSize(msize),
	}
}

func (t *concurrentTable) Type() object.Type {
	return object.TTABLE
}

func (t *concurrentTable) String() string {
	return fmt.Sprintf("table: %p", t)
}

func (t *concurrentTable) Len() int {
	t.Lock()

	n := len(t.a)

	t.Unlock()

	return n
}

func (t *concurrentTable) Get(key object.Value) object.Value {
	return t.get(normKey(key))
}

func (t *concurrentTable) Set(key, val object.Value) {
	t.set(normKey(key), val)
}

func (t *concurrentTable) Del(key object.Value) {
	t.del(normKey(key))
}

func (t *concurrentTable) Next(key object.Value) (nkey, nval object.Value, ok bool) {
	return t.next(normKey(key))
}

func (t *concurrentTable) SetList(base int, src []object.Value) {
	t.setList(base, src)
}

func (t *concurrentTable) SetMetatable(mt object.Table) {
	t.setMetatable(mt)
}

func (t *concurrentTable) Metatable() object.Table {
	mt := t.metatable()
	return mt
}

func (t *concurrentTable) ikey(key object.Value) (object.Integer, bool) {
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

func (t *concurrentTable) get(key object.Value) object.Value {
	if ikey, ok := t.ikey(key); ok {
		return t.iget(ikey)
	}

	return t.m.Get(key)
}

func (t *concurrentTable) iget(ikey object.Integer) object.Value {
	t.Lock()

	i := int(ikey)

	if 0 < i && i <= len(t.a) {
		val := t.a[i-1]

		t.Unlock()

		return val
	}

	t.Unlock()

	return t.m.Get(ikey)
}

func (t *concurrentTable) set(key, val object.Value) {
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

func (t *concurrentTable) iset(ikey object.Integer, val object.Value) {
	t.Lock()
	defer t.Unlock()

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

func (t *concurrentTable) del(key object.Value) {
	if ikey, ok := t.ikey(key); ok {
		t.idel(ikey)
	} else {
		t.m.Delete(key)
	}
}

func (t *concurrentTable) idel(ikey object.Integer) {
	t.Lock()
	defer t.Unlock()

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

func (t *concurrentTable) next(key object.Value) (nkey, nval object.Value, ok bool) {
	t.Lock()
	defer t.Unlock()

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

func (t *concurrentTable) setList(base int, src []object.Value) {
	t.Lock()
	defer t.Unlock()

	if len(src) < len(t.a)-base {
		copy(t.a[base:], src)
	} else {
		t.a = append(t.a[base:], src...)
	}
}

func (t *concurrentTable) setMetatable(mt object.Table) {
	t.mt = mt
}

func (t *concurrentTable) metatable() object.Table {
	return t.mt
}
