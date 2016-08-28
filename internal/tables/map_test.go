package tables

import (
	"github.com/hirochachacha/blua/object"

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
