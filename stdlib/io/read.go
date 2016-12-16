package io

import (
	"io"
	"io/ioutil"

	"github.com/hirochachacha/plua/internal/file"
	"github.com/hirochachacha/plua/object"
	"github.com/hirochachacha/plua/object/fnutil"
)

func _read(th object.Thread, args []object.Value, f file.File, off int, doClose bool) ([]object.Value, *object.RuntimeError) {
	if f.IsClosed() {
		return nil, object.NewRuntimeError("file is already closed")
	}

	ap := fnutil.NewArgParser(th, args)

	if len(args) == off {
		line, err := readStrippedLine(f)
		if err != nil {
			if doClose {
				f.Close()
			}
			if err != io.EOF {
				return fileResult(th, err)
			}
			return []object.Value{nil}, nil
		}
		return []object.Value{line}, nil
	}

	var rets []object.Value

	for i := range args[off:] {
		if i64, err := ap.ToGoInt64(i + off); err == nil {
			s, err := readCount(f, i64)
			if err != nil {
				if doClose {
					f.Close()
				}
				if err != io.EOF {
					return fileResult(th, err)
				}
				rets = append(rets, nil)
			} else {
				rets = append(rets, object.String(s))
			}

			continue
		}

		fmt, err := ap.ToGoString(i + off)
		if err != nil {
			return nil, ap.TypeError(i+off, "string or integer")
		}

		if len(fmt) == 0 {
			return nil, object.NewRuntimeError("invalid format")
		}

		if fmt[0] == '*' {
			fmt = fmt[1:]
		}

		var val object.Value
		var e error

		switch fmt {
		case "n":
			val, e = readNumber(f)
		case "a":
			val, e = readAll(f)
		case "l":
			val, e = readStrippedLine(f)
		case "L":
			val, e = readLine(f)
		default:
			return nil, object.NewRuntimeError("invalid format")
		}

		if e != nil {
			if doClose {
				f.Close()
			}
			if e != io.EOF {
				return fileResult(th, e)
			}

			rets = append(rets, nil)
		} else {
			rets = append(rets, val)
		}
	}

	return rets, nil
}

func readNumber(f file.File) (val object.Value, err error) {
	return newScanner(f).scanNumber()
}

func readAll(f file.File) (val object.Value, err error) {
	bs, err := ioutil.ReadAll(f)
	if err != nil {
		return nil, err
	}

	return object.String(bs), nil
}

func readStrippedLine(f file.File) (s object.Value, err error) {
	line, err := f.ReadSlice('\n')
	if err != nil {
		return nil, err
	}

	return object.String(line[:len(line)-1]), nil
}

func readLine(f file.File) (s object.Value, err error) {
	line, err := f.ReadSlice('\n')
	if err != nil {
		return nil, err
	}

	return object.String(line), nil
}

func readCount(f file.File, i int64) (s string, err error) {
	bs := make([]byte, i)

	n, err := f.Read(bs)
	if err != nil {
		return "", err
	}

	return string(bs[:n]), nil
}
