package object

import "fmt"

type StackTrace struct {
	Source     string
	Line       int
	Signature  string
	IsTailCall bool
}

type RuntimeError struct {
	Value     Value
	Level     int
	Traceback []*StackTrace
}

func NewRuntimeError(msg string) *RuntimeError {
	return &RuntimeError{Value: String(msg), Level: 1}
}

func (err *RuntimeError) Positioned() Value {
	if msg, ok := err.Value.(String); ok {
		if 0 < err.Level && err.Level < len(err.Traceback) {
			tb := err.Traceback[err.Level]
			if tb.Source != "[Go]" {
				return String(fmt.Sprintf("%s:%d: %s", tb.Source, tb.Line, msg))
			}
		}
		return msg
	}
	return err.Value
}

func (err *RuntimeError) Error() string {
	return fmt.Sprintf("runtime: %s", Repr(err.Positioned()))
}
