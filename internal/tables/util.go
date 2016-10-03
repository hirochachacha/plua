package tables

import "github.com/hirochachacha/plua/object"

func normKey(key object.Value) object.Value {
	if n, ok := key.(object.Number); ok {
		i := object.Integer(n)
		if object.Number(i) == n {
			return i
		}
	}
	return key
}
