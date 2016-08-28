package object

type iUserdata interface {
	GetValue() interface{}
	SetValue(interface{})
	Metatable() *Table
	SetMetatable(mt *Table)
}

func (ud *Userdata) GetValue() interface{} {
	return ud.Impl.GetValue()
}

func (ud *Userdata) SetValue(val interface{}) {
	ud.Impl.SetValue(val)
}

func (ud *Userdata) SetMetatable(mt *Table) {
	ud.Impl.SetMetatable(mt)
}

func (ud *Userdata) Metatable() *Table {
	return ud.Impl.Metatable()
}
