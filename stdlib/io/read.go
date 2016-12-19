package io

import (
	"io"
	"io/ioutil"

	"github.com/hirochachacha/plua/internal/file"
	"github.com/hirochachacha/plua/object"
	"github.com/hirochachacha/plua/object/fnutil"
)

func _read(th object.Thread, args []object.Value, f file.File, off int, doClose bool, raiseError bool) ([]object.Value, *object.RuntimeError) {
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
				if raiseError {
					return nil, object.NewRuntimeError(err.Error())
				}
				return fileResult(th, err)
			}
			return []object.Value{nil}, nil
		}
		return []object.Value{line}, nil
	}

	rets := make([]object.Value, 0, len(args)-off)

	for i := range args[off:] {
		if i64, err := ap.ToGoInt64(i + off); err == nil {
			s, err := readCount(f, i64)
			if err != nil {
				if err == io.EOF {
					if len(rets) == 0 {
						if doClose {
							f.Close()
						}
						rets = append(rets, nil)
					}
					return rets, nil
				}
				if doClose {
					f.Close()
				}
				if raiseError {
					return nil, object.NewRuntimeError(err.Error())
				}
				return fileResult(th, err)
			}

			rets = append(rets, s)

			continue
		}

		fmt, err := ap.ToGoString(i + off)
		if err != nil {
			return nil, ap.TypeError(i+off, "string or integer")
		}

		if len(fmt) > 0 && fmt[0] == '*' {
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
			return nil, ap.ArgError(i+off, "invalid format")
		}

		if e != nil {
			if e == io.EOF {
				if len(rets) == 0 {
					if doClose {
						f.Close()
					}
					rets = append(rets, nil)
				}
				return rets, nil
			}
			if doClose {
				f.Close()
			}
			if raiseError {
				return nil, object.NewRuntimeError(e.Error())
			}
			return fileResult(th, e)
		}

		rets = append(rets, val)
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
	line, err := f.ReadBytes('\n')
	if err != nil {
		if err == io.EOF {
			if len(line) == 0 {
				return nil, io.EOF
			}
			return object.String(line), nil
		}
		return nil, err
	}

	return object.String(line[:len(line)-1]), nil
}

func readLine(f file.File) (s object.Value, err error) {
	line, err := f.ReadBytes('\n')
	if err != nil {
		if err == io.EOF {
			if len(line) == 0 {
				return nil, io.EOF
			}
			return object.String(line), nil
		}
		return nil, err
	}

	return object.String(line), nil
}

func readCount(f file.File, i int64) (s object.Value, err error) {
	if i == 0 {
		_, err := f.ReadByte()
		if err != nil {
			return nil, err
		}

		f.UnreadByte()

		return object.String(""), nil
	}

	bs := make([]byte, i)

	var n int
	for {
		var m int
		m, err = f.Read(bs[n:])
		n += m
		if err != nil {
			break
		}
		if i == int64(n) {
			break
		}
	}

	return object.String(bs[:n]), err
}
