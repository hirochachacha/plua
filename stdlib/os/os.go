package os

import (
	"io/ioutil"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"time"

	"github.com/hirochachacha/plua/object"
	"github.com/hirochachacha/plua/object/fnutil"
)

var (
	startTime = time.Now()
	location  = startTime.Local().Location()
)

func Clock(th object.Thread, args ...object.Value) ([]object.Value, *object.RuntimeError) {
	endTime := time.Now()
	diffSecond := float64(endTime.Unix()-startTime.Unix()) / 1000

	return []object.Value{object.Number(diffSecond)}, nil
}

// date([format [, time]])
func Date(th object.Thread, args ...object.Value) ([]object.Value, *object.RuntimeError) {
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

func DiffTime(th object.Thread, args ...object.Value) ([]object.Value, *object.RuntimeError) {
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

func Execute(th object.Thread, args ...object.Value) ([]object.Value, *object.RuntimeError) {
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
func Exit(th object.Thread, args ...object.Value) ([]object.Value, *object.RuntimeError) {
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
func GetEnv(th object.Thread, args ...object.Value) ([]object.Value, *object.RuntimeError) {
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
func Remove(th object.Thread, args ...object.Value) ([]object.Value, *object.RuntimeError) {
	ap := fnutil.NewArgParser(th, args)

	name, err := ap.ToGoString(0)
	if err != nil {
		return nil, err
	}

	return fileResult(th, os.Remove(name))
}

// rename(oldname, newname)
func Rename(th object.Thread, args ...object.Value) ([]object.Value, *object.RuntimeError) {
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
func SetLocale(th object.Thread, args ...object.Value) ([]object.Value, *object.RuntimeError) {
	return nil, object.NewRuntimeError("not implemented")
}

func Time(th object.Thread, args ...object.Value) ([]object.Value, *object.RuntimeError) {
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

func TmpName(th object.Thread, args ...object.Value) ([]object.Value, *object.RuntimeError) {
	f, err := ioutil.TempFile("", "plua")
	if err != nil {
		return nil, object.NewRuntimeError(err.Error())
	}

	path := filepath.Join(os.TempDir(), f.Name())

	err = f.Close()
	if err != nil {
		return nil, object.NewRuntimeError(err.Error())
	}

	err = os.Remove(path)
	if err != nil {
		return nil, object.NewRuntimeError(err.Error())
	}

	return []object.Value{object.String(path)}, nil
}

func Open(th object.Thread, args ...object.Value) ([]object.Value, *object.RuntimeError) {
	m := th.NewTableSize(0, 11)

	m.Set(object.String("clock"), object.GoFunction(Clock))
	m.Set(object.String("date"), object.GoFunction(Date))
	m.Set(object.String("difftime"), object.GoFunction(DiffTime))
	m.Set(object.String("execute"), object.GoFunction(Execute))
	m.Set(object.String("exit"), object.GoFunction(Exit))
	m.Set(object.String("getenv"), object.GoFunction(GetEnv))
	m.Set(object.String("remove"), object.GoFunction(Remove))
	m.Set(object.String("rename"), object.GoFunction(Rename))
	m.Set(object.String("setlocale"), object.GoFunction(SetLocale))
	m.Set(object.String("time"), object.GoFunction(Time))
	m.Set(object.String("tmpname"), object.GoFunction(TmpName))

	return []object.Value{m}, nil
}
