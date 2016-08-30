package runtime

import (
	"github.com/hirochachacha/plua/object"
)

func (th *thread) Require(name string, open object.GoFunction) (object.Value, bool) {
	loaded := th.Loaded()

	if mod := loaded.Get(object.String(name)); object.ToGoBool(mod) {
		return mod, false
	}

	rets, err := open(th, object.String(name))
	if len(rets) == 0 || err != object.NoErr {
		return nil, false
	}

	globals := th.Globals()

	loaded.Set(object.String(name), rets[0])
	globals.Set(object.String(name), rets[0])

	return rets[0], true
}

func (th *thread) Repr(val object.Value) string {
	if rets, done := th.CallMetaField(val, "__tostring"); done {
		if len(rets) == 0 {
			return ""
		}

		return object.Repr(rets[0])
	}

	return object.Repr(val)
}

func (th *thread) NewMetatableNameSize(tname string, alen, mlen int) object.Table {
	reg := th.Registry()

	if mt := reg.Get(object.String(tname)); mt != nil {
		return nil
	}

	mt := th.NewTableSize(alen, mlen)
	mt.Set(object.String("__name"), object.String(tname))
	reg.Set(object.String(tname), mt)

	return mt
}

func (th *thread) GetMetatableName(tname string) object.Table {
	reg := th.Registry()

	mt, ok := reg.Get(object.String(tname)).(object.Table)
	if !ok {
		return nil
	}

	return mt
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
