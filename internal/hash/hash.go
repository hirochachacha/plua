package hash

import (
	"hash"
	"reflect"
	"unsafe"

	"github.com/dchest/siphash"

	"github.com/hirochachacha/plua/internal/limits"
	"github.com/hirochachacha/plua/internal/rand"

	"github.com/hirochachacha/plua/object"
)

type Hash struct {
	h hash.Hash64
}

func New() *Hash {
	return &Hash{h: siphash.New(rand.Bytes(16))}
}

func (h *Hash) Sum(key object.Value) (sum uint64) {
	h.h.Reset()

	switch key := key.(type) {
	case nil:
		return 0
	case object.Integer:
		h.h.Write((*(*[8]byte)(unsafe.Pointer(&key)))[:])
	case object.Number:
		h.h.Write((*(*[8]byte)(unsafe.Pointer(&key)))[:])
	case object.String:
		sheader := (*reflect.StringHeader)(unsafe.Pointer(&key))

		bheader := &reflect.SliceHeader{
			Data: sheader.Data,
			Len:  sheader.Len,
			Cap:  sheader.Len,
		}

		h.h.Write(*(*[]byte)(unsafe.Pointer(bheader)))
	case object.Boolean:
		h.h.Write((*(*[1]byte)(unsafe.Pointer(&key)))[:])
	case object.LightUserdata:
		ikey := uintptr(key.Pointer)
		h.h.Write((*(*[limits.PtrSize]byte)(unsafe.Pointer(&ikey)))[:])
	case object.GoFunction:
		ikey := uintptr(reflect.ValueOf(key).Pointer())
		h.h.Write((*(*[limits.PtrSize]byte)(unsafe.Pointer(&ikey)))[:])
	default:
		ikey := uintptr(reflect.ValueOf(key).Pointer())
		h.h.Write((*(*[limits.PtrSize]byte)(unsafe.Pointer(&ikey)))[:])
	}

	return h.h.Sum64() + 1
}
