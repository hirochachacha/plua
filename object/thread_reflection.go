package object

import (
	"fmt"
	"math"
	"reflect"

	"unicode"
	"unicode/utf8"

	"github.com/hirochachacha/blua/internal/limits"
)

type rvalue reflect.Value // encapsulation

var (
	invalid = reflect.ValueOf(nil)
)

var (
	boolMT    = "BoolMT"
	intMT     = "IntMT"
	floatMT   = "FloatMT"
	stringMT  = "StringMT"
	uintMT    = "UintMT"
	complexMT = "ComplexMT"
	arrayMT   = "ArrayMT"
	sliceMT   = "SliceMT"
	chanMT    = "ChanMT"
	funcMT    = "FuncMT"
	ifaceMT   = "InterfaceMT"
	mapMT     = "MapMT"
	ptrMT     = "PtrMT"
	structMT  = "StructMT"
	uPtrMT    = "UnsafePointerMT"
)

var primitives = []reflect.Type{
	reflect.Bool:    reflect.TypeOf(false),
	reflect.Int:     reflect.TypeOf(int(0)),
	reflect.Int8:    reflect.TypeOf(int8(0)),
	reflect.Int16:   reflect.TypeOf(int16(0)),
	reflect.Int32:   reflect.TypeOf(int32(0)),
	reflect.Int64:   reflect.TypeOf(int64(0)),
	reflect.Float32: reflect.TypeOf(float32(0)),
	reflect.Float64: reflect.TypeOf(float64(0)),
	reflect.String:  reflect.TypeOf(""),
}

// Interface{} (Go) -> reflect.Value (Go)
// reflect.ValueOf(interface{}) reflect.Value

// interface{} (Go) -> Value (Lua)
func (th *Thread) ValueOf(x interface{}) Value {
	if val, ok := ValueOf(x); ok {
		return val
	}

	if !th.hasReflection {
		th.buildReflections()
		th.hasReflection = true
	}

	return th.valueOfReflect(reflect.ValueOf(x))
}

// reflect.Value (Go) -> Value (Lua)
func (th *Thread) valueOfReflect(rval reflect.Value) Value {
	var ud *Userdata

	switch kind := rval.Kind(); kind {
	case reflect.Bool:
		if rval.Type() == primitives[kind] {
			return Boolean(rval.Bool())
		}
		ud = th.NewUserdata(rvalue(rval))
		ud.SetMetatable(th.GetMetatableName(boolMT))
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		if rval.Type() == primitives[kind] {
			return Integer(rval.Int())
		}
		ud = th.NewUserdata(rvalue(rval))
		ud.SetMetatable(th.GetMetatableName(intMT))
	case reflect.Float32, reflect.Float64:
		if rval.Type() == primitives[kind] {
			return Number(rval.Float())
		}
		ud = th.NewUserdata(rvalue(rval))
		ud.SetMetatable(th.GetMetatableName(floatMT))
	case reflect.String:
		if rval.Type() == primitives[kind] {
			return String(rval.String())
		}
		ud = th.NewUserdata(rvalue(rval))
		ud.SetMetatable(th.GetMetatableName(stringMT))
	case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64, reflect.Uintptr:
		ud = th.NewUserdata(rvalue(rval))
		ud.SetMetatable(th.GetMetatableName(uintMT))
	case reflect.Complex64, reflect.Complex128:
		ud = th.NewUserdata(rvalue(rval))
		ud.SetMetatable(th.GetMetatableName(complexMT))
	case reflect.Array:
		ud = th.NewUserdata(rvalue(rval))
		ud.SetMetatable(th.GetMetatableName(arrayMT))
	case reflect.Chan:
		if rval.IsNil() {
			return nil
		}
		ud = th.NewUserdata(rvalue(rval))
		ud.SetMetatable(th.GetMetatableName(chanMT))
	case reflect.Func:
		if rval.IsNil() {
			return nil
		}
		ud = th.NewUserdata(rvalue(rval))
		ud.SetMetatable(th.GetMetatableName(funcMT))
	case reflect.Interface:
		if rval.IsNil() {
			return nil
		}
		ud = th.NewUserdata(rvalue(rval))
		ud.SetMetatable(th.GetMetatableName(ifaceMT))
	case reflect.Map:
		if rval.IsNil() {
			return nil
		}
		ud = th.NewUserdata(rvalue(rval))
		ud.SetMetatable(th.GetMetatableName(mapMT))
	case reflect.Ptr:
		if rval.IsNil() {
			return nil
		}
		ud = th.NewUserdata(rvalue(rval))
		ud.SetMetatable(th.GetMetatableName(ptrMT))
	case reflect.Slice:
		if rval.IsNil() {
			return nil
		}
		ud = th.NewUserdata(rvalue(rval))
		ud.SetMetatable(th.GetMetatableName(sliceMT))
	case reflect.Struct:
		ud = th.NewUserdata(rvalue(rval))
		ud.SetMetatable(th.GetMetatableName(structMT))
	case reflect.UnsafePointer:
		ud = th.NewUserdata(rvalue(rval))
		ud.SetMetatable(th.GetMetatableName(uPtrMT))
	case reflect.Invalid:
		panic("unexpected")
	default:
		panic("unreachable")
	}

	return ud
}

// Value (Lua) -> reflect.Value (Go)
func toReflectValue(typ reflect.Type, val Value) reflect.Value {
	switch val := val.(type) {
	case nil:
		switch typ.Kind() {
		case reflect.Chan, reflect.Func, reflect.Map, reflect.Ptr, reflect.Interface, reflect.Slice:
			return reflect.Zero(typ)
		}
	case Boolean:
		if typ.Kind() == reflect.Bool {
			return reflect.ValueOf(val).Convert(typ)
		}
	case Integer:
		switch typ.Kind() {
		case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64,
			reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64, reflect.Uintptr,
			reflect.Float32, reflect.Float64:
			return reflect.ValueOf(val).Convert(typ)
		case reflect.String:
			return reflect.ValueOf(integerToString(val)).Convert(typ)
		}
	case Number:
		switch typ.Kind() {
		case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
			if ival, ok := numberToInteger(val); ok {
				return reflect.ValueOf(ival).Convert(typ)
			}
		case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64, reflect.Uintptr:
			if u, ok := numberToGoUint(val); ok {
				return reflect.ValueOf(u).Convert(typ)
			}
		case reflect.Float32, reflect.Float64:
			return reflect.ValueOf(val).Convert(typ)
		case reflect.String:
			return reflect.ValueOf(numberToString(val)).Convert(typ)
		}
	case String:
		switch typ.Kind() {
		case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
			if ival, ok := stringToInteger(val); ok {
				return reflect.ValueOf(ival).Convert(typ)
			}
		case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64, reflect.Uintptr:
			if u, ok := stringToGoUint(val); ok {
				return reflect.ValueOf(u).Convert(typ)
			}
		case reflect.Float32, reflect.Float64:
			if nval, ok := stringToNumber(val); ok {
				return reflect.ValueOf(nval).Convert(typ)
			}
		case reflect.String:
			return reflect.ValueOf(val).Convert(typ)
		}
	case LightUserdata:
		if typ.Kind() == reflect.UnsafePointer {
			return reflect.ValueOf(val.Pointer).Convert(typ)
		}
	case GoFunction:
		rval := reflect.ValueOf(val)
		rtyp := rval.Type()

		if rtyp == typ {
			return rval
		}
	case *Table:
		rval := reflect.ValueOf(val)
		rtyp := rval.Type()

		if rtyp == typ {
			return rval
		}
	case *Closure:
		rval := reflect.ValueOf(val)
		rtyp := rval.Type()

		if rtyp == typ {
			return rval
		}
	case *Thread:
		rval := reflect.ValueOf(val)
		rtyp := rval.Type()

		if rtyp == typ {
			return rval
		}
	case *Channel:
		rval := reflect.ValueOf(val)
		rtyp := rval.Type()

		if rtyp == typ {
			return rval
		}
	case *Userdata:
		if rval, ok := val.GetValue().(rvalue); ok {
			rval := reflect.Value(rval)
			rtyp := rval.Type()

			if rtyp == typ {
				return rval
			}
		} else {
			rval := reflect.ValueOf(val)
			rtyp := rval.Type()

			if rtyp == typ {
				return rval
			}
		}
	default:
		panic("unreachable")
	}

	return invalid
}

func (th *Thread) buildReflections() {
	th.buildBoolMT()
	th.buildIntMT()
	th.buildFloatMT()
	th.buildStringMT()
	th.buildUintMT()
	th.buildComplexMT()
	th.buildArrayMT()
	th.buildSliceMT()
	th.buildChanMT()
	th.buildPtrMT()
	th.buildStructMT()
	th.buildUnsafePointerMT()
	th.buildFuncMT()
	th.buildInterfaceMT()
	th.buildMapMT()
}

func (th *Thread) buildBoolMT() {
	mt := th.NewMetatableNameSize(boolMT, 0, 4)

	mt.Set(String("__metatable"), True)
	mt.Set(String("__tostring"), GoFunction(tostring))

	mt.Set(String("__index"), GoFunction(index))

	mt.Set(String("__eq"), cmp(func(x, y reflect.Value) bool { return x.Bool() == y.Bool() }))
}

func (th *Thread) buildIntMT() {
	mt := th.NewMetatableNameSize(intMT, 0, 20)

	mt.Set(String("__metatable"), True)
	mt.Set(String("__tostring"), GoFunction(tostring))

	mt.Set(String("__index"), GoFunction(index))

	mt.Set(String("__eq"), cmp(func(x, y reflect.Value) bool { return x.Int() == y.Int() }))
	mt.Set(String("__lt"), cmp(func(x, y reflect.Value) bool { return x.Int() < y.Int() }))
	mt.Set(String("__le"), cmp(func(x, y reflect.Value) bool { return x.Int() <= y.Int() }))

	mt.Set(String("__unm"), unary(func(x reflect.Value) reflect.Value { return reflect.ValueOf(-x.Int()) }, mt))
	mt.Set(String("__bnot"), unary(func(x reflect.Value) reflect.Value { return reflect.ValueOf(^x.Int()) }, mt))

	mt.Set(String("__add"), binary(func(x, y reflect.Value) reflect.Value { return reflect.ValueOf(x.Int() + y.Int()) }, mt))
	mt.Set(String("__sub"), binary(func(x, y reflect.Value) reflect.Value { return reflect.ValueOf(x.Int() - y.Int()) }, mt))
	mt.Set(String("__mul"), binary(func(x, y reflect.Value) reflect.Value { return reflect.ValueOf(x.Int() * y.Int()) }, mt))
	mt.Set(String("__mod"), binary(func(x, y reflect.Value) reflect.Value { return reflect.ValueOf(th.imod(x.Int(), y.Int())) }, mt))
	mt.Set(String("__pow"), binary(func(x, y reflect.Value) reflect.Value { return reflect.ValueOf(ipow(x.Int(), y.Int())) }, mt))
	mt.Set(String("__div"), binary(func(x, y reflect.Value) reflect.Value { return reflect.ValueOf(th.idiv(x.Int(), y.Int())) }, mt))
	mt.Set(String("__idiv"), binary(func(x, y reflect.Value) reflect.Value { return reflect.ValueOf(th.idiv(x.Int(), y.Int())) }, mt))
	mt.Set(String("__band"), binary(func(x, y reflect.Value) reflect.Value { return reflect.ValueOf(x.Int() & y.Int()) }, mt))
	mt.Set(String("__bor"), binary(func(x, y reflect.Value) reflect.Value { return reflect.ValueOf(x.Int() | y.Int()) }, mt))
	mt.Set(String("__bxor"), binary(func(x, y reflect.Value) reflect.Value { return reflect.ValueOf(x.Int() ^ y.Int()) }, mt))
	mt.Set(String("__shl"), binary(func(x, y reflect.Value) reflect.Value { return reflect.ValueOf(ishl(x.Int(), y.Int())) }, mt))
	mt.Set(String("__shr"), binary(func(x, y reflect.Value) reflect.Value { return reflect.ValueOf(ishr(x.Int(), y.Int())) }, mt))
}

func (th *Thread) buildFloatMT() {
	mt := th.NewMetatableNameSize(floatMT, 0, 14)

	mt.Set(String("__metatable"), True)
	mt.Set(String("__tostring"), GoFunction(tostring))

	mt.Set(String("__index"), GoFunction(index))

	mt.Set(String("__eq"), cmp(func(x, y reflect.Value) bool { return x.Float() == y.Float() }))
	mt.Set(String("__lt"), cmp(func(x, y reflect.Value) bool { return x.Float() < y.Float() }))
	mt.Set(String("__le"), cmp(func(x, y reflect.Value) bool { return x.Float() <= y.Float() }))

	mt.Set(String("__unm"), unary(func(x reflect.Value) reflect.Value { return reflect.ValueOf(-x.Float()) }, mt))

	mt.Set(String("__add"), binary(func(x, y reflect.Value) reflect.Value { return reflect.ValueOf(x.Float() + y.Float()) }, mt))
	mt.Set(String("__sub"), binary(func(x, y reflect.Value) reflect.Value { return reflect.ValueOf(x.Float() - y.Float()) }, mt))
	mt.Set(String("__mul"), binary(func(x, y reflect.Value) reflect.Value { return reflect.ValueOf(x.Float() * y.Float()) }, mt))
	mt.Set(String("__mod"), binary(func(x, y reflect.Value) reflect.Value { return reflect.ValueOf(th.fmod(x.Float(), y.Float())) }, mt))
	mt.Set(String("__pow"), binary(func(x, y reflect.Value) reflect.Value { return reflect.ValueOf(math.Pow(x.Float(), y.Float())) }, mt))
	mt.Set(String("__div"), binary(func(x, y reflect.Value) reflect.Value { return reflect.ValueOf(x.Float() / y.Float()) }, mt))
	mt.Set(String("__idiv"), binary(func(x, y reflect.Value) reflect.Value { return reflect.ValueOf(th.fidiv(x.Float(), y.Float())) }, mt))
}

func (th *Thread) buildStringMT() {
	mt := th.NewMetatableNameSize(stringMT, 0, 7)

	mt.Set(String("__metatable"), True)
	mt.Set(String("__tostring"), GoFunction(tostring))

	mt.Set(String("__index"), GoFunction(index))

	mt.Set(String("__eq"), cmp(func(x, y reflect.Value) bool { return x.String() == y.String() }))
	mt.Set(String("__lt"), cmp(func(x, y reflect.Value) bool { return x.String() < y.String() }))
	mt.Set(String("__le"), cmp(func(x, y reflect.Value) bool { return x.String() <= y.String() }))

	mt.Set(String("__concat"), binary(func(x, y reflect.Value) reflect.Value { return reflect.ValueOf(x.String() + y.String()) }, mt))
}

// uintptr also supportted
func (th *Thread) buildUintMT() {
	mt := th.NewMetatableNameSize(uintMT, 0, 20)

	mt.Set(String("__metatable"), True)
	mt.Set(String("__tostring"), GoFunction(tostring))

	mt.Set(String("__index"), GoFunction(index))

	mt.Set(String("__eq"), cmp(func(x, y reflect.Value) bool { return x.Uint() == y.Uint() }))
	mt.Set(String("__lt"), cmp(func(x, y reflect.Value) bool { return x.Uint() < y.Uint() }))
	mt.Set(String("__le"), cmp(func(x, y reflect.Value) bool { return x.Uint() <= y.Uint() }))

	mt.Set(String("__unm"), unary(func(x reflect.Value) reflect.Value { return reflect.ValueOf(-x.Uint()) }, mt))
	mt.Set(String("__bnot"), unary(func(x reflect.Value) reflect.Value { return reflect.ValueOf(^x.Uint()) }, mt))

	mt.Set(String("__add"), binary(func(x, y reflect.Value) reflect.Value { return reflect.ValueOf(x.Uint() + y.Uint()) }, mt))
	mt.Set(String("__sub"), binary(func(x, y reflect.Value) reflect.Value { return reflect.ValueOf(x.Uint() - y.Uint()) }, mt))
	mt.Set(String("__mul"), binary(func(x, y reflect.Value) reflect.Value { return reflect.ValueOf(x.Uint() * y.Uint()) }, mt))
	mt.Set(String("__mod"), binary(func(x, y reflect.Value) reflect.Value { return reflect.ValueOf(th.umod(x.Uint(), y.Uint())) }, mt))
	mt.Set(String("__pow"), binary(func(x, y reflect.Value) reflect.Value { return reflect.ValueOf(upow(x.Uint(), y.Uint())) }, mt))
	mt.Set(String("__div"), binary(func(x, y reflect.Value) reflect.Value { return reflect.ValueOf(th.udiv(x.Uint(), y.Uint())) }, mt))
	mt.Set(String("__idiv"), binary(func(x, y reflect.Value) reflect.Value { return reflect.ValueOf(th.udiv(x.Uint(), y.Uint())) }, mt))
	mt.Set(String("__band"), binary(func(x, y reflect.Value) reflect.Value { return reflect.ValueOf(x.Uint() & y.Uint()) }, mt))
	mt.Set(String("__bor"), binary(func(x, y reflect.Value) reflect.Value { return reflect.ValueOf(x.Uint() | y.Uint()) }, mt))
	mt.Set(String("__bxor"), binary(func(x, y reflect.Value) reflect.Value { return reflect.ValueOf(x.Uint() ^ y.Uint()) }, mt))
	mt.Set(String("__shl"), binary(func(x, y reflect.Value) reflect.Value { return reflect.ValueOf(x.Uint() << y.Uint()) }, mt))
	mt.Set(String("__shr"), binary(func(x, y reflect.Value) reflect.Value { return reflect.ValueOf(x.Uint() >> y.Uint()) }, mt))
}

func (th *Thread) buildComplexMT() {
	mt := th.NewMetatableNameSize(complexMT, 0, 9)

	mt.Set(String("__metatable"), True)
	mt.Set(String("__tostring"), GoFunction(tostring))

	mt.Set(String("__index"), GoFunction(index))

	mt.Set(String("__eq"), cmp(func(x, y reflect.Value) bool { return x.Complex() == y.Complex() }))

	mt.Set(String("__unm"), unary(func(x reflect.Value) reflect.Value { return reflect.ValueOf(-x.Complex()) }, mt))

	mt.Set(String("__add"), binary(func(x, y reflect.Value) reflect.Value { return reflect.ValueOf(x.Complex() + y.Complex()) }, mt))
	mt.Set(String("__sub"), binary(func(x, y reflect.Value) reflect.Value { return reflect.ValueOf(x.Complex() - y.Complex()) }, mt))
	mt.Set(String("__mul"), binary(func(x, y reflect.Value) reflect.Value { return reflect.ValueOf(x.Complex() * y.Complex()) }, mt))
	mt.Set(String("__div"), binary(func(x, y reflect.Value) reflect.Value { return reflect.ValueOf(x.Complex() / y.Complex()) }, mt))
}

func (th *Thread) buildArrayMT() {
	mt := th.NewMetatableNameSize(arrayMT, 0, 7)

	mt.Set(String("__metatable"), True)
	mt.Set(String("__tostring"), GoFunction(tostring))

	mt.Set(String("__index"), GoFunction(arrayIndex))
	mt.Set(String("__newindex"), GoFunction(arrayNewIndex))
	mt.Set(String("__len"), GoFunction(length))
	mt.Set(String("__pairs"), GoFunction(arrayPairs))

	mt.Set(String("__eq"), cmp(func(x, y reflect.Value) bool { return aeq(x, y) }))
}

func (th *Thread) buildSliceMT() {
	mt := th.NewMetatableNameSize(sliceMT, 0, 7)

	mt.Set(String("__metatable"), True)
	mt.Set(String("__tostring"), GoFunction(tostring))

	mt.Set(String("__index"), GoFunction(arrayIndex))
	mt.Set(String("__newindex"), GoFunction(arrayNewIndex))
	mt.Set(String("__len"), GoFunction(length))
	mt.Set(String("__pairs"), GoFunction(arrayPairs))

	mt.Set(String("__eq"), cmp(func(x, y reflect.Value) bool { return x.Pointer() == y.Pointer() }))
}

func (th *Thread) buildChanMT() {
	mt := th.NewMetatableNameSize(chanMT, 0, 6)

	chanIndex := th.NewTableSize(0, 3)

	chanIndex.Set(String("Send"), GoFunction(chanSend))
	chanIndex.Set(String("Recv"), GoFunction(chanRecv))
	chanIndex.Set(String("Close"), GoFunction(chanClose))

	mt.Set(String("__metatable"), True)
	mt.Set(String("__tostring"), GoFunction(tostring))

	mt.Set(String("__index"), chanIndex)
	mt.Set(String("__len"), GoFunction(length))
	mt.Set(String("__pairs"), GoFunction(chanPairs))

	mt.Set(String("__eq"), cmp(func(x, y reflect.Value) bool { return x.Pointer() == y.Pointer() }))
}

func (th *Thread) buildPtrMT() {
	mt := th.NewMetatableNameSize(ptrMT, 0, 5)

	mt.Set(String("__metatable"), True)
	mt.Set(String("__tostring"), GoFunction(tostring))

	mt.Set(String("__index"), GoFunction(ptrIndex))
	mt.Set(String("__newindex"), GoFunction(ptrNewIndex))

	mt.Set(String("__eq"), cmp(func(x, y reflect.Value) bool { return x.Pointer() == y.Pointer() }))
}

func (th *Thread) buildStructMT() {
	mt := th.NewMetatableNameSize(structMT, 0, 5)

	mt.Set(String("__metatable"), True)
	mt.Set(String("__tostring"), GoFunction(tostring))

	mt.Set(String("__index"), GoFunction(structIndex))
	mt.Set(String("__newindex"), GoFunction(structNewIndex))

	mt.Set(String("__eq"), cmp(func(x, y reflect.Value) bool { return x.Interface() == y.Interface() }))
}

func (th *Thread) buildUnsafePointerMT() {
	mt := th.NewMetatableNameSize(uPtrMT, 0, 4)

	mt.Set(String("__metatable"), True)
	mt.Set(String("__tostring"), GoFunction(tostring))

	mt.Set(String("__index"), GoFunction(index))

	mt.Set(String("__eq"), cmp(func(x, y reflect.Value) bool { return x.Pointer() == y.Pointer() }))
}

func (th *Thread) buildFuncMT() {
	mt := th.NewMetatableNameSize(funcMT, 0, 5)

	mt.Set(String("__metatable"), True)
	mt.Set(String("__tostring"), GoFunction(tostring))

	mt.Set(String("__index"), GoFunction(index))

	mt.Set(String("__call"), GoFunction(call))

	mt.Set(String("__eq"), cmp(func(x, y reflect.Value) bool { return x.Pointer() == y.Pointer() }))
}

func (th *Thread) buildInterfaceMT() {
	mt := th.NewMetatableNameSize(ifaceMT, 0, 4)

	mt.Set(String("__metatable"), True)
	mt.Set(String("__tostring"), GoFunction(tostring))

	mt.Set(String("__index"), GoFunction(ifaceIndex))

	mt.Set(String("__eq"), cmp(func(x, y reflect.Value) bool { return x.Interface() == y.Interface() }))
}

func (th *Thread) buildMapMT() {
	mt := th.NewMetatableNameSize(mapMT, 0, 7)

	mt.Set(String("__metatable"), True)
	mt.Set(String("__tostring"), GoFunction(tostring))

	mt.Set(String("__index"), GoFunction(mapIndex))
	mt.Set(String("__newindex"), GoFunction(mapNewIndex))
	mt.Set(String("__len"), GoFunction(length))
	mt.Set(String("__pairs"), GoFunction(mapPairs))

	mt.Set(String("__eq"), cmp(func(x, y reflect.Value) bool { return x.Pointer() == y.Pointer() }))
}

func length(th *Thread, args ...Value) []Value {
	ac := th.NewArgChecker(args)

	self := ac.ToFullUserdata(0)

	if !ac.OK() {
		return nil
	}

	if self, ok := self.GetValue().(rvalue); ok {
		self := reflect.Value(self)

		return []Value{Integer(self.Len())}
	}

	ac.Error("invalid userdata")

	return nil
}

func arrayIndex(th *Thread, args ...Value) []Value {
	ac := th.NewArgChecker(args)

	self := ac.ToFullUserdata(0)
	index := ac.ToGoInt(1)

	if !ac.OK() {
		return nil
	}

	index--

	if self, ok := self.GetValue().(rvalue); ok {
		self := reflect.Value(self)

		if 0 <= index && index < self.Len() {
			rval := self.Index(index)

			return []Value{th.valueOfReflect(rval)}
		}

		ac.Error(fmt.Sprintf("invalid array index %d (out of bounds for %d-element array)", index, self.Len()))
	}

	ac.Error("invalid userdata")

	return nil
}

func arrayNewIndex(th *Thread, args ...Value) []Value {
	ac := th.NewArgChecker(args)

	self := ac.ToFullUserdata(0)
	index := ac.ToGoInt(1)
	val := ac.Get(2)

	if !ac.OK() {
		return nil
	}

	index--

	if self, ok := self.GetValue().(rvalue); ok {
		self := reflect.Value(self)

		if 0 <= index && index < self.Len() {
			styp := self.Type()
			vtyp := styp.Elem()

			if rval := toReflectValue(vtyp, val); rval.IsValid() {
				self.Index(index).Set(rval)

				return nil
			}

			ac.Error(fmt.Sprintf("non-%s array index \"%s\"", vtyp, reflect.TypeOf(val)))
		}

		ac.Error(fmt.Sprintf("invalid array index %d (out of bounds for %d-element array)", index, self.Len()))
	}

	ac.Error("invalid userdata")

	return nil
}

func arrayPairs(th *Thread, args ...Value) []Value {
	ac := th.NewArgChecker(args)

	self := ac.ToFullUserdata(0)

	if !ac.OK() {
		return nil
	}

	if _, ok := self.GetValue().(rvalue); ok {
		return []Value{GoFunction(inext), ac.Get(1), Integer(0)}
	}

	ac.Error("invalid userdata")

	return nil
}

func inext(th *Thread, args ...Value) []Value {
	ac := th.NewArgChecker(args)

	self := ac.ToFullUserdata(0)
	index := ac.ToGoInt(1)

	if !ac.OK() {
		return nil
	}

	if self, ok := self.GetValue().(rvalue); ok {
		self := reflect.Value(self)

		if index >= self.Len() {
			return nil
		}

		rval := self.Index(index)

		index++

		return []Value{Integer(index), th.valueOfReflect(rval)}
	}

	ac.Error("invalid userdata")

	return nil
}

func mapIndex(th *Thread, args ...Value) []Value {
	ac := th.NewArgChecker(args)

	self := ac.ToFullUserdata(0)
	key := ac.Get(1)

	if !ac.OK() {
		return nil
	}

	if self, ok := self.GetValue().(rvalue); ok {
		self := reflect.Value(self)

		ktyp := self.Type().Key()

		if rkey := toReflectValue(ktyp, key); rkey.IsValid() {
			rval := self.MapIndex(rkey)

			return []Value{th.valueOfReflect(rval)}
		}

		ac.Error(fmt.Sprintf("cannot use %v (type %s) as type %s in map index", key, reflect.TypeOf(key), ktyp))
	}

	ac.Error("invalid userdata")

	return nil
}

func mapNewIndex(th *Thread, args ...Value) []Value {
	ac := th.NewArgChecker(args)

	self := ac.ToFullUserdata(0)
	key := ac.Get(1)
	val := ac.Get(2)

	if !ac.OK() {
		return nil
	}

	if self, ok := self.GetValue().(rvalue); ok {
		self := reflect.Value(self)

		styp := self.Type()
		ktyp := styp.Key()
		vtyp := styp.Elem()

		if rkey := toReflectValue(ktyp, key); rkey.IsValid() {
			if rval := toReflectValue(vtyp, val); rval.IsValid() {
				self.SetMapIndex(rkey, rval)

				return nil
			}

			ac.Error(fmt.Sprintf("cannot use %v (type %s) as type %s in map assignment", val, reflect.TypeOf(val), vtyp))
		}

		ac.Error(fmt.Sprintf("cannot use %v (type %s) as type %s in map index", key, reflect.TypeOf(key), ktyp))
	}

	ac.Error("invalid userdata")

	return nil
}

func mapPairs(th *Thread, args ...Value) []Value {
	ac := th.NewArgChecker(args)

	self := ac.ToFullUserdata(0)

	if !ac.OK() {
		return nil
	}

	if self, ok := self.GetValue().(rvalue); ok {
		self := reflect.Value(self)

		keys := self.MapKeys()
		length := len(keys)

		i := 0

		next := func(th *Thread, args ...Value) []Value {
			if i == length {
				return nil
			}

			key := keys[i]
			rval := self.MapIndex(key)

			i++

			return []Value{th.valueOfReflect(key), th.valueOfReflect(rval)}
		}

		return []Value{GoFunction(next)}
	}

	ac.Error("invalid userdata")

	return nil
}

func chanSend(th *Thread, args ...Value) []Value {
	ac := th.NewArgChecker(args)

	self := ac.ToFullUserdata(0)
	x := ac.Get(1)

	if !ac.OK() {
		return nil
	}

	if self, ok := self.GetValue().(rvalue); ok {
		self := reflect.Value(self)

		styp := self.Type()
		vtyp := styp.Elem()

		if x := toReflectValue(vtyp, x); x.IsValid() {
			self.Send(x)

			return nil
		}

		ac.Error(fmt.Sprintf("cannot use %v (type %s) as type %s in send", x, reflect.TypeOf(x), vtyp))
	}

	ac.Error("invalid userdata")

	return nil
}

func chanRecv(th *Thread, args ...Value) []Value {
	ac := th.NewArgChecker(args)

	self := ac.ToFullUserdata(0)

	if !ac.OK() {
		return nil
	}

	if self, ok := self.GetValue().(rvalue); ok {
		self := reflect.Value(self)

		rval, ok := self.Recv()
		if !ok {
			return []Value{nil, False}
		}

		return []Value{th.valueOfReflect(rval), True}
	}

	ac.Error("invalid userdata")

	return nil
}

func chanClose(th *Thread, args ...Value) []Value {
	ac := th.NewArgChecker(args)

	self := ac.ToFullUserdata(0)

	if !ac.OK() {
		return nil
	}

	if self, ok := self.GetValue().(rvalue); ok {
		self := reflect.Value(self)

		self.Close()

		return nil
	}

	ac.Error("invalid userdata")

	return nil
}

func chanPairs(th *Thread, args ...Value) []Value {
	ac := th.NewArgChecker(args)

	self := ac.ToFullUserdata(0)

	if !ac.OK() {
		return nil
	}

	if _, ok := self.GetValue().(rvalue); ok {
		return []Value{GoFunction(cnext), self}
	}

	ac.Error("invalid userdata")

	return nil
}

func cnext(th *Thread, args ...Value) []Value {
	ac := th.NewArgChecker(args)

	self := ac.ToFullUserdata(0)

	if !ac.OK() {
		return nil
	}

	if self, ok := self.GetValue().(rvalue); ok {
		self := reflect.Value(self)

		rval, ok := self.Recv()
		if !ok {
			return nil
		}

		return []Value{True, th.valueOfReflect(rval)}
	}

	ac.Error("invalid userdata")

	return nil
}

func ptrIndex(th *Thread, args ...Value) []Value {
	ac := th.NewArgChecker(args)

	self := ac.ToFullUserdata(0)
	name := ac.ToGoString(1)

	if !isPublic(name) {
		ac.Error(fmt.Sprintf("%s is not public method or field", name))
	}

	if !ac.OK() {
		return nil
	}

	if self, ok := self.GetValue().(rvalue); ok {
		self := reflect.Value(self)
		method := self.MethodByName(name)

		if !method.IsValid() {
			elem := self.Elem()

			method = elem.MethodByName(name)

			if !method.IsValid() {
				if elem.Kind() == reflect.Struct {
					field := elem.FieldByName(name)
					if field.IsValid() {
						return []Value{th.valueOfReflect(field)}
					}
				}
				ac.Error(fmt.Sprintf("type %s has no field or method %s", self.Type(), name))

				return nil
			}
		}

		return []Value{th.valueOfReflect(method)}
	}

	ac.Error("invalid userdata")

	return nil
}

func ptrNewIndex(th *Thread, args ...Value) []Value {
	ac := th.NewArgChecker(args)

	self := ac.ToFullUserdata(0)
	name := ac.ToGoString(1)
	val := ac.Get(2)

	if !isPublic(name) {
		ac.Error(fmt.Sprintf("%s is not public method or field", name))
	}

	if !ac.OK() {
		return nil
	}

	if self, ok := self.GetValue().(rvalue); ok {
		self := reflect.Value(self)

		elem := self.Elem()
		if elem.Kind() == reflect.Struct {
			field := elem.FieldByName(name)
			if field.IsValid() {
				if field.Kind() == reflect.Ptr {
					field = field.Elem()
				}

				if rval := toReflectValue(field.Type(), val); rval.IsValid() {
					field.Set(rval)

					return nil
				}

				ac.Error(fmt.Sprintf("cannot use %v (type %s) as type %s in field assignment", val, reflect.TypeOf(val), field.Type()))
			}
		}

		ac.Error(fmt.Sprintf("type %s has no field or method %s", self.Type(), name))
	}

	ac.Error("invalid userdata")

	return nil
}

func structIndex(th *Thread, args ...Value) []Value {
	ac := th.NewArgChecker(args)

	self := ac.ToFullUserdata(0)
	name := ac.ToGoString(1)

	if !isPublic(name) {
		ac.Error(fmt.Sprintf("%s is not public method or field", name))
	}

	if !ac.OK() {
		return nil
	}

	if self, ok := self.GetValue().(rvalue); ok {
		self := reflect.Value(self)
		method := self.MethodByName(name)

		if !method.IsValid() {
			if self.CanAddr() {
				method = self.Addr().MethodByName(name)
			} else {
				self2 := reflect.New(self.Type())
				self2.Elem().Set(self)
				method = self2.MethodByName(name)
			}

			if !method.IsValid() {
				field := self.FieldByName(name)
				if !field.IsValid() {
					ac.Error(fmt.Sprintf("type %s has no field or method %s", self.Type(), name))

					return nil
				}
				return []Value{th.valueOfReflect(field)}
			}
			ac.Error(fmt.Sprintf("type %s has no method %s", self.Type(), name))

			return nil
		}

		return []Value{th.valueOfReflect(method)}
	}

	ac.Error("invalid userdata")

	return nil
}

func structNewIndex(th *Thread, args ...Value) []Value {
	ac := th.NewArgChecker(args)

	self := ac.ToFullUserdata(0)
	name := ac.ToGoString(1)
	val := ac.Get(2)

	if !isPublic(name) {
		ac.Error(fmt.Sprintf("%s is not public method or field", name))
	}

	if !ac.OK() {
		return nil
	}

	if self, ok := self.GetValue().(rvalue); ok {
		self := reflect.Value(self)

		field := self.FieldByName(name)
		if field.IsValid() {
			if field.Kind() == reflect.Ptr {
				field = field.Elem()
			}

			if rval := toReflectValue(field.Type(), val); rval.IsValid() {
				field.Set(rval)

				return nil
			}

			ac.Error(fmt.Sprintf("cannot use %v (type %s) as type %s in field assignment", val, reflect.TypeOf(val), field.Type()))
		}

		ac.Error(fmt.Sprintf("type %s has no field %s", self.Type(), name))
	}

	ac.Error("invalid userdata")

	return nil
}

func ifaceIndex(th *Thread, args ...Value) []Value {
	ac := th.NewArgChecker(args)

	self := ac.ToFullUserdata(0)
	name := ac.ToGoString(1)

	if !isPublic(name) {
		ac.Error(fmt.Sprintf("%s is not public method or field", name))
	}

	if !ac.OK() {
		return nil
	}

	if self, ok := self.GetValue().(rvalue); ok {
		self := reflect.Value(self)
		method := self.MethodByName(name)

		if method.IsValid() {
			return []Value{th.valueOfReflect(method)}
		}

		ac.Error(fmt.Sprintf("type %s has no method %s", self.Type(), name))
	}

	ac.Error("invalid userdata")

	return nil
}

func index(th *Thread, args ...Value) []Value {
	ac := th.NewArgChecker(args)

	self := ac.ToFullUserdata(0)
	name := ac.ToGoString(1)

	if !isPublic(name) {
		ac.Error(fmt.Sprintf("%s is not public method or field", name))
	}

	if !ac.OK() {
		return nil
	}

	if self, ok := self.GetValue().(rvalue); ok {
		self := reflect.Value(self)
		method := self.MethodByName(name)

		if !method.IsValid() {
			if self.CanAddr() {
				method = self.Addr().MethodByName(name)
			} else {
				self2 := reflect.New(self.Type())
				self2.Elem().Set(self)
				method = self2.MethodByName(name)
			}

			if !method.IsValid() {
				ac.Error(fmt.Sprintf("type %s has no method %s", self.Type(), name))

				return nil
			}
		}

		return []Value{th.valueOfReflect(method)}
	}

	ac.Error("invalid userdata")

	return nil
}

func tostring(th *Thread, args ...Value) []Value {
	ac := th.NewArgChecker(args)

	self := ac.ToFullUserdata(0)

	if !ac.OK() {
		return nil
	}

	if self, ok := self.GetValue().(rvalue); ok {
		self := reflect.Value(self)

		return []Value{String(fmt.Sprintf("go %s: %v", self.Type(), self.Interface()))}
	}

	ac.Error("invalid userdata")

	return nil
}

func call(th *Thread, args ...Value) []Value {
	ac := th.NewArgChecker(args)

	self := ac.ToFullUserdata(0)

	if !ac.OK() {
		return nil
	}

	if self, ok := self.GetValue().(rvalue); ok {
		self := reflect.Value(self)

		styp := self.Type()

		var numin int
		if styp.IsVariadic() {
			numin = styp.NumIn() - 1
			if len(args)-1 > numin {
				numin = len(args) - 1
			}
		} else {
			numin = styp.NumIn()
		}

		rargs := make([]reflect.Value, numin)

		if len(args)-1 >= len(rargs) {
			for i := range rargs {
				if rarg := toReflectValue(styp.In(i), args[1+i]); rarg.IsValid() {
					rargs[i] = rarg
				} else {
					ac.Error(fmt.Sprintf("mismatched types %s and %s", styp.In(i), reflect.TypeOf(args[1+i])))

					return nil
				}
			}
		} else {
			for i, arg := range args[1:] {
				if rarg := toReflectValue(styp.In(i), arg); rarg.IsValid() {
					rargs[i] = rarg
				} else {
					ac.Error(fmt.Sprintf("mismatched types %s and %s", styp.In(i), reflect.TypeOf(arg)))

					return nil
				}
			}

			for i := len(args); i < len(rargs); i++ {
				if rarg := toReflectValue(styp.In(i), nil); rarg.IsValid() {
					rargs[i] = rarg
				} else {
					ac.Error(fmt.Sprintf("mismatched types %s and %s", styp.In(i), reflect.TypeOf(nil)))

					return nil
				}
			}
		}

		defer func() {
			if r := recover(); r != nil {
				ac.Error(fmt.Sprintf("panic: %s", r))
			}
		}()

		rrets := self.Call(rargs)

		rets := make([]Value, len(rrets))
		for i, rret := range rrets {
			rets[i] = th.valueOfReflect(rret)
		}

		return rets
	}

	ac.Error("invalid userdata")

	return nil
}

func cmp(op func(x, y reflect.Value) bool) GoFunction {
	return func(th *Thread, args ...Value) []Value {
		ac := th.NewArgChecker(args)

		x := ac.ToFullUserdata(0)
		y := ac.ToFullUserdata(1)

		if !ac.OK() {
			return nil
		}

		if xval, ok := x.GetValue().(rvalue); ok {
			if yval, ok := y.GetValue().(rvalue); ok {
				xval := reflect.Value(xval)
				yval := reflect.Value(yval)

				xtyp := xval.Type()
				ytyp := yval.Type()

				if xtyp == ytyp {
					return []Value{Boolean(op(xval, yval))}
				}

				ac.Error(fmt.Sprintf("mismatched types %s and %s", xtyp, ytyp))
			}
		}

		ac.Error("mismatched types")

		return nil
	}
}

func unary(op func(x reflect.Value) reflect.Value, mt *Table) GoFunction {
	return func(th *Thread, args ...Value) []Value {
		ac := th.NewArgChecker(args)

		x := ac.ToFullUserdata(0)

		if !ac.OK() {
			return nil
		}

		if xval, ok := x.GetValue().(rvalue); ok {
			xval := reflect.Value(xval)
			val := op(xval).Convert(xval.Type())

			ud := th.NewUserdata(val)

			ud.SetMetatable(mt)

			return []Value{ud}
		}

		ac.Error("invalid userdata")

		return nil
	}
}

func binary(op func(x, y reflect.Value) reflect.Value, mt *Table) GoFunction {
	return func(th *Thread, args ...Value) []Value {
		ac := th.NewArgChecker(args)

		x := ac.ToFullUserdata(0)
		y := ac.ToFullUserdata(1)

		if !ac.OK() {
			return nil
		}

		if xval, ok := x.GetValue().(rvalue); ok {
			if yval, ok := y.GetValue().(rvalue); ok {
				xval := reflect.Value(xval)
				yval := reflect.Value(yval)

				xtyp := xval.Type()
				ytyp := yval.Type()

				if xtyp == ytyp {
					val := op(xval, yval).Convert(xtyp)

					ud := th.NewUserdata(val)

					ud.SetMetatable(mt)

					return []Value{ud}
				}

				ac.Error(fmt.Sprintf("mismatched types %s and %s", xtyp, ytyp))
			}

			ac.Error("mismatched types")
		}

		ac.Error("invalid userdata")

		return nil
	}
}

func upow(x, y uint64) uint64 {
	prod := uint64(1)
	for y != 0 {
		if y&1 != 0 {
			prod *= x
		}
		y >>= 1
		x *= x
	}
	return prod
}

func ipow(x, y int64) int64 {
	prod := int64(1)
	for y != 0 {
		if y&1 != 0 {
			prod *= x
		}
		y >>= 1
		x *= x
	}
	return prod
}

func ishl(x, y int64) int64 {
	if y > 0 {
		return x << uint64(y)
	}
	return x >> uint64(-y)
}

func ishr(x, y int64) int64 {
	if y > 0 {
		return x >> uint64(y)
	}
	return x << uint64(-y)
}

func (th *Thread) imod(x, y int64) int64 {
	if y == 0 {
		th.Error("integer divide by zero")

		return 0
	}

	if x == limits.MinInt64 && y == -1 {
		return 0
	}

	rem := x % y

	if rem < 0 {
		rem += y
	}

	return rem
}

func (th *Thread) idiv(x, y int64) int64 {
	if y == 0 {
		th.Error("integer divide by zero")

		return 0
	}

	return x / y
}

func (th *Thread) umod(x, y uint64) uint64 {
	if y == 0 {
		th.Error("integer divide by zero")

		return 0
	}

	rem := x % y

	return rem
}

func (th *Thread) udiv(x, y uint64) uint64 {
	if y == 0 {
		th.Error("integer divide by zero")

		return 0
	}

	return x / y
}

func (th *Thread) fmod(x, y float64) float64 {
	rem := math.Mod(x, y)

	if rem < 0 {
		rem += y
	}

	return rem
}

func (th *Thread) fidiv(x, y float64) float64 {
	f, _ := math.Modf(x / y)

	return f
}

func aeq(x, y reflect.Value) bool {
	xlen := x.Len()
	ylen := y.Len()

	if xlen == ylen {
		for i := 0; i < xlen; i++ {
			if x.Index(i) != y.Index(i) {
				return false
			}
		}
		return true
	}

	return false
}

func isPublic(name string) bool {
	r, _ := utf8.DecodeRuneInString(name)

	return unicode.IsUpper(r)
}
