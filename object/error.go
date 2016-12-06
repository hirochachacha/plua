package object

import (
	"fmt"

	"github.com/hirochachacha/plua/position"
)

type RuntimeError struct {
	Value     Value
	Level     int
	Traceback []position.Position
}

func NewRuntimeError(msg string) *RuntimeError {
	return &RuntimeError{Value: String(msg), Level: 1}
}

func (err *RuntimeError) Positioned() Value {
	if msg, ok := err.Value.(String); ok {
		if len(err.Traceback) > 0 {
			return String(fmt.Sprintf("%s: %s", err.Traceback[0], msg))
		}
		return msg
	}
	return err.Value
}

func (err *RuntimeError) Error() string {
	return fmt.Sprintf("runtime: %s", Repr(err.Positioned()))
}
