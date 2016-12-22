local function h()
	tb = debug.traceback("", 2)
end

local function g()
	return h()
end

local function f()
	g()
end

f()

iter = string.gmatch(tb, ".-\n")
assert(iter() == "\n")
assert(iter() == "stack traceback:\n")
assert(iter() == "	testdata/traceback.lua:10: in local 'f'\n")
