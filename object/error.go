package object

import "github.com/hirochachacha/plua/position"

type RuntimeError struct {
	Value     Value
	Level     int
	Traceback []position.Position
}

func NewRuntimeError(msg string) *RuntimeError {
	return &RuntimeError{Value: String(msg)}
}

func (err *RuntimeError) Positioned() Value {
	if msg, ok := err.Value.(String); ok {
		if len(err.Traceback) > 0 {
			traceback := err.Traceback[len(err.Traceback)-1]
			if traceback.IsValid() {
				msg = String(traceback.String()) + ": " + msg
			}
		}
		return msg
	}
	return err.Value
}

func (err *RuntimeError) Error() string {
	msg := Repr(err.Value)
	if len(err.Traceback) > 0 {
		msg = msg + " raised from "
		for _, tb := range err.Traceback {
			if tb.IsValid() {
				msg += tb.String() + ", "
			}
		}
	}
	return "runtime: " + msg
}
