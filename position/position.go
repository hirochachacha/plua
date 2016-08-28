package position

import (
	"fmt"
	"strings"
)

var NoPos Position

type Position struct {
	Filename string
	Line     int
	Column   int
}

func (pos Position) LessThan(other Position) bool {
	switch {
	case pos.Line < other.Line:
		return true
	case pos.Line > other.Line:
		return false
	default:
		return pos.Column < other.Column
	}
}

func (pos Position) Offset(s string) Position {
	newpos := pos
	for _, r := range s {
		if r == '\n' {
			newpos.Line++
			newpos.Column = 0
		} else {
			newpos.Column++
		}
	}
	return newpos
}

func (pos Position) OffsetColumn(off int) Position {
	newpos := pos
	newpos.Column += off
	return newpos
}

func (pos Position) String() string {
	s := shorten(pos.Filename)

	if pos.IsValid() {
		if s != "" {
			s += ":"
		}
		if pos.Column > 0 {
			s += fmt.Sprintf("%d:%d", pos.Line, pos.Column)
		} else {
			s += fmt.Sprint(pos.Line)
		}
	}
	if s == "" {
		s = "-"
	}
	return s
}

func (pos Position) IsValid() bool {
	return pos.Line > 0
}

func shorten(s string) string {
	if len(s) == 0 {
		return ""
	}

	switch s[0] {
	case '=', '@':
		return s[1:]
	}

	i := strings.IndexRune(s, '\n')
	if i == -1 {
		return "[string \"" + s + "\""
	}
	return "[string \"" + s[:i] + "\""
}
