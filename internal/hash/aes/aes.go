package aes

import (
	"math/rand"
	"reflect"
	"runtime"
	"time"
	"unsafe"
)

type digest64 struct {
	seed   uintptr
	digest uintptr
}

func (d *digest64) Hash(p []byte) uint64 {
	header := (*reflect.SliceHeader)(unsafe.Pointer(&p))

	var digest uintptr

	switch header.Len {
	case 4:
		digest = aeshash32(unsafe.Pointer(header.Data), uintptr(header.Len), d.seed)
	case 8:
		digest = aeshash64(unsafe.Pointer(header.Data), uintptr(header.Len), d.seed)
	default:
		digest = aeshash(unsafe.Pointer(header.Data), uintptr(header.Len), d.seed)
	}

	return uint64(digest)
}

func New(seed uintptr) *digest64 {
	return &digest64{seed: seed}
}

// used in asm_{386,amd64}.s
const hashRandomBytes = 32

var aeskeysched [hashRandomBytes]byte

var cpuid_ecx uint32

var canUse bool

func CanUse() bool {
	return canUse
}

func init() {
	cpuid()

	canUse = runtime.GOOS != "nacl" &&
		(cpuid_ecx&(1<<25)) != 0 && // aes (aesenc)
		(cpuid_ecx&(1<<9)) != 0 && // sse3 (pshufb)
		(cpuid_ecx&(1<<19)) != 0 // sse4.1 (pinsr{d,q})

	if canUse {
		rand.Seed(time.Now().Unix())

		for i := 0; i < 4; i++ {
			j := 8 * i
			x := rand.Int63()
			for x >= 0x80 {
				aeskeysched[j] = byte(x) | 0x80
				x >>= 7
				j++
			}
		}
	}
}

func cpuid()

func aeshash(p unsafe.Pointer, s, h uintptr) uintptr
func aeshash32(p unsafe.Pointer, s, h uintptr) uintptr
func aeshash64(p unsafe.Pointer, s, h uintptr) uintptr

func aeshashbody(h uintptr) uintptr
