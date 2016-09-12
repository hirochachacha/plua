package position

import (
	"fmt"
	"strings"

	"github.com/hirochachacha/plua/internal/version"
)

var NoPos = Position{
	Line:   -1,
	Column: -1,
}

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
	case '=':
		s = s[1:]
		if len(s) > version.LUA_IDSIZE {
			return s[:version.LUA_IDSIZE]
		}
		return s
	case '@':
		s = s[1:]
		if len(s) > version.LUA_IDSIZE {
			return s[:version.LUA_IDSIZE-3] + "..."
		}
		return s
	default:
		i := strings.IndexRune(s, '\n')
		if i == -1 {
			s = "[string \"" + s

			if len(s) > version.LUA_IDSIZE-2 {
				return s[:version.LUA_IDSIZE-5] + "...\"]"
			}
			return s + "\"]"
		}

		s = "[string \"" + s[:i]

		if len(s) > version.LUA_IDSIZE-2 {
			return s[:version.LUA_IDSIZE-5] + "...\"]"
		}

		return s + "...\"]"
	}
}
