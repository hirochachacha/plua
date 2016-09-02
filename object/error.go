package object

import "github.com/hirochachacha/plua/position"

type RuntimeError struct {
	Value Value
	Level int
	Pos   position.Position
}

func NewRuntimeError(msg string) *RuntimeError {
	return &RuntimeError{Value: String(msg)}
}

func (err *RuntimeError) Positioned() Value {
	if msg, ok := ToGoString(err.Value); ok {
		if err.Pos.IsValid() {
			msg = err.Pos.String() + ": " + msg
		}
		return String(msg)
	}
	return err.Value
}

func (err *RuntimeError) Error() string {
	msg := Repr(err.Value)
	if err.Pos.IsValid() {
		msg = msg + " raised from " + err.Pos.String()
	}
	return "object: " + msg
}
