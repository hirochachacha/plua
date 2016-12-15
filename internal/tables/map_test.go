package tables

import (
	"testing"
	"testing/quick"

	"github.com/hirochachacha/plua/object"
)

func TestMap(t *testing.T) {
	f := func(mm map[object.Integer]object.Integer) bool {
		m := newMap()

		l := 0
		if m.Len() != l {
			return false
		}

		for k, v := range mm {
			m.Set(k, v)

			l++

			if m.Len() != l {
				return false
			}
		}

		var pkey object.Value
		var key object.Value
		var val object.Value
		for {
			key, val, _ = m.Next(pkey)
			if val == nil {
				break
			}

			k1, v1, _ := m.Next(pkey)
			if key != k1 || val != v1 {
				return false
			}

			if m.Get(key) != val {
				return false
			}

			ikey := key.(object.Integer)
			ival := val.(object.Integer)

			if mm[ikey] != ival {
				return false
			}

			delete(mm, ikey)

			if m.Len() != l {
				return false
			}

			pkey = key
		}

		if len(mm) != 0 {
			return false
		}

		return true
	}

	if err := quick.Check(f, nil); err != nil {
		t.Error(err)
	}
}

func TestConcurrentMap(t *testing.T) {
	f := func(mm map[object.Integer]object.Integer) bool {
		m := newConcurrentMap()

		l := 0
		if m.Len() != l {
			return false
		}

		for k, v := range mm {
			m.Set(k, v)

			l++

			if m.Len() != l {
				return false
			}
		}

		var pkey object.Value
		var key object.Value
		var val object.Value
		for {
			key, val, _ = m.Next(pkey)
			if val == nil {
				break
			}

			k1, v1, _ := m.Next(pkey)
			if key != k1 || val != v1 {
				return false
			}

			if m.Get(key) != val {
				return false
			}

			ikey := key.(object.Integer)
			ival := val.(object.Integer)

			if mm[ikey] != ival {
				return false
			}

			delete(mm, ikey)

			if m.Len() != l {
				return false
			}

			pkey = key
		}

		if len(mm) != 0 {
			return false
		}

		return true
	}

	if err := quick.Check(f, nil); err != nil {
		t.Error(err)
	}
}
