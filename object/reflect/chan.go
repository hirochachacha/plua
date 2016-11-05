package reflect

import (
	"fmt"
	"reflect"

	"github.com/hirochachacha/plua/internal/tables"
	"github.com/hirochachacha/plua/object"
	"github.com/hirochachacha/plua/object/fnutil"
)

func buildChanMT() {
	mt := tables.NewTableSize(0, 6)

	cindex := tables.NewTableSize(0, 3)

	cindex.Set(object.String("Send"), object.GoFunction(csend))
	cindex.Set(object.String("Recv"), object.GoFunction(crecv))
	cindex.Set(object.String("Close"), object.GoFunction(cclose))

	mt.Set(object.TM_METATABLE, object.True)
	mt.Set(object.TM_TOSTRING, object.GoFunction(tostring))

	mt.Set(object.TM_INDEX, cindex)
	mt.Set(object.TM_LEN, object.GoFunction(length))
	mt.Set(object.TM_PAIRS, object.GoFunction(cpairs))

	mt.Set(object.TM_EQ, cmp(func(x, y reflect.Value) bool { return x.Pointer() == y.Pointer() }))

	chanMT = mt
}

func csend(th object.Thread, args ...object.Value) ([]object.Value, *object.RuntimeError) {
	ap := fnutil.NewArgParser(th, args)

	self, err := ap.ToFullUserdata(0)
	if err != nil {
		return nil, err
	}

	x, err := ap.ToValue(1)
	if err != nil {
		return nil, err
	}

	if self, ok := self.Value.(reflect.Value); ok {
		styp := self.Type()
		vtyp := styp.Elem()

		if x := toReflectValue(vtyp, x); x.IsValid() {
			self.Send(x)

			return nil, nil
		}

		return nil, object.NewRuntimeError(fmt.Sprintf("cannot use %v (type %s) as type %s in send", x, reflect.TypeOf(x), vtyp))
	}

	return nil, errInvalidUserdata
}

func crecv(th object.Thread, args ...object.Value) ([]object.Value, *object.RuntimeError) {
	ap := fnutil.NewArgParser(th, args)

	self, err := ap.ToFullUserdata(0)
	if err != nil {
		return nil, err
	}

	if self, ok := self.Value.(reflect.Value); ok {
		rval, ok := self.Recv()
		if !ok {
			return []object.Value{nil, object.False}, nil
		}

		return []object.Value{valueOfReflect(rval, false), object.True}, nil
	}

	return nil, errInvalidUserdata
}

func cclose(th object.Thread, args ...object.Value) ([]object.Value, *object.RuntimeError) {
	ap := fnutil.NewArgParser(th, args)

	self, err := ap.ToFullUserdata(0)
	if err != nil {
		return nil, err
	}

	if self, ok := self.Value.(reflect.Value); ok {
		self.Close()

		return nil, nil
	}

	return nil, errInvalidUserdata
}

func cpairs(th object.Thread, args ...object.Value) ([]object.Value, *object.RuntimeError) {
	ap := fnutil.NewArgParser(th, args)

	self, err := ap.ToFullUserdata(0)
	if err != nil {
		return nil, err
	}

	return []object.Value{object.GoFunction(cnext), self}, nil
}

func cnext(th object.Thread, args ...object.Value) ([]object.Value, *object.RuntimeError) {
	ap := fnutil.NewArgParser(th, args)

	self, err := ap.ToFullUserdata(0)
	if err != nil {
		return nil, err
	}

	if self, ok := self.Value.(reflect.Value); ok {
		rval, ok := self.Recv()
		if !ok {
			return nil, nil
		}

		return []object.Value{object.True, valueOfReflect(rval, false)}, nil
	}

	return nil, errInvalidUserdata
}
