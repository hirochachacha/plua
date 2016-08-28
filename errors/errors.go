package errors

import (
	"fmt"

	"github.com/hirochachacha/blua/position"
)

type ErrorClass struct {
	name string
}

func newClass(name string) *ErrorClass {
	return &ErrorClass{name: name}
}

func (e *ErrorClass) New(msg string, args ...interface{}) *Error {
	return &Error{
		class:  e,
		format: msg,
		args:   args,
	}
}

func (e *ErrorClass) Wrap(err error) *Error {
	if err == nil {
		return nil
	}

	return &Error{
		class: e,
		err:   err,
	}
}

func (e *ErrorClass) WrapWith(pos position.Position, err error) *Error {
	return &Error{
		class: e,
		pos:   pos,
		err:   err,
	}
}

type Error struct {
	class  *ErrorClass
	pos    position.Position
	format string
	args   []interface{}
	err    error
}

func (e *Error) With(pos position.Position, args ...interface{}) *Error {
	return &Error{
		class:  e.class,
		pos:    pos,
		format: e.format,
		args:   args,
		err:    e.err,
	}
}

func (e *Error) Error() string {
	if e.pos.Filename != "" || e.pos.IsValid() {
		return e.pos.String() + ": " + e.message()
	}
	return e.message()
}

func (e *Error) message() string {
	if e.format != "" {
		return e.class.name + ":" + fmt.Sprintf(e.format, e.args...)
	}
	if e.err != nil {
		return e.class.name + ":" + e.err.Error()
	}

	return e.class.name
}

func (e *Error) Is(ec *ErrorClass) bool {
	return e.class == ec
}
