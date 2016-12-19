package debug

import (
	"bytes"
	"strconv"

	"github.com/hirochachacha/plua/object"
)

func getTraceback(th object.Thread, msg string, level int) string {
	buf := new(bytes.Buffer)

	if msg != "" {
		buf.WriteString(msg)
		buf.WriteByte('\n')
	}

	buf.WriteString("stack traceback:")

	for _, tb := range th.Traceback(level) {
		buf.WriteString("\n\t")

		var write bool

		if tb.Source != "" {
			buf.WriteString(tb.Source)
			buf.WriteByte(':')
			write = true
		}

		if tb.Line > 0 {
			buf.WriteString(strconv.Itoa(tb.Line))
			buf.WriteByte(':')
			write = true
		}

		if write {
			buf.WriteString(" in ")
		}

		buf.WriteString(tb.Signature)
	}

	return buf.String()
}
