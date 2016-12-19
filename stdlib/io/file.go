package io

import (
	"fmt"
	"io"

	"github.com/hirochachacha/plua/internal/file"
	"github.com/hirochachacha/plua/internal/version"
	"github.com/hirochachacha/plua/object"
	"github.com/hirochachacha/plua/object/fnutil"
)

func toFile(ap *fnutil.ArgParser, n int) (file.File, *object.RuntimeError) {
	ud, err := ap.ToFullUserdata(n)
	if err != nil {
		return nil, ap.TypeError(n, "FILE*")
	}

	f, ok := ud.Value.(file.File)
	if !ok {
		return nil, ap.TypeError(n, "FILE*")
	}

	if f.IsClosed() {
		return nil, object.NewRuntimeError("attempt to use a closed file")
	}

	return f, nil
}

func fclose(th object.Thread, args ...object.Value) ([]object.Value, *object.RuntimeError) {
	ap := fnutil.NewArgParser(th, args)

	f, err := toFile(ap, 0)
	if err != nil {
		return nil, err
	}

	return fileResult(th, f.Close())
}

func fflush(th object.Thread, args ...object.Value) ([]object.Value, *object.RuntimeError) {
	ap := fnutil.NewArgParser(th, args)

	f, err := toFile(ap, 0)
	if err != nil {
		return nil, err
	}

	return fileResult(th, f.Flush())
}

func flines(th object.Thread, args ...object.Value) ([]object.Value, *object.RuntimeError) {
	ap := fnutil.NewArgParser(th, args)

	f, err := toFile(ap, 0)
	if err != nil {
		return nil, err
	}

	if len(args)-1 > version.MAXARGLINE {
		return nil, object.NewRuntimeError("too many arguments")
	}

	fnargs := append([]object.Value{}, args...)

	retfn := func(_ object.Thread, _ ...object.Value) ([]object.Value, *object.RuntimeError) {
		return _read(th, fnargs, f, 1, false, true)
	}

	return []object.Value{object.GoFunction(retfn)}, nil
}

func fread(th object.Thread, args ...object.Value) ([]object.Value, *object.RuntimeError) {
	ap := fnutil.NewArgParser(th, args)

	f, err := toFile(ap, 0)
	if err != nil {
		return nil, err
	}

	return _read(th, args, f, 1, false, false)
}

func fseek(th object.Thread, args ...object.Value) ([]object.Value, *object.RuntimeError) {
	ap := fnutil.NewArgParser(th, args)

	f, err := toFile(ap, 0)
	if err != nil {
		return nil, err
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
		n, e = f.Seek(offset, io.SeekStart)
	case "cur":
		n, e = f.Seek(offset, io.SeekCurrent)
	case "end":
		n, e = f.Seek(offset, io.SeekEnd)
	default:
		return nil, ap.OptionError(1, whence)
	}

	if e != nil {
		return fileResult(th, e)
	}

	return []object.Value{object.Integer(n)}, nil
}

func fsetvbuf(th object.Thread, args ...object.Value) ([]object.Value, *object.RuntimeError) {
	ap := fnutil.NewArgParser(th, args)

	f, err := toFile(ap, 0)
	if err != nil {
		return nil, err
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

func fwrite(th object.Thread, args ...object.Value) ([]object.Value, *object.RuntimeError) {
	ap := fnutil.NewArgParser(th, args)

	f, err := toFile(ap, 0)
	if err != nil {
		return nil, err
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

	return []object.Value{args[0]}, nil
}

func ftostring(th object.Thread, args ...object.Value) ([]object.Value, *object.RuntimeError) {
	ap := fnutil.NewArgParser(th, args)

	ud, err := ap.ToFullUserdata(0)
	if err != nil {
		return nil, ap.TypeError(0, "FILE*")
	}

	f, ok := ud.Value.(file.File)
	if !ok {
		return nil, ap.TypeError(0, "FILE*")
	}

	if f.IsClosed() {
		return []object.Value{object.String("file (closed)")}, nil
	}

	return []object.Value{object.String(fmt.Sprintf("file (%p)", f))}, nil
}
