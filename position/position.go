package position

import (
	"fmt"

	"github.com/hirochachacha/plua/internal/util"
)

var NoPos = Position{
	Line:   -1,
	Column: -1,
}

type Position struct {
	SourceName string
	Line       int
	Column     int
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
	s := util.Shorten(pos.SourceName)
	if s == "" {
		s = "?"
	}
	line := pos.Line
	if !pos.IsValid() {
		line = -1
	}
	return fmt.Sprintf("%s:%d", s, line)
}

func (pos Position) IsValid() bool {
	return pos.Line > 0
}
