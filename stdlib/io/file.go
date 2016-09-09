package io

import (
	"fmt"

	"github.com/hirochachacha/plua/internal/file"
	"github.com/hirochachacha/plua/object"
	"github.com/hirochachacha/plua/object/fnutil"
)

func FClose(th object.Thread, args ...object.Value) ([]object.Value, *object.RuntimeError) {
	ap := fnutil.NewArgParser(th, args)

	ud, err := ap.ToFullUserdata(0)
	if err != nil {
		return nil, err
	}

	f, ok := ud.Value.(file.File)
	if !ok {
		return nil, ap.TypeError(0, "FILE*")
	}

	return fileResult(th, f.Close())
}

func FFlush(th object.Thread, args ...object.Value) ([]object.Value, *object.RuntimeError) {
	ap := fnutil.NewArgParser(th, args)

	ud, err := ap.ToFullUserdata(0)
	if err != nil {
		return nil, err
	}

	f, ok := ud.Value.(file.File)
	if !ok {
		return nil, ap.TypeError(0, "FILE*")
	}

	return fileResult(th, f.Flush())
}

func FLines(th object.Thread, args ...object.Value) ([]object.Value, *object.RuntimeError) {
	ap := fnutil.NewArgParser(th, args)

	ud, err := ap.ToFullUserdata(0)
	if err != nil {
		return nil, err
	}

	f, ok := ud.Value.(file.File)
	if !ok {
		return nil, ap.TypeError(0, "FILE*")
	}

	retfn := func(_ object.Thread, _ ...object.Value) ([]object.Value, *object.RuntimeError) {
		return fread(th, args, f, 1)
	}

	return []object.Value{object.GoFunction(retfn)}, nil
}

func FRead(th object.Thread, args ...object.Value) ([]object.Value, *object.RuntimeError) {
	ap := fnutil.NewArgParser(th, args)

	ud, err := ap.ToFullUserdata(0)
	if err != nil {
		return nil, err
	}

	f, ok := ud.Value.(file.File)
	if !ok {
		return nil, ap.TypeError(0, "FILE*")
	}

	return fread(th, args, f, 1)
}

func FSeek(th object.Thread, args ...object.Value) ([]object.Value, *object.RuntimeError) {
	ap := fnutil.NewArgParser(th, args)

	ud, err := ap.ToFullUserdata(0)
	if err != nil {
		return nil, err
	}

	f, ok := ud.Value.(file.File)
	if !ok {
		return nil, ap.TypeError(0, "FILE*")
	}

	whence, err := ap.OptGoString(1, "cur")
	if err != nil {
		return nil, err
	}

	offset, err := ap.OptGoInt64(2, 0)
	if err != nil {
		return nil, err
	}

	var n int64
	var e error

	switch whence {
	case "set":
		n, e = f.Seek(offset, 0)
	case "cur":
		n, e = f.Seek(offset, 1)
	case "end":
		n, e = f.Seek(offset, 2)
	default:
		return nil, ap.OptionError(1, whence)
	}

	if e != nil {
		return fileResult(th, e)
	}

	return []object.Value{object.Integer(n)}, nil
}

func FSetvbuf(th object.Thread, args ...object.Value) ([]object.Value, *object.RuntimeError) {
	ap := fnutil.NewArgParser(th, args)

	ud, err := ap.ToFullUserdata(0)
	if err != nil {
		return nil, err
	}

	f, ok := ud.Value.(file.File)
	if !ok {
		return nil, ap.TypeError(0, "FILE*")
	}

	mode, err := ap.ToGoString(1)
	if err != nil {
		return nil, err
	}

	size, err := ap.OptGoInt(2, 0)
	if err != nil {
		return nil, err
	}

	var e error

	switch mode {
	case "no":
		e = f.Setvbuf(file.IONBF, size)
	case "line":
		e = f.Setvbuf(file.IOLBF, size)
	case "full":
		e = f.Setvbuf(file.IOFBF, size)
	default:
		return nil, ap.OptionError(1, mode)
	}

	return fileResult(th, e)
}

func FWrite(th object.Thread, args ...object.Value) ([]object.Value, *object.RuntimeError) {
	ap := fnutil.NewArgParser(th, args)

	ud, err := ap.ToFullUserdata(0)
	if err != nil {
		return nil, err
	}

	f, ok := ud.Value.(file.File)
	if !ok {
		return nil, ap.TypeError(0, "FILE*")
	}

	for i := range args[1:] {
		s, err := ap.ToGoString(i + 1)
		if err != nil {
			return nil, err
		}

		_, e := f.WriteString(s)
		if e != nil {
			return fileResult(th, e)
		}
	}

	return fileResult(th, nil)
}

func FToString(th object.Thread, args ...object.Value) ([]object.Value, *object.RuntimeError) {
	ap := fnutil.NewArgParser(th, args)

	ud, err := ap.ToFullUserdata(0)
	if err != nil {
		return nil, err
	}

	f, ok := ud.Value.(file.File)
	if !ok {
		return nil, ap.TypeError(0, "FILE*")
	}

	if f.IsClosed() {
		return []object.Value{object.String("closed file")}, nil
	}

	return []object.Value{object.String(fmt.Sprintf("file (%p)", f))}, nil
}
