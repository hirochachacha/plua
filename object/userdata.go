package object

type Userdata interface {
	Value

	GetValue() interface{}
	SetValue(interface{})
	Metatable() Table
	SetMetatable(mt Table)
}
