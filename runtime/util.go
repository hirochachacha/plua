package runtime

import "github.com/hirochachacha/plua/object"

// for debug
func (th *thread) getFuncName(fn object.Value) string {
	loaded := th.Loaded()

	var key object.Value
	var val object.Value
	for {
		key, val, _ = loaded.Next(key)
		if val == nil {
			break
		}

		if modname, ok := key.(object.String); ok {
			if module, ok := val.(object.Table); ok {
				var mkey object.Value
				var mval object.Value
				for {
					mkey, mval, _ = module.Next(mkey)
					if mval == nil {
						break
					}

					if fname, ok := mkey.(object.String); ok {
						if object.Equal(mval, fn) {
							if modname == "_G" {
								return string(fname)
							}
							return string(modname) + "." + string(fname)
						}
					}
				}
			}
		}
	}

	return "?"
}
