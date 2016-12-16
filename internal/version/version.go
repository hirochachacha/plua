package version

const (
	LUA_SIGNATURE = "Lua"

	LUAC_VERSION = 0x53
	LUAC_FORMAT  = 0

	LUAC_DATA = "\x19\x93\r\n\x1a\n"
	LUAC_INT  = 0x5678
	LUAC_NUM  = 370.5

	LUA_ENV = "_ENV"

	// list field per flush
	LUA_FPF = 50

	// max size of short version of source name
	LUA_IDSIZE = 60

	LUA_MAJOR_VERSION = "5"
	LUA_MINOR_VERSION = "3"
	LUA_VERSION       = LUA_MAJOR_VERSION + "." + LUA_MINOR_VERSION
	LUA_NAME          = "Lua " + LUA_VERSION

	MAX_TAG_LOOP     = 2000
	MAX_VM_RECURSION = 200

	MAXUPVAL = 255
	MAXVAR   = 200

	MAXUNPACK = 2000
	MAXMOVE   = 2000
)
