package io

import (
	"io"
	"io/ioutil"

	"github.com/hirochachacha/plua/internal/file"
	"github.com/hirochachacha/plua/internal/strconv"
	"github.com/hirochachacha/plua/object"
	"github.com/hirochachacha/plua/object/fnutil"
)

func _read(th object.Thread, args []object.Value, f file.File, off int) ([]object.Value, *object.RuntimeError) {
	ap := fnutil.NewArgParser(th, args)

	if len(args) == off {
		line, err := readStrippedLine(f)
		if err != nil {
			if err != io.EOF {
				return fileResult(th, err)
			}

			return []object.Value{nil}, nil
		}
		return []object.Value{object.String(line)}, nil
	}

	var rets []object.Value

	for i := range args[off+1:] {
		if i64, err := ap.ToGoInt64(i + off + 1); err == nil {
			s, err := readCount(f, i64)
			if err != nil {
				if err != io.EOF {
					return fileResult(th, err)
				}

				rets = append(rets, nil)
			} else {
				rets = append(rets, object.String(s))
			}

			continue
		}

		fmt, err := ap.ToGoString(i + off + i)
		if err != nil {
			return nil, ap.TypeError(i+off+1, "string or integer")
		}

		if len(fmt) == 0 {
			return nil, object.NewRuntimeError("invalid format")
		}

		if fmt[0] == '*' {
			fmt = fmt[1:]
		}

		switch fmt {
		case "n":
			f64, err := readFloat(f)
			if err != nil {
				if err != io.EOF {
					return fileResult(th, err)
				}

				rets = append(rets, nil)
			} else {
				rets = append(rets, object.Number(f64))
			}
		case "i":
			i64, err := readInteger(f)
			if err != nil {
				if err != io.EOF {
					return fileResult(th, err)
				}

				rets = append(rets, nil)
			} else {
				rets = append(rets, object.Integer(i64))
			}
		case "a":
			s, err := readAll(f)
			if err != nil {
				if err != io.EOF {
					return fileResult(th, err)
				}

				rets = append(rets, nil)
			} else {
				rets = append(rets, object.String(s))
			}
		case "l":
			s, err := readStrippedLine(f)
			if err != nil {
				if err != io.EOF {
					return fileResult(th, err)
				}

				rets = append(rets, nil)
			} else {
				rets = append(rets, object.String(s))
			}
		case "L":
			s, err := readLine(f)
			if err != nil {
				if err != io.EOF {
					return fileResult(th, err)
				}

				rets = append(rets, nil)
			} else {
				rets = append(rets, object.String(s))
			}
		default:
			return nil, object.NewRuntimeError("invalid format")
		}
	}

	return rets, nil
}

func readFloat(f file.File) (f64 float64, err error) {
	return strconv.ScanFloat(f)
}

func readInteger(f file.File) (i64 int64, err error) {
	return strconv.ScanInt(f)
}

func readAll(f file.File) (s string, err error) {
	bs, err := ioutil.ReadAll(f)
	if err != nil {
		return "", err
	}

	return string(bs), nil
}

func readStrippedLine(f file.File) (s string, err error) {
	line, err := f.ReadSlice('\n')
	if err != nil {
		return "", err
	}

	return string(line[:len(line)-1]), nil
}

func readLine(f file.File) (s string, err error) {
	line, err := f.ReadSlice('\n')
	if err != nil {
		return "", err
	}

	return string(line), nil
}

func readCount(f file.File, i int64) (s string, err error) {
	bs := make([]byte, i)

	n, err := f.Read(bs)
	if err != nil {
		return "", err
	}

	return string(bs[:n]), nil
}
