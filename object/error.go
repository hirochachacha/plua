package object

import "fmt"

type StackTrace struct {
	Source     string
	Line       int
	Signature  string
	IsTailCall bool
}

type RuntimeError struct {
	RawValue  Value
	Level     int
	Traceback []*StackTrace
}

func NewRuntimeError(msg string) *RuntimeError {
	return &RuntimeError{RawValue: String(msg), Level: 1}
}

func (err *RuntimeError) Value() Value {
	if msg, ok := err.RawValue.(String); ok {
		if 0 < err.Level && err.Level < len(err.Traceback) {
			tb := err.Traceback[err.Level]
			if tb.Source != "[Go]" {
				return String(fmt.Sprintf("%s:%d: %s", tb.Source, tb.Line, msg))
			}
		}
	}
	return err.RawValue
}

func (err *RuntimeError) Error() string {
	return fmt.Sprintf("runtime: %s", Repr(err.Value()))
}
