package os

import (
	"os"
	"os/exec"
	"syscall"

	"github.com/hirochachacha/plua/object"
)

func fileResult(th object.Thread, err error) ([]object.Value, *object.RuntimeError) {
	if err == nil {
		return []object.Value{object.True}, nil
	}

	switch err {
	case os.ErrInvalid:
		return []object.Value{nil, object.String(err.Error()), object.Integer(syscall.EINVAL)}, nil
	case os.ErrPermission:
		return []object.Value{nil, object.String(err.Error()), object.Integer(syscall.EPERM)}, nil
	case os.ErrExist:
		return []object.Value{nil, object.String(err.Error()), object.Integer(syscall.EEXIST)}, nil
	case os.ErrNotExist:
		return []object.Value{nil, object.String(err.Error()), object.Integer(syscall.ENOENT)}, nil
	default:
		switch err := err.(type) {
		case *os.PathError:
			if errno, ok := err.Err.(syscall.Errno); ok {
				return []object.Value{nil, object.String(err.Error()), object.Integer(errno)}, nil
			}
		case *os.LinkError:
			if errno, ok := err.Err.(syscall.Errno); ok {
				return []object.Value{nil, object.String(err.Error()), object.Integer(errno)}, nil
			}
		}
	}

	return []object.Value{nil, object.String(err.Error()), object.Integer(-1)}, nil
}

func execResult(th object.Thread, err error) ([]object.Value, *object.RuntimeError) {
	if err == nil {
		return []object.Value{object.True, object.String("exit"), object.Integer(0)}, nil
	}

	if exitErr, ok := err.(*exec.ExitError); ok {
		if status, ok := exitErr.Sys().(syscall.WaitStatus); ok {
			if status.Signaled() {
				return []object.Value{object.False, object.String("signal"), object.Integer(status.ExitStatus())}, nil
			}
			return []object.Value{object.False, object.String("exit"), object.Integer(status.ExitStatus())}, nil
		}
		return []object.Value{object.False, object.String("exit"), object.Integer(-1)}, nil
	}

	if eErr, ok := err.(*exec.Error); ok {
		switch eErr.Err {
		case exec.ErrNotFound:
			return []object.Value{object.False, object.String("exit"), object.Integer(syscall.ENOENT)}, nil
		case os.ErrPermission:
			return []object.Value{object.False, object.String("exit"), object.Integer(syscall.EPERM)}, nil
		case os.ErrNotExist:
			return []object.Value{object.False, object.String("exit"), object.Integer(syscall.ENOENT)}, nil
		}
	}

	return []object.Value{object.False, object.String("exit"), object.Integer(-1)}, nil
}
