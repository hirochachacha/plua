package rand

import (
	crand "crypto/rand"
	mrand "math/rand"

	"sync"
	"unsafe"
)

var src mrand.Source
var lk sync.Mutex

func Int63() (n int64) {
	lk.Lock()
	n = src.Int63()
	lk.Unlock()
	return
}

func Bytes(n int) (bs []byte) {
	bs = make([]byte, n)

	lk.Lock()

	i := 0

	var x int64
	for j := 0; j < n/8; j++ {
		x = src.Int63()

		bs[i] = byte(x)
		bs[i+1] = byte(x >> 8)
		bs[i+2] = byte(x >> 16)
		bs[i+3] = byte(x >> 24)
		bs[i+4] = byte(x >> 32)
		bs[i+5] = byte(x >> 40)
		bs[i+6] = byte(x >> 48)
		bs[i+7] = byte(x >> 56)

		i += 8
	}

	if m := n % 8; m != 0 {
		x = src.Int63()

		for j := 0; j < m; j++ {
			bs[i+j] = byte(x >> uint(8*j))
		}
	}

	lk.Unlock()

	return
}

func init() {
	bs := make([]byte, 8)
	_, err := crand.Read(bs)
	if err != nil {
		panic(err)
	}

	src = mrand.NewSource(*(*int64)(unsafe.Pointer(&bs[0])))
}
