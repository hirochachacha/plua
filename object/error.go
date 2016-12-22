package object

import (
	"fmt"
	"io"
	"os"
)

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

func PrintError(err error) error {
	return FprintError(os.Stderr, err)
}

func FprintError(w io.Writer, err error) error {
	return fprintError(w, err)
}

type errWriter struct {
	w   io.Writer
	err error
}

func (w *errWriter) Write(p []byte) (n int, err error) {
	if w.err == nil {
		n, w.err = w.w.Write(p)
	}
	return n, w.err
}

func fprintError(w io.Writer, err error) error {
	ew := &errWriter{w: w}
	if err, ok := err.(*RuntimeError); ok {
		fmt.Fprintln(ew, err)
		fmt.Fprint(ew, "stack traceback:")
		tb := err.Traceback
		if len(tb) <= 22 {
			for _, st := range tb {
				printStackTrace(ew, st)
			}
		} else {
			for _, st := range tb[:10] {
				printStackTrace(ew, st)
			}
			fmt.Fprint(ew, "\n\t")
			fmt.Fprint(ew, "...")
			for _, st := range tb[len(tb)-11:] {
				printStackTrace(ew, st)
			}
		}
		fmt.Fprintln(ew)
	} else {
		fmt.Fprintln(ew, err)
	}
	return ew.err
}

func printStackTrace(w io.Writer, st *StackTrace) {
	fmt.Fprint(w, "\n\t")

	var write bool

	if st.Source != "" {
		fmt.Fprint(w, st.Source)
		fmt.Fprint(w, ":")
		write = true
	}

	if st.Line > 0 {
		fmt.Fprint(w, st.Line)
		fmt.Fprint(w, ":")
		write = true
	}

	if write {
		fmt.Fprint(w, " in ")
	}

	fmt.Fprint(w, st.Signature)

	if st.IsTailCall {
		fmt.Fprint(w, "\n\t")
		fmt.Fprint(w, "(...tail calls...)")
	}
}
