// +build !windows

package load

import "github.com/hirochachacha/plua/internal/version"

const (
	root        = "/usr/local/"
	ldir        = root + "share/lua/" + version.LUA_VERSION + "/"
	defaultPath = ldir + "?.lua;" + ldir + "?/init.lua;" + "./?.lua;" + "./?/init.lua"

	dsep   = "/"
	psep   = ";"
	mark   = "?"
	edir   = "!"
	ignore = "-"

	config = dsep + "\n" + psep + "\n" + mark + "\n" + edir + "\n" + ignore + "\n"
)
