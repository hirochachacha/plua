package tables

import (
	"unsafe"

	"github.com/hirochachacha/plua/internal/hash"
	"github.com/hirochachacha/plua/object"
)

const (
	minMapSize = 8
	growRate   = 2
)

const (
	intSize    = 4 << (^uint(0) >> 63)
	bucketSize = unsafe.Sizeof(bucket{})
)

type bucket struct {
	key      object.Value
	val      object.Value
	hash     uint64
	next     *bucket
	isActive bool
}

type luaMap struct {
	length    int
	buckets   []bucket
	inactives []bucket
	h         *hash.Hash

	lastKey   object.Value
	lastIndex int
}

func newMap() *luaMap {
	return newMapSize(0)
}

func newMapSize(size int) *luaMap {
	if size < minMapSize {
		size = minMapSize
	} else {
		size = roundup(size)
	}

	buckets := make([]bucket, size)
	return &luaMap{
		buckets:   buckets,
		inactives: buckets,
		h:         hash.New(),
	}
}

func (m *luaMap) Len() int {
	return m.length
}

func (m *luaMap) Cap() int {
	return len(m.buckets)
}

func (m *luaMap) Get(key object.Value) object.Value {
	hash := m.hash(key)
	index := m.mod(hash)

	_, elem := m.findBucket(index, hash, key, nil)
	if elem == nil {
		return nil
	}

	return elem.val
}

func (m *luaMap) Set(key, val object.Value) (elem *bucket) {
	return m.setBucket(m.hash(key), key, val)
}

func (m *luaMap) Delete(key object.Value) {
	hash := m.hash(key)
	index := m.mod(hash)

	var i int
	var elem, prev *bucket

	i, elem = m.findBucket(index, hash, key, &prev)
	if elem == nil {
		return
	}

	m.deleteBucket(elem, prev)

	if len(m.inactives) <= i {
		m.inactives = m.buckets[:i]
	}
}

func (m *luaMap) Next(key object.Value) (nkey, nval object.Value, ok bool) {
	if key == nil {
		nkey, nval, m.lastIndex = m.inext(-1)
		m.lastKey = nkey
		ok = true

		return
	}

	var index int
	if key == m.lastKey {
		index = m.lastIndex
	} else {
		index = m.findIndex(key)
		if index == -1 {
			return
		}
	}

	e := &m.buckets[index]

	// see deleteBucket
	if e != nil && e.isActive && e.key != key {
		nkey, nval = e.key, e.val
	} else {
		nkey, nval, m.lastIndex = m.inext(index)
	}

	m.lastKey = nkey
	ok = true

	return
}

func (m *luaMap) hash(key object.Value) uint64 {
	return m.h.Hash(key)
}

// a % 2^n == a & (2^n-1)
func (m *luaMap) mod(hash uint64) int {
	return int(hash & uint64(m.Cap()-1))
}

func (m *luaMap) index(key object.Value) int {
	return m.mod(m.hash(key))
}

func (m *luaMap) insertBucket(index int, key object.Value) (elem *bucket) {
	elem = &m.buckets[index]

	// collision
	if elem.isActive {
		_new := m.findEmptyBucket()
		if _new == nil {
			m.grow()

			elem = m.insertBucket(m.index(key), key)

			return
		}

		other := &m.buckets[m.index(elem.key)]
		if other != nil && other != elem {
			for other.next != elem {
				other = other.next
			}
			other.next = _new
			_new.isActive = true
			_new.key = elem.key
			_new.val = elem.val
			_new.hash = elem.hash
			_new.next = elem.next

			elem.next = nil
			elem.val = nil
		} else {
			_new.next = elem.next
			elem.next = _new
			elem = _new
		}
	}

	m.length++

	elem.isActive = true
	elem.key = key

	return
}

func (m *luaMap) findBucket(index int, hash uint64, key object.Value, prev **bucket) (i int, elem *bucket) {
	var _prev *bucket

	elem = &m.buckets[index]

	for elem != nil {
		if elem.hash == hash && object.Equal(elem.key, key) {
			break
		}

		_prev = elem
		elem = elem.next
	}

	if prev != nil {
		*prev = _prev
	}

	return
}

func (m *luaMap) findEmptyBucket() (elem *bucket) {
	for i := len(m.inactives) - 1; i >= 0; i-- {
		elem = &m.inactives[i]
		if !elem.isActive {
			if i == 0 {
				m.inactives = m.buckets[:0]
			} else {
				m.inactives = m.buckets[:i-1]
			}

			return
		}
	}

	return nil
}

func (m *luaMap) setBucket(hash uint64, key, val object.Value) *bucket {
	index := m.mod(hash)

	_, elem := m.findBucket(index, hash, key, nil)
	if elem == nil {
		elem = m.insertBucket(index, key)
	}

	elem.val = val
	elem.hash = hash

	return elem
}

func (m *luaMap) deleteBucket(elem, prev *bucket) {
	next := elem.next
	if prev != nil {
		prev.next = next
	} else if next != nil {
		_next := &m.buckets[m.index(next.key)]

		if _next == elem {
			elem.key = next.key
			elem.val = next.val
			elem.hash = next.hash
			elem.next = next.next

			elem = next
		}
	}

	elem.isActive = false
	elem.key = nil
	elem.val = nil
	elem.hash = 0
	elem.next = nil

	m.length--
}

func (m *luaMap) findIndex(key object.Value) int {
	hash := m.hash(key)
	index := m.mod(hash)

	elem := &m.buckets[index]
	base := int64(uintptr(unsafe.Pointer(elem)))

	for {
		if elem == nil {
			return -1
		}

		if elem.hash == hash && object.Equal(elem.key, key) {
			break
		}

		elem = elem.next
	}

	end := int64(uintptr(unsafe.Pointer(elem)))

	offset := int((end - base) / int64(bucketSize))

	return index + offset
}

func (m *luaMap) inext(index int) (newKey, val object.Value, newIndex int) {
	for i := index + 1; i < m.Cap(); i++ {
		e := &m.buckets[i]
		if e.isActive {
			return e.key, e.val, i
		}
	}

	return nil, nil, -1
}

func (m *luaMap) grow() {
	old := m.buckets

	m.buckets = make([]bucket, len(old)*growRate)
	m.inactives = m.buckets
	m.length = 0

	for _, elem := range old {
		if elem.isActive {
			m.setBucket(elem.hash, elem.key, elem.val)
		}
	}
}

// round up to power of two
func roundup(x int) int {
	if intSize == 8 {
		x--
		x |= x >> 1
		x |= x >> 2
		x |= x >> 4
		x |= x >> 8
		x |= x >> 16
		x++
	} else {
		x--
		x |= x >> 1
		x |= x >> 2
		x |= x >> 4
		x |= x >> 8
		x++
	}

	return x
}

// for debugging
// func (m *luaMap) dump() {
// fmt.Printf("len: %d, cap: %d\n", m.Len(), m.Cap())
// for i := 0; i < m.Cap(); i++ {
// e := &m.buckets[i]
// if e.isActive {
// fmt.Printf("%4d: index: %4d, hash: %20d, key: %10v, val: %10v\n", i, m.mod(e.hash), e.hash, e.key, e.val)
// }
// }
// fmt.Println("")
// }
