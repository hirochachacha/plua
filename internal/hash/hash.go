package hash

import (
	"reflect"
	"unsafe"

	"github.com/hirochachacha/blua/internal/hash/aes"
	"github.com/hirochachacha/blua/internal/hash/sip"
	"github.com/hirochachacha/blua/internal/limits"
	"github.com/hirochachacha/blua/internal/rand"

	"github.com/hirochachacha/blua/object"
)

type iHash interface {
	Hash([]byte) uint64
}

type Hash struct {
	h iHash
}

func New() *Hash {
	if aes.CanUse() {
		return &Hash{h: newAESHash()}
	}
	return &Hash{newSipHash()}
}

func (h *Hash) Hash(key object.Value) (sum uint64) {
	switch key := key.(type) {
	case nil:
		return 0
	case object.Integer:
		sum = h.h.Hash((*(*[8]byte)(unsafe.Pointer(&key)))[:])
	case object.Number:
		sum = h.h.Hash((*(*[8]byte)(unsafe.Pointer(&key)))[:])
	case object.String:
		sheader := (*reflect.StringHeader)(unsafe.Pointer(&key))

		bheader := &reflect.SliceHeader{
			Data: sheader.Data,
			Len:  sheader.Len,
			Cap:  sheader.Len,
		}

		sum = h.h.Hash(*(*[]byte)(unsafe.Pointer(bheader)))
	case object.Boolean:
		sum = h.h.Hash((*(*[1]byte)(unsafe.Pointer(&key)))[:])
	case object.LightUserdata:
		ikey := uintptr(key.Pointer)
		sum = h.h.Hash((*(*[limits.PtrSize]byte)(unsafe.Pointer(&ikey)))[:])
	case object.GoFunction:
		ikey := uintptr(reflect.ValueOf(key).Pointer())
		sum = h.h.Hash((*(*[limits.PtrSize]byte)(unsafe.Pointer(&ikey)))[:])
	default:
		ikey := uintptr(reflect.ValueOf(key).Pointer())
		sum = h.h.Hash((*(*[limits.PtrSize]byte)(unsafe.Pointer(&ikey)))[:])
	}

	return sum + 1
}

func newAESHash() iHash {
	var seed uintptr

	if limits.PtrSize == 4 {
		seed = uintptr(rand.Int63() >> 31)
	} else {
		seed = uintptr(rand.Int63())
	}

	return aes.New(seed)
}

func newSipHash() iHash {
	return sip.New(rand.Bytes(16))
}
