package object

type iTable interface {
	Len() int
	ALen() int
	ACap() int
	MLen() int
	MCap() int
	Get(key Value) Value
	Set(Key, val Value)
	Del(key Value)
	IGet(i int) Value
	ISet(i int, val Value)
	IDel(i int)
	Next(key Value) (nkey, nval Value, ok bool)
	INext(i int) (ni int, nval Value, ok bool)
	Sort(less func(x, y Value) bool)
	SetList(base int, src []Value)
	Metatable() *Table
	SetMetatable(mt *Table)
}

func (t *Table) Len() int {
	return t.Impl.Len()
}

func (t *Table) ALen() int {
	return t.Impl.ALen()
}

func (t *Table) ACap() int {
	return t.Impl.ACap()
}

func (t *Table) MLen() int {
	return t.Impl.MLen()
}

func (t *Table) MCap() int {
	return t.Impl.MCap()
}

func (t *Table) Get(key Value) Value {
	return t.Impl.Get(key)
}

func (t *Table) Set(key, val Value) {
	t.Impl.Set(key, val)
}

func (t *Table) Del(key Value) {
	t.Impl.Del(key)
}

func (t *Table) IGet(i int) Value {
	return t.Impl.IGet(i)
}

func (t *Table) ISet(i int, val Value) {
	t.Impl.ISet(i, val)
}

func (t *Table) IDel(i int) {
	t.Impl.IDel(i)
}

func (t *Table) Next(key Value) (nkey, nval Value, ok bool) {
	return t.Impl.Next(key)
}

func (t *Table) INext(i int) (ni int, nval Value, ok bool) {
	return t.Impl.INext(i)
}

func (t *Table) Sort(less func(x, y Value) bool) {
	t.Impl.Sort(less)
}

func (t *Table) SetList(base int, src []Value) {
	t.Impl.SetList(base, src)
}

func (t *Table) SetMetatable(mt *Table) {
	t.Impl.SetMetatable(mt)
}

func (t *Table) Metatable() *Table {
	return t.Impl.Metatable()
}
