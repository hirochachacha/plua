package math

import (
	"math"
	"math/rand"

	"github.com/hirochachacha/plua/object"
	"github.com/hirochachacha/plua/object/fnutil"
)

func Abs(th object.Thread, args ...object.Value) ([]object.Value, *object.RuntimeError) {
	ap := fnutil.NewArgParser(th, args)

	f, err := ap.ToGoFloat64(0)
	if err != nil {
		return nil, err
	}

	if i, ok := args[0].(object.Integer); ok {
		if i == object.MinInteger {
			return []object.Value{object.Infinity}, nil
		}

		if i < 0 {
			return []object.Value{-i}, nil
		}

		return []object.Value{i}, nil
	}

	return []object.Value{object.Number(math.Abs(f))}, nil
}

func Acos(th object.Thread, args ...object.Value) ([]object.Value, *object.RuntimeError) {
	ap := fnutil.NewArgParser(th, args)

	f, err := ap.ToGoFloat64(0)
	if err != nil {
		return nil, err
	}

	return []object.Value{object.Number(math.Acos(f))}, nil
}

func Asin(th object.Thread, args ...object.Value) ([]object.Value, *object.RuntimeError) {
	ap := fnutil.NewArgParser(th, args)

	f, err := ap.ToGoFloat64(0)
	if err != nil {
		return nil, err
	}

	return []object.Value{object.Number(math.Asin(f))}, nil
}

func Atan(th object.Thread, args ...object.Value) ([]object.Value, *object.RuntimeError) {
	ap := fnutil.NewArgParser(th, args)

	f, err := ap.ToGoFloat64(0)
	if err != nil {
		return nil, err
	}

	return []object.Value{object.Number(math.Atan(f))}, nil
}

func Ceil(th object.Thread, args ...object.Value) ([]object.Value, *object.RuntimeError) {
	ap := fnutil.NewArgParser(th, args)

	f, err := ap.ToGoFloat64(0)
	if err != nil {
		return nil, err
	}

	return []object.Value{object.Number(math.Ceil(f))}, nil
}

func Cos(th object.Thread, args ...object.Value) ([]object.Value, *object.RuntimeError) {
	ap := fnutil.NewArgParser(th, args)

	f, err := ap.ToGoFloat64(0)
	if err != nil {
		return nil, err
	}

	return []object.Value{object.Number(math.Cos(f))}, nil
}

func Deg(th object.Thread, args ...object.Value) ([]object.Value, *object.RuntimeError) {
	ap := fnutil.NewArgParser(th, args)

	f, err := ap.ToGoFloat64(0)
	if err != nil {
		return nil, err
	}

	return []object.Value{object.Number((f * 180) / math.Pi)}, nil
}

func Exp(th object.Thread, args ...object.Value) ([]object.Value, *object.RuntimeError) {
	ap := fnutil.NewArgParser(th, args)

	f, err := ap.ToGoFloat64(0)
	if err != nil {
		return nil, err
	}

	return []object.Value{object.Number(math.Exp(f))}, nil
}

func Floor(th object.Thread, args ...object.Value) ([]object.Value, *object.RuntimeError) {
	ap := fnutil.NewArgParser(th, args)

	f, err := ap.ToGoFloat64(0)
	if err != nil {
		return nil, err
	}

	return []object.Value{object.Number(math.Floor(f))}, nil
}

func Fmod(th object.Thread, args ...object.Value) ([]object.Value, *object.RuntimeError) {
	ap := fnutil.NewArgParser(th, args)

	x, err := ap.ToGoFloat64(0)
	if err != nil {
		return nil, err
	}

	y, err := ap.ToGoFloat64(0)
	if err != nil {
		return nil, err
	}

	return []object.Value{object.Number(math.Mod(x, y))}, nil
}

// log(x, [, base])
func Log(th object.Thread, args ...object.Value) ([]object.Value, *object.RuntimeError) {
	ap := fnutil.NewArgParser(th, args)

	f, err := ap.ToGoFloat64(0)
	if err != nil {
		return nil, err
	}

	if _, ok := ap.Get(0); !ok {
		return []object.Value{object.Number(math.Log(f))}, nil
	}

	base, err := ap.ToGoInt64(1)
	if err != nil {
		return nil, err
	}

	switch base {
	case 2:
		return []object.Value{object.Number(math.Log2(f))}, nil
	case 10:
		return []object.Value{object.Number(math.Log10(f))}, nil
	default:
		return []object.Value{object.Number(math.Log(f) / math.Log(float64(base)))}, nil
	}
}

func Max(th object.Thread, args ...object.Value) ([]object.Value, *object.RuntimeError) {
	ap := fnutil.NewArgParser(th, args)

	maxi := 0

	max, err := ap.ToNumber(0)
	if err != nil {
		return nil, err
	}

	for i := range args[1:] {
		f, err := ap.ToNumber(i + 1)
		if err != nil {
			return nil, err
		}

		if max < f {
			maxi = i + 1
			max = f
		}
	}

	maxv := args[maxi]

	if _, ok := maxv.(object.Number); !ok {
		if i, ok := object.ToInteger(maxv); ok {
			return []object.Value{i}, nil
		}
	}

	return []object.Value{max}, nil
}

func Min(th object.Thread, args ...object.Value) ([]object.Value, *object.RuntimeError) {
	ap := fnutil.NewArgParser(th, args)

	mini := 0

	min, err := ap.ToNumber(0)
	if err != nil {
		return nil, err
	}

	for i := range args[1:] {
		f, err := ap.ToNumber(i + 1)
		if err != nil {
			return nil, err
		}

		if min > f {
			mini = i + 1
			min = f
		}
	}

	minv := args[mini]

	if _, ok := minv.(object.Number); !ok {
		if i, ok := object.ToInteger(minv); ok {
			return []object.Value{i}, nil
		}
	}

	return []object.Value{min}, nil
}

func Modf(th object.Thread, args ...object.Value) ([]object.Value, *object.RuntimeError) {
	ap := fnutil.NewArgParser(th, args)

	f, err := ap.ToGoFloat64(0)
	if err != nil {
		return nil, err
	}

	i, frac := math.Modf(f)

	return []object.Value{object.Integer(i), object.Number(frac)}, nil
}

func Rad(th object.Thread, args ...object.Value) ([]object.Value, *object.RuntimeError) {
	ap := fnutil.NewArgParser(th, args)

	f, err := ap.ToGoFloat64(0)
	if err != nil {
		return nil, err
	}

	return []object.Value{object.Number((math.Pi * f) / 180)}, nil
}

// random([m, [, n]])
func Random(th object.Thread, args ...object.Value) ([]object.Value, *object.RuntimeError) {
	ap := fnutil.NewArgParser(th, args)

	m := int64(1)

	n, err := ap.ToGoInt64(0)
	if err != nil {
		return nil, err
	}

	if _, ok := ap.Get(1); ok {
		m = n
		n, err = ap.ToGoInt64(1)
		if err != nil {
			return nil, err
		}
	}

	if n < m {
		return nil, ap.ArgError(1, "interval is empty")
	}

	return []object.Value{object.Integer(rand.Int63n(n-m) + m)}, nil
}

func RandomSeed(th object.Thread, args ...object.Value) ([]object.Value, *object.RuntimeError) {
	ap := fnutil.NewArgParser(th, args)

	i, err := ap.ToGoInt64(0)
	if err != nil {
		return nil, err
	}

	rand.Seed(i)

	return nil, nil
}

func Sin(th object.Thread, args ...object.Value) ([]object.Value, *object.RuntimeError) {
	ap := fnutil.NewArgParser(th, args)

	f, err := ap.ToGoFloat64(0)
	if err != nil {
		return nil, err
	}

	return []object.Value{object.Number(math.Sin(f))}, nil
}

func Sqrt(th object.Thread, args ...object.Value) ([]object.Value, *object.RuntimeError) {
	ap := fnutil.NewArgParser(th, args)

	f, err := ap.ToGoFloat64(0)
	if err != nil {
		return nil, err
	}

	return []object.Value{object.Number(math.Sqrt(f))}, nil
}

func Tan(th object.Thread, args ...object.Value) ([]object.Value, *object.RuntimeError) {
	ap := fnutil.NewArgParser(th, args)

	f, err := ap.ToGoFloat64(0)
	if err != nil {
		return nil, err
	}

	return []object.Value{object.Number(math.Tan(f))}, nil
}

func ToInteger(th object.Thread, args ...object.Value) ([]object.Value, *object.RuntimeError) {
	ap := fnutil.NewArgParser(th, args)

	i, err := ap.ToInteger(0)
	if err != nil {
		return nil, err
	}

	return []object.Value{i}, nil
}

func Type(th object.Thread, args ...object.Value) ([]object.Value, *object.RuntimeError) {
	ap := fnutil.NewArgParser(th, args)

	val, err := ap.ToTypes(0, object.TNUMBER)
	if err != nil {
		return nil, err
	}

	switch val.(type) {
	case object.Integer:
		return []object.Value{object.String("integer")}, nil
	case object.Number:
		return []object.Value{object.String("float")}, nil
	}

	return nil, nil
}

func Ult(th object.Thread, args ...object.Value) ([]object.Value, *object.RuntimeError) {
	ap := fnutil.NewArgParser(th, args)

	x, err := ap.ToGoInt64(0)
	if err != nil {
		return nil, err
	}

	y, err := ap.ToGoInt64(1)
	if err != nil {
		return nil, err
	}

	return []object.Value{object.Boolean(uint64(x) < uint64(y))}, nil
}

func Open(th object.Thread, args ...object.Value) ([]object.Value, *object.RuntimeError) {
	m := th.NewTableSize(0, 27)

	m.Set(object.String("huge"), object.Infinity)
	m.Set(object.String("pi"), object.Number(math.Pi))
	m.Set(object.String("mininteger"), object.MinInteger)
	m.Set(object.String("maxinteger"), object.MaxInteger)

	m.Set(object.String("abs"), object.GoFunction(Abs))
	m.Set(object.String("acos"), object.GoFunction(Acos))
	m.Set(object.String("asin"), object.GoFunction(Asin))
	m.Set(object.String("atan"), object.GoFunction(Atan))
	m.Set(object.String("ceil"), object.GoFunction(Ceil))
	m.Set(object.String("cos"), object.GoFunction(Cos))
	m.Set(object.String("deg"), object.GoFunction(Deg))
	m.Set(object.String("exp"), object.GoFunction(Exp))
	m.Set(object.String("floor"), object.GoFunction(Floor))
	m.Set(object.String("fmod"), object.GoFunction(Fmod))
	m.Set(object.String("log"), object.GoFunction(Log))
	m.Set(object.String("max"), object.GoFunction(Max))
	m.Set(object.String("min"), object.GoFunction(Min))
	m.Set(object.String("modf"), object.GoFunction(Modf))
	m.Set(object.String("rad"), object.GoFunction(Rad))
	m.Set(object.String("random"), object.GoFunction(Random))
	m.Set(object.String("randomseed"), object.GoFunction(RandomSeed))
	m.Set(object.String("sin"), object.GoFunction(Sin))
	m.Set(object.String("sqrt"), object.GoFunction(Sqrt))
	m.Set(object.String("tan"), object.GoFunction(Tan))
	m.Set(object.String("tointeger"), object.GoFunction(ToInteger))
	m.Set(object.String("type"), object.GoFunction(Type))
	m.Set(object.String("ult"), object.GoFunction(Ult))

	return []object.Value{m}, nil
}
