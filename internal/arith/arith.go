package arith

import (
	"math"

	"github.com/hirochachacha/plua/object"
)

func Len(x object.Value) object.Value {
	switch x := x.(type) {
	case object.String:
		return object.Integer(len(x))
	case object.Table:
		return object.Integer(x.Len())
	default:
		return nil
	}
}

func Unm(x object.Value) object.Value {
	if x, ok := x.(object.Integer); ok {
		return -x
	}

	if x, ok := object.ToNumber(x); ok {
		return -x
	}

	return nil
}

func Bnot(x object.Value) object.Value {
	if x, ok := object.ToInteger(x); ok {
		return ^x
	}

	return nil
}

func Not(x object.Value) object.Value {
	return !object.ToBoolean(x)
}

func Equal(x, y object.Value) object.Value {
	return object.Boolean(object.Equal(x, y))
}

func NotEqual(x, y object.Value) object.Value {
	return !object.Boolean(object.Equal(x, y))
}

func ltIntNum(x object.Integer, y object.Number) bool {
	if -1<<52 <= x && x <= 1<<52 {
		return object.Number(x) < y
	}
	if y >= -object.Number(object.MinInteger) {
		return true
	}
	if y > object.Number(object.MinInteger) {
		return x < object.Integer(y)
	}
	return false
}

func leIntNum(x object.Integer, y object.Number) bool {
	if -1<<52 <= x && x <= 1<<52 {
		return object.Number(x) <= y
	}
	if y >= -object.Number(object.MinInteger) {
		return true
	}
	if y >= object.Number(object.MinInteger) {
		return x <= object.Integer(y)
	}
	return false
}

func LessThan(x, y object.Value) object.Value {
	switch x := x.(type) {
	case object.Integer:
		switch y := y.(type) {
		case object.Integer:
			return object.Boolean(x < y)
		case object.Number:
			return object.Boolean(ltIntNum(x, y))
		}
	case object.Number:
		switch y := y.(type) {
		case object.Integer:
			if math.IsNaN(float64(x)) {
				return object.False
			}
			return object.Boolean(!leIntNum(y, x))
		case object.Number:
			return object.Boolean(x < y)
		}
	case object.String:
		if y, ok := y.(object.String); ok {
			return object.Boolean(x < y)
		}
	}

	return nil
}

func LessThanOrEqualTo(x, y object.Value) object.Value {
	switch x := x.(type) {
	case object.Integer:
		switch y := y.(type) {
		case object.Integer:
			return object.Boolean(x <= y)
		case object.Number:
			return object.Boolean(leIntNum(x, y))
		}
	case object.Number:
		switch y := y.(type) {
		case object.Integer:
			if math.IsNaN(float64(x)) {
				return object.False
			}
			return object.Boolean(!ltIntNum(y, x))
		case object.Number:
			return object.Boolean(x <= y)
		}
	case object.String:
		if y, ok := y.(object.String); ok {
			return object.Boolean(x <= y)
		}
	}

	return nil
}

func Add(x, y object.Value) object.Value {
	if x, ok := x.(object.Integer); ok {
		if y, ok := y.(object.Integer); ok {
			return x + y
		}
	}

	if x, ok := object.ToNumber(x); ok {
		if y, ok := object.ToNumber(y); ok {
			return x + y
		}
	}

	return nil
}

func Sub(x, y object.Value) object.Value {
	if x, ok := x.(object.Integer); ok {
		if y, ok := y.(object.Integer); ok {
			return x - y
		}
	}

	if x, ok := object.ToNumber(x); ok {
		if y, ok := object.ToNumber(y); ok {
			return x - y
		}
	}

	return nil
}

func Mul(x, y object.Value) object.Value {
	if x, ok := x.(object.Integer); ok {
		if y, ok := y.(object.Integer); ok {
			return x * y
		}
	}

	if x, ok := object.ToNumber(x); ok {
		if y, ok := object.ToNumber(y); ok {
			return x * y
		}
	}

	return nil
}

func Div(x, y object.Value) object.Value {
	if x, ok := object.ToNumber(x); ok {
		if y, ok := object.ToNumber(y); ok {
			return x / y
		}
	}

	return nil
}

func Idiv(x, y object.Value) (object.Value, bool) {
	if x, ok := x.(object.Integer); ok {
		if y, ok := y.(object.Integer); ok {
			if y == 0 {
				return nil, false
			}

			z := x / y

			if (x^y) < 0 && x%y != 0 {
				z--
			}

			return z, true
		}
	}

	if xf, ok := object.ToGoFloat64(x); ok {
		if yf, ok := object.ToGoFloat64(y); ok {
			zf, frac := math.Modf(xf / yf)

			if math.Signbit(xf) != math.Signbit(yf) && frac != 0 {
				zf--
			}

			return object.Number(zf), true
		}
	}

	return nil, true
}

func Mod(x, y object.Value) (object.Value, bool) {
	if x, ok := x.(object.Integer); ok {
		if y, ok := y.(object.Integer); ok {
			if y == 0 {
				return nil, false
			}

			if x == object.MinInteger && y == -1 {
				return object.Integer(0), true
			}

			rem := x % y

			if rem < 0 {
				rem += y
			}

			return rem, true
		}
	}

	if x, ok := object.ToNumber(x); ok {
		if y, ok := object.ToNumber(y); ok {
			rem := object.Number(math.Mod(float64(x), float64(y)))

			if rem < 0 {
				rem += y
			}

			return rem, true
		}
	}

	return nil, true
}

func Pow(x, y object.Value) object.Value {
	if x, ok := object.ToNumber(x); ok {
		if y, ok := object.ToNumber(y); ok {
			return object.Number(math.Pow(float64(x), float64(y)))
		}
	}

	return nil
}

func Band(x, y object.Value) object.Value {
	if x, ok := object.ToInteger(x); ok {
		if y, ok := object.ToInteger(y); ok {
			return x & y
		}
	}

	return nil
}

func Bor(x, y object.Value) object.Value {
	if x, ok := object.ToInteger(x); ok {
		if y, ok := object.ToInteger(y); ok {
			return x | y
		}
	}

	return nil
}

func Bxor(x, y object.Value) object.Value {
	if x, ok := object.ToInteger(x); ok {
		if y, ok := object.ToInteger(y); ok {
			return x ^ y
		}
	}

	return nil
}

func Shl(x, y object.Value) object.Value {
	if x, ok := object.ToInteger(x); ok {
		if y, ok := object.ToInteger(y); ok {
			if y > 0 {
				return x << uint64(y)
			}

			return x >> uint64(-y)
		}
	}

	return nil
}

func Shr(x, y object.Value) object.Value {
	if x, ok := object.ToInteger(x); ok {
		if y, ok := object.ToInteger(y); ok {
			if y > 0 {
				return x >> uint64(y)
			}

			return x << uint64(-y)
		}
	}

	return nil
}

// not arithmetic though
func Concat(x, y object.Value) object.Value {
	if x, ok := object.ToString(x); ok {
		if y, ok := object.ToString(y); ok {
			return x + y
		}
	}

	return nil
}
