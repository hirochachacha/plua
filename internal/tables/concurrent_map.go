package tables

// Ported from java.util.concurrent.ConcurrentHashMap.
// Originally, written by Doug Lea with assistance from members of JCP JSR-166
// Expert Group and released to the public domain, as explained at
// http://creativecommons.org/licenses/publicdomain

import (
	"sync"
	"sync/atomic"
	"unsafe"

	"github.com/hirochachacha/plua/internal/hash"
	"github.com/hirochachacha/plua/object"
)

const (
	retryCount = 2

	threshold = 0.75

	nSegmentBits = 4
	nSegments    = 16 // 2 ** nSegmentBits
	segmentShift = 64 - nSegmentBits

	minConcurrentMapSize = nSegments * 4 // 64
)

type concurrentMap struct {
	segments []*segment
	h        *hash.Hash

	lastKey          object.Value
	nextIndex        int
	nextSegmentIndex int
	nextBucket       *sbucket
	nextBuckets      []unsafe.Pointer
}

func newConcurrentMap() *concurrentMap {
	return newConcurrentMapSize(0)
}

func newConcurrentMapSize(size int) *concurrentMap {
	var sizePerSegment int
	if size < minConcurrentMapSize {
		size = minConcurrentMapSize
		sizePerSegment = size / nSegments
	} else {
		sizePerSegment = roundup(size / nSegments)
	}

	segments := make([]*segment, nSegments)
	for i := range segments {
		buckets := make([]unsafe.Pointer, sizePerSegment)
		segments[i] = &segment{
			threshold: int32(float64(sizePerSegment) * threshold),
			buckets:   unsafe.Pointer(&buckets),
		}
	}

	return &concurrentMap{
		segments: segments,
		h:        hash.New(),
	}
}

func (m *concurrentMap) Cap() int {
	sum := 0

	for _, segment := range m.segments {
		segment.m.Lock()
	}
	for _, segment := range m.segments {
		sum += len(*(*[]unsafe.Pointer)(segment.buckets))
	}
	for _, segment := range m.segments {
		segment.m.Unlock()
	}

	return sum
}

func (m *concurrentMap) Len() int {
	var sum int32
	var check int32
	var mcsum int32
	mc := make([]int32, len(m.segments))

	// non blocking
L:
	for i := 0; i < retryCount; i++ {
		sum = 0
		check = 0
		mcsum = 0

		for i, segment := range m.segments {
			sum += atomic.LoadInt32(&segment.length)
			mc[i] = segment.modCount
			mcsum += mc[i]
		}

		if mcsum != 0 {
			for i, segment := range m.segments {
				check += atomic.LoadInt32(&segment.length)
				if mc[i] != segment.modCount {
					continue L
				}
			}
		}

		// success
		if check == sum {
			return int(sum)
		}
	}

	// fallback to locking
	sum = 0
	for _, segment := range m.segments {
		segment.m.Lock()
	}
	for _, segment := range m.segments {
		sum += segment.length
	}
	for _, segment := range m.segments {
		segment.m.Unlock()
	}

	return int(sum)
}

func (m *concurrentMap) Get(key object.Value) object.Value {
	hash := m.hash(key)
	segment := m.segmentFor(hash)

	return segment.get(hash, key)
}

func (m *concurrentMap) Set(key, val object.Value) {
	hash := m.hash(key)
	segment := m.segmentFor(hash)

	segment.set(hash, key, val)
}

func (m *concurrentMap) Delete(key object.Value) {
	hash := m.hash(key)
	segment := m.segmentFor(hash)

	segment.delete(hash, key)
}

func (m *concurrentMap) Next(key object.Value) (nkey, nval object.Value, ok bool) {
	if key == nil {
		ok = true

		m.nextSegmentIndex = len(m.segments) - 1
		m.nextIndex = -1

		m.advance()

		bucket := m.nextBucket

		if bucket == nil {
			m.lastKey = nil

			return
		}

		nkey = bucket.key
		nval = *(*object.Value)(atomic.LoadPointer(&bucket.val))

		m.lastKey = nkey

		m.advance()

		return
	}

	if key == m.lastKey {
		ok = true

		bucket := m.nextBucket

		if bucket == nil {
			m.lastKey = nil

			return
		}

		nkey = bucket.key
		nval = *(*object.Value)(atomic.LoadPointer(&bucket.val))

		m.lastKey = nkey

		m.advance()

		return
	}

	hash := m.hash(key)
	sindex := hash >> segmentShift
	segment := m.segments[sindex]

	if atomic.LoadInt32(&segment.length) != 0 {
		buckets := *(*[]unsafe.Pointer)(atomic.LoadPointer(&segment.buckets))
		index := hash & uint64(len(buckets)-1)
		elem := (*sbucket)(atomic.LoadPointer(&buckets[index]))
		for elem != nil {
			if elem.hash == hash && object.Equal(elem.key, key) {
				ok = true

				m.nextIndex = int(index)
				m.nextSegmentIndex = int(sindex)
				m.nextBucket = elem
				m.nextBuckets = buckets

				m.advance()

				bucket := m.nextBucket

				if bucket == nil {
					m.lastKey = nil

					return
				}

				nkey = bucket.key
				nval = *(*object.Value)(atomic.LoadPointer(&bucket.val))
				ok = true
				m.lastKey = nkey

				m.advance()

				return
			}

			elem = elem.next
		}

		return
	}

	return
}

func (m *concurrentMap) advance() {
	if m.nextBucket != nil {
		m.nextBucket = m.nextBucket.next

		if m.nextBucket != nil {
			return
		}
	}
	for m.nextIndex >= 0 {
		m.nextBucket = (*sbucket)(atomic.LoadPointer(&m.nextBuckets[m.nextIndex]))

		m.nextIndex--

		if m.nextBucket != nil {
			return
		}
	}
	for m.nextSegmentIndex >= 0 {
		segment := m.segments[m.nextSegmentIndex]

		m.nextSegmentIndex--

		if atomic.LoadInt32(&segment.length) != 0 {
			m.nextBuckets = *(*[]unsafe.Pointer)(atomic.LoadPointer(&segment.buckets))
			for j := len(m.nextBuckets) - 1; j >= 0; j-- {
				m.nextBucket = (*sbucket)(atomic.LoadPointer(&m.nextBuckets[j]))
				if m.nextBucket != nil {
					m.nextIndex = j - 1
					return
				}
			}
		}
	}
}

func (m *concurrentMap) segmentFor(hash uint64) *segment {
	return m.segments[hash>>segmentShift]
}

func (m *concurrentMap) hash(key object.Value) uint64 {
	return m.h.Hash(key)
}

type sbucket struct {
	key      object.Value
	val      unsafe.Pointer
	hash     uint64
	next     *sbucket
	isActive bool
}

type segment struct {
	length  int32
	buckets unsafe.Pointer

	threshold int32
	modCount  int32

	m sync.Mutex
}

func (s *segment) get(hash uint64, key object.Value) object.Value {
	if atomic.LoadInt32(&s.length) != 0 {
		buckets := *(*[]unsafe.Pointer)(atomic.LoadPointer(&s.buckets))
		index := hash & uint64(len(buckets)-1)
		elem := (*sbucket)(atomic.LoadPointer(&buckets[index]))
		for elem != nil {
			if elem.hash == hash && object.Equal(elem.key, key) {
				val := *(*object.Value)(atomic.LoadPointer(&elem.val))
				if val == nil {
					s.m.Lock()

					val = *(*object.Value)(elem.val)

					s.m.Unlock()
				}

				return val
			}

			elem = elem.next
		}
	}

	return nil
}

func (s *segment) set(hash uint64, key, val object.Value) {
	s.m.Lock()

	s.unsafeSet(hash, key, val)

	s.m.Unlock()
}

func (s *segment) unsafeSet(hash uint64, key, val object.Value) {
	if s.length > s.threshold {
		s.grow()
	}

	buckets := *(*[]unsafe.Pointer)(s.buckets)
	index := hash & uint64(len(buckets)-1)
	first := (*sbucket)(buckets[index])
	elem := first
	for elem != nil {
		if elem.hash == hash && object.Equal(elem.key, key) {
			atomic.StorePointer(&elem.val, unsafe.Pointer(&val))

			return
		}

		elem = elem.next
	}

	bucket := &sbucket{
		key:  key,
		val:  unsafe.Pointer(&val),
		hash: hash,
		next: first,
	}

	s.modCount++

	atomic.StorePointer(&buckets[index], unsafe.Pointer(bucket))
	atomic.AddInt32(&s.length, 1)
}

func (s *segment) delete(hash uint64, key object.Value) {
	s.m.Lock()
	defer s.m.Unlock()

	buckets := *(*[]unsafe.Pointer)(s.buckets)
	index := hash & uint64(len(buckets)-1)
	first := (*sbucket)(buckets[index])
	elem := first
	for elem != nil {
		if elem.hash == hash && object.Equal(elem.key, key) {
			s.modCount++

			bucket := elem.next
			for p := first; p != elem; p = p.next {
				bucket = &sbucket{
					key:  p.key,
					val:  p.val,
					hash: p.hash,
					next: bucket,
				}
			}

			atomic.StorePointer(&buckets[index], unsafe.Pointer(bucket))

			atomic.AddInt32(&s.length, -1)

			return
		}

		elem = elem.next
	}
}

func (s *segment) grow() {
	old := *(*[]unsafe.Pointer)(s.buckets)

	length := len(old) * growRate

	buckets := make([]unsafe.Pointer, length)

	for _, p := range old {
		elem := (*sbucket)(p)
		if elem != nil {
			next := elem.next

			index := elem.hash & uint64(length-1)

			if next == nil {
				buckets[index] = unsafe.Pointer(elem)
			} else {
				lastElem := elem
				lastIndex := index

				for last := next; last != nil; last = last.next {
					i := last.hash & uint64(length-1)
					if i != lastIndex {
						lastIndex = i
						lastElem = last
					}
				}

				buckets[lastIndex] = unsafe.Pointer(lastElem)

				for p := elem; p != lastElem; p = p.next {
					i := p.hash & uint64(length-1)
					n := buckets[i]
					buckets[i] = unsafe.Pointer(&sbucket{key: p.key, hash: p.hash, val: p.val, next: (*sbucket)(n)})
				}
			}
		}
	}

	s.threshold = int32(float64(length) * threshold)

	atomic.StorePointer(&s.buckets, unsafe.Pointer(&buckets))
}
