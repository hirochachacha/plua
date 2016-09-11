package tables

import (
	"fmt"

	"github.com/hirochachacha/plua/object"

	"testing"
)

func TestMap(t *testing.T) {
	m := newMap()
	if m.Len() != 0 {
		t.Fail()
	}

	m.Set(object.Integer(10), object.Integer(5))
	if m.Len() != 1 {
		t.Fail()
	}

	m.Set(object.Integer(6), object.Integer(7))
	if m.Len() != 2 {
		t.Fail()
	}

	m.Set(object.String("fake"), object.Number(7))
	if m.Len() != 3 {
		t.Fail()
	}

	m.Set(object.String("ake"), object.Number(7))
	if m.Len() != 4 {
		t.Fail()
	}

	m.Set(object.String("ke"), object.Number(7))
	if m.Len() != 5 {
		t.Fail()
	}

	m.Set(object.String("e"), object.Number(7))
	if m.Len() != 6 {
		t.Fail()
	}

	m.Set(object.String("ez"), object.Number(7))
	if m.Len() != 7 {
		t.Fail()
	}

	m.Set(object.String("ezk"), object.Number(7))
	if m.Len() != 8 {
		t.Fail()
	}

	m.Set(object.String("ezkl"), object.Number(7))
	if m.Len() != 9 {
		t.Fail()
	}

	m.Set(object.String("ezkli"), object.Number(7))
	if m.Len() != 10 {
		t.Fail()
	}

	m.Set(object.String("ezklij"), object.Number(7))
	if m.Len() != 11 {
		t.Fail()
	}

	if m.Get(object.Integer(10)) != object.Integer(5) {
		t.Fail()
	}

	if m.Get(object.Integer(6)) != object.Integer(7) {
		t.Fail()
	}

	if m.Get(object.String("fake")) != object.Number(7) {
		t.Fail()
	}

	if m.Get(object.String("ake")) != object.Number(7) {
		t.Fail()
	}

	if m.Get(object.String("ke")) != object.Number(7) {
		t.Fail()
	}

	if m.Get(object.String("e")) != object.Number(7) {
		t.Fail()
	}

	if m.Get(object.String("ez")) != object.Number(7) {
		t.Fail()
	}

	if m.Get(object.String("ezk")) != object.Number(7) {
		t.Fail()
	}

	if m.Get(object.String("ezkl")) != object.Number(7) {
		t.Fail()
	}

	if m.Get(object.String("ezkli")) != object.Number(7) {
		t.Fail()
	}

	if m.Get(object.String("ezklij")) != object.Number(7) {
		t.Fail()
	}

	m.Delete(object.String("ezklij"))

	if m.Get(object.String("ezklij")) != nil {
		t.Fail()
	}

	// m.dump()
}

func TestMapNext(t *testing.T) {
	m := newMap()

	mm := make(map[object.Integer]bool)

	for i := 0; i < 1000; i++ {
		m.Set(object.String(fmt.Sprintf("%d", i)), object.Integer(i))

		mm[object.Integer(i)] = false
	}

	var count int

	var key object.Value
	var val object.Value
	for {
		key, val, _ = m.Next(key)
		if val == nil {
			break
		}

		i, ok := val.(object.Integer)
		if !ok {
			t.Fatalf("unknown value: %v", val)
		}

		if ok := mm[i]; ok {
			t.Fatal("unexpected loop")
		}

		mm[i] = true

		count++
	}

	if count != 1000 {
		t.Fatal("short numbers")
	}
}

func TestConcurrentMap(t *testing.T) {
	m := newConcurrentMap()
	if m.Len() != 0 {
		t.Fail()
	}

	m.Set(object.Integer(10), object.Integer(5))
	if m.Len() != 1 {
		t.Fail()
	}

	m.Set(object.Integer(6), object.Integer(7))
	if m.Len() != 2 {
		t.Fail()
	}

	m.Set(object.String("fake"), object.Number(7))
	if m.Len() != 3 {
		t.Fail()
	}

	m.Set(object.String("ake"), object.Number(7))
	if m.Len() != 4 {
		t.Fail()
	}

	m.Set(object.String("ke"), object.Number(7))
	if m.Len() != 5 {
		t.Fail()
	}

	m.Set(object.String("e"), object.Number(7))
	if m.Len() != 6 {
		t.Fail()
	}

	m.Set(object.String("ez"), object.Number(7))
	if m.Len() != 7 {
		t.Fail()
	}

	m.Set(object.String("ezk"), object.Number(7))
	if m.Len() != 8 {
		t.Fail()
	}

	m.Set(object.String("ezkl"), object.Number(7))
	if m.Len() != 9 {
		t.Fail()
	}

	m.Set(object.String("ezkli"), object.Number(7))
	if m.Len() != 10 {
		t.Fail()
	}

	m.Set(object.String("ezklij"), object.Number(7))
	if m.Len() != 11 {
		t.Fail()
	}

	if m.Get(object.Integer(10)) != object.Integer(5) {
		t.Fail()
	}

	if m.Get(object.Integer(6)) != object.Integer(7) {
		t.Fail()
	}

	if m.Get(object.String("fake")) != object.Number(7) {
		t.Fail()
	}

	if m.Get(object.String("ake")) != object.Number(7) {
		t.Fail()
	}

	if m.Get(object.String("ke")) != object.Number(7) {
		t.Fail()
	}

	if m.Get(object.String("e")) != object.Number(7) {
		t.Fail()
	}

	if m.Get(object.String("ez")) != object.Number(7) {
		t.Fail()
	}

	if m.Get(object.String("ezk")) != object.Number(7) {
		t.Fail()
	}

	if m.Get(object.String("ezkl")) != object.Number(7) {
		t.Fail()
	}

	if m.Get(object.String("ezkli")) != object.Number(7) {
		t.Fail()
	}

	if m.Get(object.String("ezklij")) != object.Number(7) {
		t.Fail()
	}

	m.Delete(object.String("ezklij"))

	if m.Get(object.String("ezklij")) != nil {
		t.Fail()
	}

	// m.dump()
}

func TestConcurrent(t *testing.T) {
	m := newConcurrentMap()
	if m.Len() != 0 {
		t.Fail()
	}

	m.Set(object.Integer(10), object.Integer(5))
	if m.Len() != 1 {
		t.Fail()
	}

	m.Set(object.Integer(6), object.Integer(7))
	if m.Len() != 2 {
		t.Fail()
	}

	m.Set(object.String("fake"), object.Number(7))
	if m.Len() != 3 {
		t.Fail()
	}

	m.Set(object.String("ake"), object.Number(7))
	if m.Len() != 4 {
		t.Fail()
	}

	m.Set(object.String("ke"), object.Number(7))
	if m.Len() != 5 {
		t.Fail()
	}

	m.Set(object.String("e"), object.Number(7))
	if m.Len() != 6 {
		t.Fail()
	}

	m.Set(object.String("ez"), object.Number(7))
	if m.Len() != 7 {
		t.Fail()
	}

	m.Set(object.String("ezk"), object.Number(7))
	if m.Len() != 8 {
		t.Fail()
	}

	m.Set(object.String("ezkl"), object.Number(7))
	if m.Len() != 9 {
		t.Fail()
	}

	m.Set(object.String("ezkli"), object.Number(7))
	if m.Len() != 10 {
		t.Fail()
	}

	m.Set(object.String("ezklij"), object.Number(7))
	if m.Len() != 11 {
		t.Fail()
	}

	if m.Get(object.Integer(10)) != object.Integer(5) {
		t.Fail()
	}

	if m.Get(object.Integer(6)) != object.Integer(7) {
		t.Fail()
	}

	if m.Get(object.String("fake")) != object.Number(7) {
		t.Fail()
	}

	if m.Get(object.String("ake")) != object.Number(7) {
		t.Fail()
	}

	if m.Get(object.String("ke")) != object.Number(7) {
		t.Fail()
	}

	if m.Get(object.String("e")) != object.Number(7) {
		t.Fail()
	}

	if m.Get(object.String("ez")) != object.Number(7) {
		t.Fail()
	}

	if m.Get(object.String("ezk")) != object.Number(7) {
		t.Fail()
	}

	if m.Get(object.String("ezkl")) != object.Number(7) {
		t.Fail()
	}

	if m.Get(object.String("ezkli")) != object.Number(7) {
		t.Fail()
	}

	if m.Get(object.String("ezklij")) != object.Number(7) {
		t.Fail()
	}

	m.Delete(object.String("ezklij"))

	if m.Get(object.String("ezklij")) != nil {
		t.Fail()
	}

	// m.dump()
}

func TestConcurrentMapNext(t *testing.T) {
	m := newConcurrentMap()

	mm := make(map[object.Integer]bool)

	for i := 0; i < 1000; i++ {
		m.Set(object.String(fmt.Sprintf("%d", i)), object.Integer(i))

		mm[object.Integer(i)] = false
	}

	var count int

	var key object.Value
	var val object.Value
	for {
		key, val, _ = m.Next(key)
		if val == nil {
			break
		}

		i, ok := val.(object.Integer)
		if !ok {
			continue
			// t.Fatalf("unknown value: %v", val)
		}

		if ok := mm[i]; ok {
			t.Fatal("unexpected loop")
		}

		mm[i] = true

		count++
	}

	if count != 1000 {
		t.Fatal("short numbers")
	}
}
