package tables

import (
	"fmt"
	"sync"

	"github.com/hirochachacha/plua/internal/limits"
	"github.com/hirochachacha/plua/object"
)

type concurrentTable struct {
	alen int
	a    []object.Value
	m    *concurrentMap

	sync.Mutex

	mt object.Table
}

func NewConcurrentTableSize(asize, msize int) object.Table {
	return &concurrentTable{
		alen: 0,
		a:    make([]object.Value, asize),
		m:    newConcurrentMapSize(msize),
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

	alen := t.alen

	t.Unlock()

	return int(alen)
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

func (t *concurrentTable) Get(key object.Value) object.Value {
	return t.get(key)
}

func (t *concurrentTable) Set(key, val object.Value) {
	t.set(key, val)
}

func (t *concurrentTable) Del(key object.Value) {
	t.del(key)
}

func (t *concurrentTable) IGet(i int) object.Value {
	return t.iget(i)
}

func (t *concurrentTable) ISet(i int, val object.Value) {
	t.iset(i, val)
}

func (t *concurrentTable) IDel(i int) {
	t.idel(i)
}

func (t *concurrentTable) Next(key object.Value) (nkey, nval object.Value, ok bool) {
	return t.next(key)
}

func (t *concurrentTable) INext(i int) (ni int, nval object.Value, ok bool) {
	return t.inext(i)
}

func (t *concurrentTable) Sort(less func(x, y object.Value) bool) {
	t.Lock()

	ts := &tableSorter{a: t.a[:t.alen], less: less}

	ts.Sort()

	t.Unlock()
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

func (t *concurrentTable) get(key object.Value) object.Value {
	if ikey, ok := t.ikey(key); ok {
		return t.iget(int(ikey))
	}

	return t.m.Get(key)
}

func (t *concurrentTable) set(key, val object.Value) {
	if ikey, ok := t.ikey(key); ok {
		t.iset(int(ikey), val)
	} else {
		if val == nil {
			t.m.Delete(key)
		} else {
			t.m.Set(key, val)
		}
	}
}

func (t *concurrentTable) del(key object.Value) {
	if ikey, ok := t.ikey(key); ok {
		t.idel(int(ikey))
	} else {
		t.m.Delete(key)
	}
}

func (t *concurrentTable) iget(i int) object.Value {
	t.Lock()
	defer t.Unlock()

	if 0 < i && i <= len(t.a) {
		return t.a[i-1]
	}

	return t.m.Get(object.Integer(i))
}

func (t *concurrentTable) iset(i int, val object.Value) {
	if val == nil {
		t.Lock()

		acap := len(t.a)
		alen := t.alen

		switch {
		case 0 < i && i <= acap:
			if alen > i-1 {
				t.alen = i - 1
			}
			t.a[i-1] = nil
		case i == acap+1:
			// do nothing
		default:
			t.m.Delete(object.Integer(i))
		}

		t.Unlock()
	} else {
		t.Lock()

		acap := len(t.a)

		switch {
		case 0 < i && i <= acap:
			alen := t.alen
			if t.a[i-1] == nil && alen == i {
				for j := i; j < acap; j++ {
					alen++
					if t.a[j] == nil {
						break
					}
				}
				t.alen = alen
			}

			t.a[i-1] = val
		case i == acap+1:
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

		t.Unlock()
	}
}

func (t *concurrentTable) idel(i int) {
	t.Lock()

	acap := len(t.a)

	switch {
	case 0 < i && i <= acap:
		t.a[i-1] = nil

		copy(t.a[i-1:], t.a[i:])

		t.alen--
	case i == acap+1:
		// do nothing
	default:
		t.m.Delete(object.Integer(i))
	}

	t.Unlock()
}

func (t *concurrentTable) next(key object.Value) (nkey, nval object.Value, ok bool) {
	t.Lock()
	defer t.Unlock()

	alen := t.alen

	if key == nil {
		if alen > 0 {
			return object.Integer(1), t.a[0], true
		}

		return t.m.Next(nil)
	}

	if ikey, ok := t.ikey(key); ok {
		if ikey < object.Integer(alen) {
			return ikey + 1, t.a[int(ikey)], true
		}
		if ikey == object.Integer(alen) {
			return t.m.Next(nil)
		}
	}

	return t.m.Next(key)
}

func (t *concurrentTable) inext(i int) (ni int, nval object.Value, ok bool) {
	t.Lock()
	defer t.Unlock()

	alen := t.alen

	if alen <= i {
		return -1, nil, false
	}

	return i + 1, t.a[i], true
}

func (t *concurrentTable) setList(base int, src []object.Value) {
	t.Lock()
	defer t.Unlock()

	if len(src) < len(t.a)-base {
		copy(t.a[base:], src)
	} else {
		t.a = append(t.a[base:], src...)
	}

	if t.alen < base+len(src) {
		t.alen = base + len(src)
	}
}

func (t *concurrentTable) setMetatable(mt object.Table) {
	t.mt = mt
}

func (t *concurrentTable) metatable() object.Table {
	return t.mt
}
