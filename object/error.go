package object

import (
	"strings"

	"github.com/hirochachacha/plua/internal/strconv"
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
			traceback := err.Traceback[0]
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
	if val, ok := err.Value.(String); ok {
		msg = strconv.Quote(string(val))
	}
	if len(err.Traceback) > 0 {
		msg = msg + " from " + err.Traceback[0].String()
		if len(err.Traceback) > 1 {
			msg += " via "
			for _, tb := range err.Traceback[1:] {
				msg += tb.String() + ", "
			}
			if strings.HasSuffix(msg, ", ") {
				msg = msg[:len(msg)-2]
			}
		}
	}
	return "runtime: " + msg
}
