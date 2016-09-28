package object

type Table interface {
	Value

	Len() int
	Get(key Value) Value
	Set(Key, val Value)
	Del(key Value)
	Next(key Value) (nkey, nval Value, ok bool)
	Sort(less func(x, y Value) bool)
	SetList(base int, src []Value)
	Metatable() Table
	SetMetatable(mt Table)
}
