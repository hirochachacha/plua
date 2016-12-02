package io

import (
	"io/ioutil"
	"os"
	"os/exec"
	"strings"

	"github.com/hirochachacha/plua/internal/file"
	"github.com/hirochachacha/plua/object"
	"github.com/hirochachacha/plua/object/fnutil"
)

func Open(th object.Thread, args ...object.Value) ([]object.Value, *object.RuntimeError) {
	fileIndex := th.NewTableSize(0, 7)

	fileIndex.Set(object.String("close"), object.GoFunction(fclose))
	fileIndex.Set(object.String("flush"), object.GoFunction(fflush))
	fileIndex.Set(object.String("lines"), object.GoFunction(flines))
	fileIndex.Set(object.String("read"), object.GoFunction(fread))
	fileIndex.Set(object.String("seek"), object.GoFunction(fseek))
	fileIndex.Set(object.String("setvbuf"), object.GoFunction(fsetvbuf))
	fileIndex.Set(object.String("write"), object.GoFunction(fwrite))

	mt := th.NewTableSize(0, 3)

	mt.Set(object.String("__index"), fileIndex)
	mt.Set(object.String("__tostring"), object.GoFunction(ftostring))
	mt.Set(object.TM_NAME, object.String("FILE*"))

	stdin := &object.Userdata{
		Value:     file.NewFile(os.Stdin, os.O_RDONLY),
		Metatable: mt,
	}

	stdout := &object.Userdata{
		Value:     file.NewFile(os.Stdout, os.O_WRONLY),
		Metatable: mt,
	}

	stderr := &object.Userdata{
		Value:     file.NewFile(os.Stderr, os.O_WRONLY),
		Metatable: mt,
	}

	var _input = stdin
	var _output = stdout

	// close([file])
	var _close = func(th object.Thread, args ...object.Value) ([]object.Value, *object.RuntimeError) {
		if len(args) == 0 {
			return fileResult(th, _output.Value.(file.File).Close())
		}

		return fclose(th, args...)
	}

	var flush = func(th object.Thread, args ...object.Value) ([]object.Value, *object.RuntimeError) {
		if len(args) == 0 {
			return fileResult(th, _output.Value.(file.File).Flush())
		}

		return fflush(th, args...)
	}

	// input([file])
	var input = func(th object.Thread, args ...object.Value) ([]object.Value, *object.RuntimeError) {
		if len(args) == 0 {
			return []object.Value{_input}, nil
		}

		ap := fnutil.NewArgParser(th, args)

		var ud *object.Userdata
		var f file.File
		var e error

		if fname, err := ap.ToGoString(0); err == nil {
			f, e = file.OpenFile(fname, os.O_RDONLY, 0644)
			if e != nil {
				return fileResult(th, e)
			}

			ud = &object.Userdata{
				Value:     f,
				Metatable: mt,
			}
		} else {
			ud, e = ap.ToFullUserdata(0)
			if e != nil {
				return nil, ap.TypeError(0, "FILE* or string")
			}

			_, ok := ud.Value.(file.File)
			if !ok {
				return nil, ap.TypeError(0, "FILE*")
			}
		}

		_input = ud

		return []object.Value{ud}, nil
	}

	// lines([filename, ...])
	var lines = func(th object.Thread, args ...object.Value) ([]object.Value, *object.RuntimeError) {
		ap := fnutil.NewArgParser(th, args)

		fname, err := ap.ToGoString(0)
		if err != nil {
			return nil, err
		}

		f, e := file.OpenFile(fname, os.O_RDONLY, 0644)
		if e != nil {
			return fileResult(th, e)
		}

		retfn := func(_ object.Thread, _ ...object.Value) ([]object.Value, *object.RuntimeError) {
			return _read(th, args, f, 1)
		}

		return []object.Value{object.GoFunction(retfn)}, nil
	}

	// open(filename, [, mode])
	var open = func(th object.Thread, args ...object.Value) ([]object.Value, *object.RuntimeError) {
		ap := fnutil.NewArgParser(th, args)

		fname, err := ap.ToGoString(0)
		if err != nil {
			return nil, err
		}

		mode, err := ap.OptGoString(1, "r")
		if err != nil {
			return nil, err
		}

		var f file.File
		var e error

		switch mode {
		case "r":
			f, e = file.OpenFile(fname, os.O_RDONLY, 0644)
		case "w":
			f, e = file.OpenFile(fname, os.O_WRONLY|os.O_TRUNC|os.O_CREATE, 0644)
		case "a":
			f, e = file.OpenFile(fname, os.O_WRONLY|os.O_APPEND|os.O_CREATE, 0644)
		case "r+":
			f, e = file.OpenFile(fname, os.O_RDWR, 0644)
		case "w+":
			f, e = file.OpenFile(fname, os.O_RDWR|os.O_TRUNC|os.O_CREATE, 0644)
		case "a+":
			f, e = file.OpenFile(fname, os.O_RDWR|os.O_APPEND|os.O_CREATE, 0644)
		default:
			return nil, ap.OptionError(1, mode)
		}
		if e != nil {
			return fileResult(th, e)
		}

		ud := &object.Userdata{
			Value:     f,
			Metatable: mt,
		}

		return []object.Value{ud}, nil
	}

	// output([file])
	var output = func(th object.Thread, args ...object.Value) ([]object.Value, *object.RuntimeError) {
		if len(args) == 0 {
			return []object.Value{_output}, nil
		}

		ap := fnutil.NewArgParser(th, args)

		var ud *object.Userdata
		var f file.File
		var e error

		if fname, err := ap.ToGoString(0); err == nil {
			f, e = file.OpenFile(fname, os.O_WRONLY|os.O_TRUNC|os.O_CREATE, 0644)
			if e != nil {
				return fileResult(th, e)
			}

			ud = &object.Userdata{
				Value:     f,
				Metatable: mt,
			}
		} else {
			ud, e = ap.ToFullUserdata(0)
			if e != nil {
				return nil, ap.TypeError(0, "FILE* or string")
			}

			_, ok := ud.Value.(file.File)
			if !ok {
				return nil, ap.TypeError(0, "FILE*")
			}
		}

		_output = ud

		return []object.Value{ud}, nil
	}

	var popen = func(th object.Thread, args ...object.Value) ([]object.Value, *object.RuntimeError) {
		ap := fnutil.NewArgParser(th, args)

		prog, err := ap.ToGoString(0)
		if err != nil {
			return nil, err
		}

		mode, err := ap.OptGoString(1, "r")
		if err != nil {
			return nil, err
		}

		progArgs := strings.Fields(prog)

		cmd := exec.Command(progArgs[0], progArgs[1:]...)

		var ud *object.Userdata

		switch mode {
		case "r":
			r, e := cmd.StdoutPipe()
			if e != nil {
				return execResult(th, e)
			}

			f, ok := r.(*os.File)
			if !ok {
				return nil, object.NewRuntimeError("pipe is not *os.File")
			}

			ud = &object.Userdata{
				Value:     file.NewFile(f, os.O_RDONLY),
				Metatable: mt,
			}
		case "w":
			w, e := cmd.StdinPipe()
			if e != nil {
				return execResult(th, e)
			}

			f, ok := w.(*os.File)
			if !ok {
				return nil, object.NewRuntimeError("pipe is not *os.File")
			}

			ud = &object.Userdata{
				Value:     file.NewFile(f, os.O_WRONLY),
				Metatable: mt,
			}
		default:
			return nil, ap.OptionError(1, mode)
		}

		return []object.Value{ud}, nil
	}

	var read = func(th object.Thread, args ...object.Value) ([]object.Value, *object.RuntimeError) {
		return _read(th, args, _input.Value.(file.File), 0)
	}

	var tmpfile = func(th object.Thread, args ...object.Value) ([]object.Value, *object.RuntimeError) {
		f, err := ioutil.TempFile("", "plua")
		if err != nil {
			return fileResult(th, err)
		}

		ud := &object.Userdata{
			Value:     file.NewFile(f, os.O_RDWR),
			Metatable: mt,
		}

		return []object.Value{ud}, nil
	}

	var _type = func(th object.Thread, args ...object.Value) ([]object.Value, *object.RuntimeError) {
		ap := fnutil.NewArgParser(th, args)

		ud, err := ap.ToFullUserdata(0)
		if err != nil {
			return nil, nil
		}

		f, ok := ud.Value.(file.File)
		if !ok {
			return nil, nil
		}

		if f.IsClosed() {
			return []object.Value{object.String("closed file")}, nil
		}

		return []object.Value{object.String("file")}, nil
	}

	var write = func(th object.Thread, args ...object.Value) ([]object.Value, *object.RuntimeError) {
		f := _output.Value.(file.File)

		ap := fnutil.NewArgParser(th, args)

		for i := range args {
			s, err := ap.ToGoString(i)
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

	m := th.NewTableSize(0, 14)

	m.Set(object.String("stdin"), stdin)
	m.Set(object.String("stdout"), stdout)
	m.Set(object.String("stderr"), stderr)

	m.Set(object.String("close"), object.GoFunction(_close))
	m.Set(object.String("flush"), object.GoFunction(flush))
	m.Set(object.String("input"), object.GoFunction(input))
	m.Set(object.String("lines"), object.GoFunction(lines))
	m.Set(object.String("open"), object.GoFunction(open))
	m.Set(object.String("output"), object.GoFunction(output))
	m.Set(object.String("popen"), object.GoFunction(popen))
	m.Set(object.String("read"), object.GoFunction(read))
	m.Set(object.String("tmpfile"), object.GoFunction(tmpfile))
	m.Set(object.String("type"), object.GoFunction(_type))
	m.Set(object.String("write"), object.GoFunction(write))

	return []object.Value{m}, nil
}
