package tables

import (
	"fmt"
	"sync"

	"github.com/hirochachacha/plua/object"
)

type lockedTable struct {
	t *table
	m sync.Mutex
}

func NewLockedTableSize(asize, msize int) object.Table {
	return &lockedTable{
		t: NewTableSize(asize, msize).(*table),
	}
}

func (t *lockedTable) Type() object.Type {
	return object.TTABLE
}

func (t *lockedTable) String() string {
	return fmt.Sprintf("table: %p", t)
}

func (t *lockedTable) Len() int {
	t.m.Lock()

	length := t.t.Len()

	t.m.Unlock()

	return length
}

func (t *lockedTable) Get(key object.Value) object.Value {
	t.m.Lock()

	val := t.t.Get(key)

	t.m.Unlock()

	return val
}

func (t *lockedTable) Set(key, val object.Value) {
	t.m.Lock()

	t.t.Set(key, val)

	t.m.Unlock()
}

func (t *lockedTable) Del(key object.Value) {
	t.m.Lock()

	t.t.Del(key)

	t.m.Unlock()
}

func (t *lockedTable) Next(key object.Value) (nkey, nval object.Value, ok bool) {
	t.m.Lock()

	nkey, nval, ok = t.t.Next(key)

	t.m.Unlock()

	return
}

func (t *lockedTable) Sort(less func(x, y object.Value) bool) {
	t.m.Lock()

	ts := &tableSorter{a: t.t.a[:t.t.alen], less: less}

	ts.Sort()

	t.m.Unlock()
}

func (t *lockedTable) SetList(base int, src []object.Value) {
	t.m.Lock()

	t.t.SetList(base, src)

	t.m.Unlock()
}

func (t *lockedTable) SetMetatable(mt object.Table) {
	t.m.Lock()

	t.t.SetMetatable(mt)

	t.m.Unlock()
}

func (t *lockedTable) Metatable() object.Table {
	t.m.Lock()

	mt := t.t.Metatable()

	t.m.Unlock()

	return mt
}
