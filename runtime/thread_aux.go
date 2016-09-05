package runtime

import "github.com/hirochachacha/plua/object"

func (th *thread) Require(name string, open object.GoFunction) (object.Value, bool) {
	loaded := th.Loaded()

	if mod := loaded.Get(object.String(name)); object.ToGoBool(mod) {
		return mod, false
	}

	rets, err := open(th, object.String(name))
	if len(rets) == 0 || err != nil {
		return nil, false
	}

	globals := th.Globals()

	loaded.Set(object.String(name), rets[0])
	globals.Set(object.String(name), rets[0])

	return rets[0], true
}

func (th *thread) GetMetaField(val object.Value, field string) object.Value {
	mt := th.GetMetatable(val)
	if mt == nil {
		return nil
	}

	return mt.Get(object.String(field))
}

func (th *thread) CallMetaField(val object.Value, field string) (rets []object.Value, done bool) {
	if fn := th.GetMetaField(val, field); fn != nil {
		rets, _ := th.Call(fn, val)

		return rets, true
	}

	return nil, false
}
