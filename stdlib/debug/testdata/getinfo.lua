

d = debug.getinfo(1)
assert(d.currentline == 3)
assert(d.source == "@testdata/getinfo.lua" or d.source == "@testdata\\getinfo.lua")
assert(d.what == "main")

function f()
	d = debug.getinfo(1)
	assert(d.currentline == 9)
	assert(d.name == "f")
	assert(d.what == "Lua")
	assert(d.namewhat == "global")
	assert(d.nparams == 0)
	assert(d.isvararg == false)
	assert(d.istailcall == false)
	assert(d.linedefined == 8)
	assert(d.lastlinedefined == 19)
end

d = debug.getinfo(f)

assert(d.currentline == -1)
assert(d.what == "Lua")
assert(d.nparams == 0)
assert(d.isvararg == false)
assert(d.istailcall == false)
assert(d.linedefined == 8)
assert(d.lastlinedefined == 19)

f()

d = debug.getinfo(0)

assert(d.namewhat == "field")
assert(d.linedefined == -1)
assert(d.what == "Go" or d.what == "C")
assert(d.isvararg == true)
assert(d.currentline == -1)
assert(d.source == "=[Go]" or d.source == "=[C]")
assert(d.name == "getinfo")
assert(d.istailcall == false)
assert(d.short_src == "[Go]" or d.short_src == "[C]")
assert(d.nups == 0)
assert(d.nparams == 0)
assert(d.lastlinedefined == -1)


local function x(a)
	y()
end

function y()
	d = debug.getinfo(2)
	assert(d.namewhat == "local")
	assert(d.linedefined == 49)
	assert(d.isvararg == false)
	assert(d.currentline == 50)
	assert(d.name == "x")
	assert(d.istailcall == false)
	assert(d.nups == 1)
	assert(d.nparams == 1)
	assert(d.lastlinedefined == 51)
end

x()

function x()
	y()
end

function y()
	return z()
end

function z()
	return w()
end

function w()
	d = debug.getinfo(2)

	assert(d.what == "Lua")
	assert(d.namewhat == "local")
	assert(d.name == "x")
	assert(d.nparams == 0)
	assert(d.isvararg == false)
	assert(d.nups == 1)
	assert(d.istailcall == false)
	assert(d.currentline == 69)
	assert(d.linedefined == 68)
	assert(d.lastlinedefined == 70)
end

x()

pcall(function()
	d = debug.getinfo(2)
end)

assert(d.name == "pcall")

xpcall(function()
	d1 = debug.getinfo(2)
	d2 = debug.getinfo(3)

	error("err")
end, function()
	d3 = debug.getinfo(2)
end)

assert(d1.name == "xpcall")
assert(d2.what == "main")
assert(d3.name == "error" or d3.name == "xpcall") -- plua returns "xpcall" here, but OK.

debug.sethook(function()
	assert(debug.getinfo(0).name == "getinfo")
	assert(debug.getinfo(1).namewhat == "hook" or debug.getinfo(1).namewhat == "") -- lua 5.3.3 returns "", but it should return "hook"
	assert(debug.getinfo(2).name == "sethook")
	assert(debug.getinfo(3).what == "main")
end, "c")

debug.sethook()

-- test tagmethod information
local a = {}
local function index(t)
  local info = debug.getinfo(1);
  assert(info.namewhat == "metamethod")
  a.op = info.name
  return info.name
end
setmetatable(a, {
  __index = index
})

assert(a[3] == "__index")
