package pattern

import (
	"sync/atomic"
)

const (
	stringPrefix = "s"
	bytesPrefix  = "b"
)

// Simple LRU Cache
type Cache struct {
	m     map[string]*element
	l     *list
	count int
	limit int

	used int32
}

func NewCache(limit int) *Cache {
	return &Cache{
		m:     make(map[string]*element),
		l:     new(list),
		limit: limit,
	}
}

func (c *Cache) set(key string, pat *Pattern) {
	if e, ok := c.m[key]; ok {
		c.l.moveToFirst(e)
		return
	}

	e := c.l.insertFirst(key, pat)
	c.m[key] = e

	if c.count == c.limit {
		last := c.l.removeLast()
		if last != nil {
			delete(c.m, last.key)
		}
	} else {
		c.count++
	}
}

func (c *Cache) GetOrCompile(pattern string) (*Pattern, error) {
	key := stringPrefix + pattern

	if atomic.CompareAndSwapInt32(&c.used, 0, 1) {
		defer atomic.StoreInt32(&c.used, 0)

		if e, ok := c.m[key]; ok {
			return e.val, nil
		}

		pat, err := Compile(pattern)
		if err != nil {
			return nil, err
		}

		c.set(key, pat)

		return pat, nil
	}

	pat, err := Compile(pattern)
	if err != nil {
		return nil, err
	}

	if atomic.CompareAndSwapInt32(&c.used, 0, 1) {
		c.set(key, pat)

		atomic.StoreInt32(&c.used, 0)
	}

	return pat, nil
}

func (c *Cache) GetOrCompileBytes(pattern []byte, byteMatch bool) (*Pattern, error) {
	var key string
	if byteMatch {
		key = bytesPrefix + string(pattern)
	} else {
		key = stringPrefix + string(pattern)
	}

	if atomic.CompareAndSwapInt32(&c.used, 0, 1) {
		defer atomic.StoreInt32(&c.used, 0)

		if e, ok := c.m[key]; ok {
			return e.val, nil
		}

		pat, err := CompileBytes(pattern, byteMatch)
		if err != nil {
			return nil, err
		}

		c.set(key, pat)

		return pat, nil
	}

	pat, err := Compile(key)
	if err != nil {
		return nil, err
	}

	if atomic.CompareAndSwapInt32(&c.used, 0, 1) {
		c.set(key, pat)

		atomic.StoreInt32(&c.used, 0)
	}

	return pat, nil
}

type list struct {
	first *element
	last  *element
}

type element struct {
	key string
	val *Pattern

	prev *element
	next *element
}

func (l *list) moveToFirst(e *element) {
	first := l.first

	if e == first {
		return
	}

	prev := e.prev
	next := e.next

	l.first = e
	e.prev = nil
	e.next = first
	first.prev = e

	prev.next = next

	if e == l.last {
		return
	}

	next.prev = prev
}

func (l *list) removeLast() (last *element) {
	last = l.last
	if last == nil {
		return
	}

	prev := last.prev

	l.last = prev

	if prev == nil {
		return
	}

	prev.next = nil

	return
}

func (l *list) insertFirst(key string, val *Pattern) *element {
	first := l.first

	e := &element{
		key:  key,
		val:  val,
		next: first,
	}

	l.first = e

	return e
}
