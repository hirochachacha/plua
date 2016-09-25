package load

import (
	"bytes"
	"fmt"
	"os"
	"strings"

	"github.com/hirochachacha/plua/internal/compiler_pool"
	"github.com/hirochachacha/plua/object"
	"github.com/hirochachacha/plua/object/fnutil"
)

var luaPath string

func init() {
	luaPath = os.Getenv("LUA_luaPath_5_3")
	if luaPath == "" {
		luaPath = os.Getenv("LUA_luaPath")
		if luaPath == "" {
			luaPath = defaultPath
		}
	}
}

func searchpath(th object.Thread, args ...object.Value) ([]object.Value, *object.RuntimeError) {
	ap := fnutil.NewArgParser(th, args)

	name, err := ap.ToGoString(0)
	if err != nil {
		return nil, err
	}

	path, err := ap.ToGoString(1)
	if err != nil {
		return nil, err
	}

	sep, err := ap.OptGoString(2, ".")
	if err != nil {
		return nil, err
	}

	rep, err := ap.OptGoString(3, dsep)
	if err != nil {
		return nil, err
	}

	var errmsg []string

	name = strings.Replace(name, sep, rep, -1)
	for _, p := range strings.Split(path, psep) {
		fpath := strings.Replace(p, mark, name, -1)
		_, err := os.Stat(fpath)
		if err == nil {
			return []object.Value{object.String(fpath)}, nil
		}

		errmsg = append(errmsg, fmt.Sprintf("no file '%s'", fpath))
	}

	return []object.Value{nil, object.String("\n\t" + strings.Join(errmsg, "\n\t"))}, nil
}

func preloadSearcher(th object.Thread, args ...object.Value) ([]object.Value, *object.RuntimeError) {
	ap := fnutil.NewArgParser(th, args)

	modname, err := ap.ToString(0)
	if err != nil {
		return nil, err
	}

	preload := th.Preload()

	t := preload.Get(modname)
	if t == nil {
		return []object.Value{object.String(fmt.Sprintf("\n\tno field package.preload['%s']", modname))}, nil
	}

	return []object.Value{t}, nil
}

func makeSearchers(m object.Table) []object.Value {
	luaSearcher := func(th object.Thread, args ...object.Value) ([]object.Value, *object.RuntimeError) {
		ap := fnutil.NewArgParser(th, args)

		modname, err := ap.ToString(0)
		if err != nil {
			return nil, err
		}

		loadpath, ok := object.ToString(m.Get(object.String("path")))
		if !ok {
			return nil, object.NewRuntimeError("'package.path' must be a string")
		}

		rets, err := searchpath(th, modname, loadpath)
		if err != nil {
			return nil, err
		}

		switch len(rets) {
		case 0:
			return nil, nil
		case 1:
			fpath := string(rets[0].(object.String))

			p, err := compiler_pool.CompileFile(fpath)
			if err != nil {
				return nil, object.NewRuntimeError(fmt.Sprintf("error loading module '%s' from file '%s':\n\t%s", modname, fpath, err.Error()))
			}

			return []object.Value{th.NewClosure(p), object.String(fpath)}, nil
		case 2:
			errmsg := rets[1].(object.String)

			return []object.Value{errmsg}, nil
		default:
			panic("unreachable")
		}
	}

	return []object.Value{object.GoFunction(preloadSearcher), object.GoFunction(luaSearcher)}
}

func makeRequire(m object.Table) object.GoFunction {
	Require := func(th object.Thread, args ...object.Value) ([]object.Value, *object.RuntimeError) {
		ap := fnutil.NewArgParser(th, args)

		modname, err := ap.ToString(0)
		if err != nil {
			return nil, err
		}

		loaded := th.Loaded()
		if val := loaded.Get(modname); val != nil {
			return []object.Value{val}, nil
		}

		var errbuf bytes.Buffer

		searchers, ok := m.Get(object.String("searchers")).(object.Table)
		if !ok {
			return nil, object.NewRuntimeError("'package.searchers' must be a table")
		}

		var i int
		var val object.Value

		for {
			i, val, ok = searchers.INext(i)
			if !ok {
				return nil, object.NewRuntimeError("invalid key to 'inext'")
			}

			if val == nil {
				break
			}

			if searcher, ok := val.(object.GoFunction); ok {
				rets, err := searcher(th, modname)
				if err != nil {
					return nil, err
				}
				if len(rets) > 0 {
					var fpath object.Value
					if len(rets) > 1 {
						fpath = rets[1]
					}

					if object.ToType(rets[0]) == object.TFUNCTION {
						rets, err := th.Call(rets[0], nil, modname, fpath)
						if err != nil {
							return nil, err
						}

						// treat self require
						if val := loaded.Get(modname); val != nil {
							return []object.Value{val}, nil
						}

						if len(rets) == 0 || rets[0] == nil {
							loaded.Set(modname, object.True)

							return []object.Value{object.True}, nil
						}

						if rets[0] == object.False {
							return []object.Value{object.False}, nil
						}

						loaded.Set(modname, rets[0])

						return []object.Value{rets[0]}, nil
					} else if s, ok := object.ToGoString(rets[0]); ok {
						errbuf.WriteString(s)
					}
				}
			}
		}

		return nil, object.NewRuntimeError(fmt.Sprintf("module '%s' not found:%s", modname, errbuf.String()))
	}

	return object.GoFunction(Require)
}

func Open(th object.Thread, args ...object.Value) ([]object.Value, *object.RuntimeError) {
	m := th.NewTableSize(0, 7)

	m.Set(object.String("preload"), th.Preload())
	m.Set(object.String("path"), object.String(luaPath))
	m.Set(object.String("cpath"), object.String("")) // stub for test
	m.Set(object.String("config"), object.String(config))
	m.Set(object.String("loaded"), th.Loaded())

	m.Set(object.String("searchpath"), object.GoFunction(searchpath))

	m.Set(object.String("searchers"), th.NewTableArray(makeSearchers(m)))

	th.Globals().Set(object.String("require"), makeRequire(m))

	return []object.Value{m}, nil
}
