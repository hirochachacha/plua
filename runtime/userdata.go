package runtime

import (
	"github.com/hirochachacha/blua/object"
)

type userdata struct {
	val interface{}
	mt  object.Table
}

func newUserdata(val interface{}) object.Userdata {
	return &userdata{
		val: val,
	}
}

func (ud *userdata) Type() object.Type {
	return object.TUSERDATA
}

func (ud *userdata) GetValue() interface{} {
	return ud.val
}

func (ud *userdata) SetValue(val interface{}) {
	ud.val = val
}

func (ud *userdata) SetMetatable(mt object.Table) {
	ud.mt = mt
}

func (ud *userdata) Metatable() object.Table {
	return ud.mt
}
