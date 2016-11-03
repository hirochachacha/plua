package tables

import (
	"testing"

	"github.com/hirochachacha/plua/object"
)

func TestTableArray(t *testing.T) {
	tab := NewTableSize(0, 0)
	tab.Set(object.Integer(1), object.Integer(5))
	if tab.Get(object.Integer(1)) != object.Integer(5) {
		t.Fail()
	}
	tab.Set(object.Integer(2), object.Integer(6))
	if tab.Get(object.Integer(2)) != object.Integer(6) {
		t.Fail()
	}
	tab.Set(object.Integer(3), object.Integer(7))
	if tab.Get(object.Integer(3)) != object.Integer(7) {
		t.Fail()
	}
	tab.Set(object.Integer(5), object.Integer(9))
	if tab.Get(object.Integer(5)) != object.Integer(9) {
		t.Fail()
	}
	if tab.Len() != 3 {
		t.Fail()
	}
	tab.Set(object.Integer(4), object.Integer(8))
	if tab.Get(object.Integer(4)) != object.Integer(8) {
		t.Fail()
	}
	if tab.Len() != 5 {
		t.Fail()
	}

	tab.Set(object.Integer(3), nil)
	if tab.Get(object.Integer(3)) != nil {
		t.Fail()
	}

	k, v, _ := tab.Next(nil)
	if k != object.Integer(1) || v != object.Integer(5) {
		t.Fail()
	}
	k, v, _ = tab.Next(k)
	if k != object.Integer(2) || v != object.Integer(6) {
		t.Fail()
	}
	k, v, _ = tab.Next(k)
	if k != object.Integer(4) || v != object.Integer(8) {
		t.Fail()
	}
	k, v, _ = tab.Next(k)
	if k != object.Integer(5) || v != object.Integer(9) {
		t.Fail()
	}
	k, v, _ = tab.Next(k)
	if k != nil || v != nil {
		t.Fail()
	}
}

func TestConcurrentTableArray(t *testing.T) {
	tab := NewConcurrentTableSize(0, 0)
	tab.Set(object.Integer(1), object.Integer(5))
	if tab.Get(object.Integer(1)) != object.Integer(5) {
		t.Fail()
	}
	tab.Set(object.Integer(2), object.Integer(6))
	if tab.Get(object.Integer(2)) != object.Integer(6) {
		t.Fail()
	}
	tab.Set(object.Integer(3), object.Integer(7))
	if tab.Get(object.Integer(3)) != object.Integer(7) {
		t.Fail()
	}
	tab.Set(object.Integer(5), object.Integer(9))
	if tab.Get(object.Integer(5)) != object.Integer(9) {
		t.Fail()
	}
	if tab.Len() != 3 {
		t.Fail()
	}
	tab.Set(object.Integer(4), object.Integer(8))
	if tab.Get(object.Integer(4)) != object.Integer(8) {
		t.Fail()
	}
	if tab.Len() != 5 {
		t.Fail()
	}

	tab.Set(object.Integer(3), nil)
	if tab.Get(object.Integer(3)) != nil {
		t.Fail()
	}

	k, v, _ := tab.Next(nil)
	if k != object.Integer(1) || v != object.Integer(5) {
		t.Fail()
	}
	k, v, _ = tab.Next(k)
	if k != object.Integer(2) || v != object.Integer(6) {
		t.Fail()
	}
	k, v, _ = tab.Next(k)
	if k != object.Integer(4) || v != object.Integer(8) {
		t.Fail()
	}
	k, v, _ = tab.Next(k)
	if k != object.Integer(5) || v != object.Integer(9) {
		t.Fail()
	}
	k, v, _ = tab.Next(k)
	if k != nil || v != nil {
		t.Fail()
	}
}
