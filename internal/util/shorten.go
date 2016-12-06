package util

import (
	"strings"

	"github.com/hirochachacha/plua/internal/version"
)

func Shorten(s string) string {
	if len(s) == 0 {
		return ""
	}

	switch s[0] {
	case '=':
		s = s[1:]
		if len(s) >= version.LUA_IDSIZE {
			return s[:version.LUA_IDSIZE-1]
		}
		return s
	case '@':
		s = s[1:]
		if len(s) >= version.LUA_IDSIZE {
			return s[:version.LUA_IDSIZE-4] + "..."
		}
		return s
	default:
		i := strings.IndexRune(s, '\n')
		if i == -1 {
			s = "[string \"" + s

			if len(s) >= version.LUA_IDSIZE-3 {
				return s[:version.LUA_IDSIZE-6] + "...\"]"
			}
			return s + "\"]"
		}

		s = "[string \"" + s[:i]

		if len(s) >= version.LUA_IDSIZE-3 {
			return s[:version.LUA_IDSIZE-6] + "...\"]"
		}

		return s + "...\"]"
	}
}
