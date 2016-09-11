package object

import (
	"strings"

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
		tb := err.Traceback[0]
		if tb.IsValid() {
			msg += tb.String()
		}
		if len(err.Traceback) > 1 {
			var valid bool

			for _, tb := range err.Traceback[1:] {
				if tb.IsValid() {
					valid = true
					break
				}
			}

			if valid {
				msg += " via "
				for _, tb := range err.Traceback[1:] {
					if tb.IsValid() {
						msg += tb.String() + ", "
					}
				}
				if strings.HasSuffix(msg, ", ") {
					msg = msg[:len(msg)-2]
				}
			}
		}
	}
	return "runtime: " + msg
}
