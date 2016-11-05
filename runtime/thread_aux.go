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
