package os

import (
	"io/ioutil"
	"os"
	"os/exec"
	"strings"
	"time"

	"github.com/hirochachacha/plua/object"
	"github.com/hirochachacha/plua/object/fnutil"
)

var (
	startTime = time.Now()
	location  = startTime.Local().Location()
)

func clock(th object.Thread, args ...object.Value) ([]object.Value, *object.RuntimeError) {
	endTime := time.Now()
	diffSecond := float64(endTime.Unix()-startTime.Unix()) / 1000

	return []object.Value{object.Number(diffSecond)}, nil
}

// date([format [, time]])
func date(th object.Thread, args ...object.Value) ([]object.Value, *object.RuntimeError) {
	ap := fnutil.NewArgParser(th, args)

	format, err := ap.OptGoString(0, "%c")
	if err != nil {
		return nil, err
	}

	sec, err := ap.OptGoInt64(1, time.Now().Unix())
	if err != nil {
		return nil, err
	}

	t := time.Unix(sec, 0)

	if len(format) == 0 {
		return []object.Value{object.String("")}, nil
	}

	if format[0] == '!' {
		format = format[1:]
	} else {
		t = t.Local()
	}

	if format == "*t" {
		return []object.Value{dateTable(th, t)}, nil
	}

	s, err := dateFormat(th, format, t)
	if err != nil {
		return nil, err
	}

	return []object.Value{object.String(s)}, nil
}

func difftime(th object.Thread, args ...object.Value) ([]object.Value, *object.RuntimeError) {
	ap := fnutil.NewArgParser(th, args)

	t2, err := ap.ToGoInt64(0)
	if err != nil {
		return nil, err
	}

	t1, err := ap.ToGoInt64(1)
	if err != nil {
		return nil, err
	}

	return []object.Value{object.Integer(t2 - t1)}, nil
}

func execute(th object.Thread, args ...object.Value) ([]object.Value, *object.RuntimeError) {
	if len(args) == 0 {
		return []object.Value{object.True}, nil
	}

	ap := fnutil.NewArgParser(th, args)

	prog, err := ap.ToGoString(0)
	if err != nil {
		return nil, err
	}

	progArgs := strings.Fields(prog)

	cmd := exec.Command(progArgs[0], progArgs[1:]...)
	cmd.Stdin = os.Stdin
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	return execResult(th, cmd.Run())
}

// close is not supportted.
// exit([code [, close]])
func exit(th object.Thread, args ...object.Value) ([]object.Value, *object.RuntimeError) {
	if len(args) == 0 {
		os.Exit(0)
	}

	ap := fnutil.NewArgParser(th, args)

	i, err := ap.ToGoInt(0)
	if err == nil {
		os.Exit(i)
	}

	val, _ := ap.Get(0)

	if b, ok := val.(object.Boolean); ok {
		if b {
			os.Exit(1)
		}

		os.Exit(0)
	}

	return nil, ap.TypeError(0, "integer or boolean")
}

// getenv(varname)
func getenv(th object.Thread, args ...object.Value) ([]object.Value, *object.RuntimeError) {
	ap := fnutil.NewArgParser(th, args)

	key, err := ap.ToGoString(0)
	if err != nil {
		return nil, err
	}

	val := os.Getenv(key)

	if val == "" {
		return []object.Value{nil}, nil
	}

	return []object.Value{object.String(val)}, nil
}

// remove(filename)
func remove(th object.Thread, args ...object.Value) ([]object.Value, *object.RuntimeError) {
	ap := fnutil.NewArgParser(th, args)

	name, err := ap.ToGoString(0)
	if err != nil {
		return nil, err
	}

	return fileResult(th, os.Remove(name))
}

// rename(oldname, newname)
func rename(th object.Thread, args ...object.Value) ([]object.Value, *object.RuntimeError) {
	ap := fnutil.NewArgParser(th, args)

	old, err := ap.ToGoString(0)
	if err != nil {
		return nil, err
	}

	_new, err := ap.ToGoString(1)
	if err != nil {
		return nil, err
	}

	return fileResult(th, os.Rename(old, _new))
}

// setlocale(locale [, category])
func setlocale(th object.Thread, args ...object.Value) ([]object.Value, *object.RuntimeError) {
	return nil, object.NewRuntimeError("not implemented")
}

func _time(th object.Thread, args ...object.Value) ([]object.Value, *object.RuntimeError) {
	if len(args) == 0 {
		return []object.Value{object.Integer(time.Now().Unix())}, nil
	}

	ap := fnutil.NewArgParser(th, args)

	t, err := ap.ToTable(0)
	if err != nil {
		return nil, err
	}

	day, ok := object.ToGoInt(t.Get(object.String("day")))
	if !ok {
		return nil, object.NewRuntimeError("field 'day' missing in date table")
	}
	month, ok := object.ToGoInt(t.Get(object.String("month")))
	if !ok {
		return nil, object.NewRuntimeError("field 'month' missing in date table")
	}
	year, ok := object.ToGoInt(t.Get(object.String("year")))
	if !ok {
		return nil, object.NewRuntimeError("field 'year' missing in date table")
	}

	sec, _ := object.ToGoInt(t.Get(object.String("sec")))
	min, _ := object.ToGoInt(t.Get(object.String("min")))
	hour, ok := object.ToGoInt(t.Get(object.String("hour")))
	if !ok {
		hour = 12
	}

	unix := time.Date(year, time.Month(month), day, hour, min, sec, 0, location).Unix()

	return []object.Value{object.Integer(unix)}, nil
}

func tmpname(th object.Thread, args ...object.Value) ([]object.Value, *object.RuntimeError) {
	f, err := ioutil.TempFile("", "plua")
	if err != nil {
		return nil, object.NewRuntimeError(err.Error())
	}

	err = f.Close()
	if err != nil {
		return nil, object.NewRuntimeError(err.Error())
	}

	err = os.Remove(f.Name())
	if err != nil {
		return nil, object.NewRuntimeError(err.Error())
	}

	return []object.Value{object.String(f.Name())}, nil
}

func Open(th object.Thread, args ...object.Value) ([]object.Value, *object.RuntimeError) {
	m := th.NewTableSize(0, 11)

	m.Set(object.String("clock"), object.GoFunction(clock))
	m.Set(object.String("date"), object.GoFunction(date))
	m.Set(object.String("difftime"), object.GoFunction(difftime))
	m.Set(object.String("execute"), object.GoFunction(execute))
	m.Set(object.String("exit"), object.GoFunction(exit))
	m.Set(object.String("getenv"), object.GoFunction(getenv))
	m.Set(object.String("remove"), object.GoFunction(remove))
	m.Set(object.String("rename"), object.GoFunction(rename))
	m.Set(object.String("setlocale"), object.GoFunction(setlocale))
	m.Set(object.String("time"), object.GoFunction(_time))
	m.Set(object.String("tmpname"), object.GoFunction(tmpname))

	return []object.Value{m}, nil
}
