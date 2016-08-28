package runtime

import (
	"github.com/hirochachacha/blua/object"
)

type environment struct {
	registry   object.Table
	loaded     object.Table
	preload    object.Table
	globals    object.Table                     // default _ENV (_G)
	metatables [object.MaxType + 1]object.Table // metatable for basic type
}

func newEnvironment() *environment {
	registry := newLockedTableSize(2, 2)
	loaded := newLockedTableSize(0, 0)
	preload := newLockedTableSize(0, 0)
	globals := newConcurrentTableSize(0, 0)

	registry.ISet(2, globals)
	registry.Set(object.String("_LOADED"), loaded)
	registry.Set(object.String("_PRELOAD"), preload)

	return &environment{
		registry: registry,
		loaded:   loaded,
		preload:  preload,
		globals:  globals,
	}
}

func (env *environment) getMetatable(val object.Value) object.Table {
	switch val := val.(type) {
	case object.Table:
		return val.Metatable()
	case object.Userdata:
		return val.Metatable()
	}

	return env.metatables[object.ToType(val)+1]
}

func (env *environment) setMetatable(val object.Value, mt object.Table) {
	switch val := val.(type) {
	case object.Table:
		val.SetMetatable(mt)
	case object.Userdata:
		val.SetMetatable(mt)
	default:
		env.metatables[object.ToType(val)+1] = mt
	}
}
