package object

type Table interface {
	Value

	Len() int
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
	Metatable() Table
	SetMetatable(mt Table)
}
