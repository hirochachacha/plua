package object

import (
	"fmt"
	"io"
)

type StackTrace struct {
	Source    string
	Line      int
	Signature string
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

func PrintError(w io.Writer, err error) {
	if err, ok := err.(*RuntimeError); ok {
		fmt.Fprintln(w, err)
		fmt.Fprint(w, "stack traceback:")
		for _, tb := range err.Traceback {
			fmt.Fprint(w, "\n\t")

			var write bool

			if tb.Source != "" {
				fmt.Fprint(w, tb.Source)
				fmt.Fprint(w, ":")
				write = true
			}

			if tb.Line > 0 {
				fmt.Fprint(w, tb.Line)
				fmt.Fprint(w, ":")
				write = true
			}

			if write {
				fmt.Fprint(w, " in ")
			}

			fmt.Fprint(w, tb.Signature)
		}
		fmt.Fprintln(w)
	} else {
		fmt.Fprintln(w, err)
	}
}
