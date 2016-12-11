package math

import (
	"math"
	"math/rand"

	"github.com/hirochachacha/plua/object"
	"github.com/hirochachacha/plua/object/fnutil"
)

func abs(th object.Thread, args ...object.Value) ([]object.Value, *object.RuntimeError) {
	ap := fnutil.NewArgParser(th, args)

	f, err := ap.ToGoFloat64(0)
	if err != nil {
		return nil, err
	}

	if i, ok := args[0].(object.Integer); ok {
		if i < 0 {
			return []object.Value{-i}, nil
		}

		return []object.Value{i}, nil
	}

	return []object.Value{object.Number(math.Abs(f))}, nil
}

func acos(th object.Thread, args ...object.Value) ([]object.Value, *object.RuntimeError) {
	ap := fnutil.NewArgParser(th, args)

	f, err := ap.ToGoFloat64(0)
	if err != nil {
		return nil, err
	}

	return []object.Value{object.Number(math.Acos(f))}, nil
}

func asin(th object.Thread, args ...object.Value) ([]object.Value, *object.RuntimeError) {
	ap := fnutil.NewArgParser(th, args)

	f, err := ap.ToGoFloat64(0)
	if err != nil {
		return nil, err
	}

	return []object.Value{object.Number(math.Asin(f))}, nil
}

func atan(th object.Thread, args ...object.Value) ([]object.Value, *object.RuntimeError) {
	ap := fnutil.NewArgParser(th, args)

	f, err := ap.ToGoFloat64(0)
	if err != nil {
		return nil, err
	}

	return []object.Value{object.Number(math.Atan(f))}, nil
}

func ceil(th object.Thread, args ...object.Value) ([]object.Value, *object.RuntimeError) {
	ap := fnutil.NewArgParser(th, args)

	f, err := ap.ToGoFloat64(0)
	if err != nil {
		return nil, err
	}

	return []object.Value{object.Number(math.Ceil(f))}, nil
}

func cos(th object.Thread, args ...object.Value) ([]object.Value, *object.RuntimeError) {
	ap := fnutil.NewArgParser(th, args)

	f, err := ap.ToGoFloat64(0)
	if err != nil {
		return nil, err
	}

	return []object.Value{object.Number(math.Cos(f))}, nil
}

func deg(th object.Thread, args ...object.Value) ([]object.Value, *object.RuntimeError) {
	ap := fnutil.NewArgParser(th, args)

	f, err := ap.ToGoFloat64(0)
	if err != nil {
		return nil, err
	}

	return []object.Value{object.Number((f * 180) / math.Pi)}, nil
}

func exp(th object.Thread, args ...object.Value) ([]object.Value, *object.RuntimeError) {
	ap := fnutil.NewArgParser(th, args)

	f, err := ap.ToGoFloat64(0)
	if err != nil {
		return nil, err
	}

	return []object.Value{object.Number(math.Exp(f))}, nil
}

func floor(th object.Thread, args ...object.Value) ([]object.Value, *object.RuntimeError) {
	ap := fnutil.NewArgParser(th, args)

	f, err := ap.ToGoFloat64(0)
	if err != nil {
		return nil, err
	}

	return []object.Value{object.Number(math.Floor(f))}, nil
}

func fmod(th object.Thread, args ...object.Value) ([]object.Value, *object.RuntimeError) {
	ap := fnutil.NewArgParser(th, args)

	x, err := ap.ToGoFloat64(0)
	if err != nil {
		return nil, err
	}

	y, err := ap.ToGoFloat64(1)
	if err != nil {
		return nil, err
	}

	return []object.Value{object.Number(math.Mod(x, y))}, nil
}

// log(x, [, base])
func log(th object.Thread, args ...object.Value) ([]object.Value, *object.RuntimeError) {
	ap := fnutil.NewArgParser(th, args)

	f, err := ap.ToGoFloat64(0)
	if err != nil {
		return nil, err
	}

	if len(args) == 1 {
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

func max(th object.Thread, args ...object.Value) ([]object.Value, *object.RuntimeError) {
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

func min(th object.Thread, args ...object.Value) ([]object.Value, *object.RuntimeError) {
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

func modf(th object.Thread, args ...object.Value) ([]object.Value, *object.RuntimeError) {
	ap := fnutil.NewArgParser(th, args)

	f, err := ap.ToGoFloat64(0)
	if err != nil {
		return nil, err
	}

	int, frac := math.Modf(f)

	if math.IsNaN(frac) && !math.IsNaN(int) {
		frac = 0
	}

	fval := object.Number(int)

	if ival, ok := object.ToInteger(fval); ok {
		return []object.Value{ival, object.Number(frac)}, nil
	}

	return []object.Value{fval, object.Number(frac)}, nil
}

func rad(th object.Thread, args ...object.Value) ([]object.Value, *object.RuntimeError) {
	ap := fnutil.NewArgParser(th, args)

	f, err := ap.ToGoFloat64(0)
	if err != nil {
		return nil, err
	}

	return []object.Value{object.Number((math.Pi * f) / 180)}, nil
}

// random([m, [, n]])
func random(th object.Thread, args ...object.Value) ([]object.Value, *object.RuntimeError) {
	ap := fnutil.NewArgParser(th, args)

	switch len(args) {
	case 0:
		return []object.Value{object.Number(rand.Float64())}, nil
	case 1:
		m := int64(1)

		n, err := ap.ToGoInt64(0)
		if err != nil {
			return nil, err
		}

		if n < m {
			return nil, ap.ArgError(1, "interval is empty")
		}

		return []object.Value{object.Integer(rand.Int63n(n-m) + m)}, nil
	default:
		m, err := ap.ToGoInt64(0)
		if err != nil {
			return nil, err
		}

		n, err := ap.ToGoInt64(1)
		if err != nil {
			return nil, err
		}

		if n < m {
			return nil, ap.ArgError(1, "interval is empty")
		}

		return []object.Value{object.Integer(rand.Int63n(n-m) + m)}, nil
	}
}

func randomseed(th object.Thread, args ...object.Value) ([]object.Value, *object.RuntimeError) {
	ap := fnutil.NewArgParser(th, args)

	i, err := ap.ToGoInt64(0)
	if err != nil {
		return nil, err
	}

	rand.Seed(i)

	return nil, nil
}

func sin(th object.Thread, args ...object.Value) ([]object.Value, *object.RuntimeError) {
	ap := fnutil.NewArgParser(th, args)

	f, err := ap.ToGoFloat64(0)
	if err != nil {
		return nil, err
	}

	return []object.Value{object.Number(math.Sin(f))}, nil
}

func sqrt(th object.Thread, args ...object.Value) ([]object.Value, *object.RuntimeError) {
	ap := fnutil.NewArgParser(th, args)

	f, err := ap.ToGoFloat64(0)
	if err != nil {
		return nil, err
	}

	return []object.Value{object.Number(math.Sqrt(f))}, nil
}

func tan(th object.Thread, args ...object.Value) ([]object.Value, *object.RuntimeError) {
	ap := fnutil.NewArgParser(th, args)

	f, err := ap.ToGoFloat64(0)
	if err != nil {
		return nil, err
	}

	return []object.Value{object.Number(math.Tan(f))}, nil
}

func tointeger(th object.Thread, args ...object.Value) ([]object.Value, *object.RuntimeError) {
	ap := fnutil.NewArgParser(th, args)

	i, err := ap.ToInteger(0)
	if err != nil {
		if len(args) != 0 {
			return []object.Value{nil}, nil
		}
		return nil, err
	}

	return []object.Value{i}, nil
}

func _type(th object.Thread, args ...object.Value) ([]object.Value, *object.RuntimeError) {
	ap := fnutil.NewArgParser(th, args)

	val, err := ap.ToTypes(0, object.TNUMBER)
	if err != nil {
		if len(args) != 0 {
			return []object.Value{nil}, nil
		}

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

func ult(th object.Thread, args ...object.Value) ([]object.Value, *object.RuntimeError) {
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

	m.Set(object.String("abs"), object.GoFunction(abs))
	m.Set(object.String("acos"), object.GoFunction(acos))
	m.Set(object.String("asin"), object.GoFunction(asin))
	m.Set(object.String("atan"), object.GoFunction(atan))
	m.Set(object.String("ceil"), object.GoFunction(ceil))
	m.Set(object.String("cos"), object.GoFunction(cos))
	m.Set(object.String("deg"), object.GoFunction(deg))
	m.Set(object.String("exp"), object.GoFunction(exp))
	m.Set(object.String("floor"), object.GoFunction(floor))
	m.Set(object.String("fmod"), object.GoFunction(fmod))
	m.Set(object.String("log"), object.GoFunction(log))
	m.Set(object.String("max"), object.GoFunction(max))
	m.Set(object.String("min"), object.GoFunction(min))
	m.Set(object.String("modf"), object.GoFunction(modf))
	m.Set(object.String("rad"), object.GoFunction(rad))
	m.Set(object.String("random"), object.GoFunction(random))
	m.Set(object.String("randomseed"), object.GoFunction(randomseed))
	m.Set(object.String("sin"), object.GoFunction(sin))
	m.Set(object.String("sqrt"), object.GoFunction(sqrt))
	m.Set(object.String("tan"), object.GoFunction(tan))
	m.Set(object.String("tointeger"), object.GoFunction(tointeger))
	m.Set(object.String("type"), object.GoFunction(_type))
	m.Set(object.String("ult"), object.GoFunction(ult))

	return []object.Value{m}, nil
}
