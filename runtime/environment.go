package runtime

import (
	"github.com/hirochachacha/plua/object"
)

type environment struct {
	loaded     object.Table
	preload    object.Table
	globals    object.Table                     // default _ENV (_G)
	metatables [object.MaxType + 1]object.Table // metatable for basic type
}

func newEnvironment() *environment {
	loaded := newLockedTableSize(0, 0)
	preload := newLockedTableSize(0, 0)
	globals := newConcurrentTableSize(0, 0)

	return &environment{
		loaded:  loaded,
		preload: preload,
		globals: globals,
	}
}

func (env *environment) getMetatable(val object.Value) object.Table {
	switch val := val.(type) {
	case object.Table:
		return val.Metatable()
	case *object.Userdata:
		return val.Metatable
	}

	return env.metatables[object.ToType(val)+1]
}

func (env *environment) setMetatable(val object.Value, mt object.Table) {
	switch val := val.(type) {
	case object.Table:
		val.SetMetatable(mt)
	case *object.Userdata:
		val.Metatable = mt
	default:
		env.metatables[object.ToType(val)+1] = mt
	}
}
