ok, msg = pcall(error, "test", 1)
assert(not ok and msg == "test")
ok, msg = pcall(error, "test", 2)
assert(not ok and (msg == "testdata/error.lua:3: test" or msg == "testdata\\error.lua:3: test"))
ok, msg = pcall(error, "test", 3)
assert(not ok and msg == "test")

local function h()
	error("test", 2)
end

local function g()
	return h()
end

local function f()
	g()
end

ok, msg = pcall(f)
assert(not ok and (msg == "testdata/error.lua:17: test" or msg == "testdata\\error.lua:17: test"))
