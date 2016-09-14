

d = debug.getinfo(1)
assert(d.currentline == 3)
assert(d.source == "@testdata/getinfo.lua")
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
