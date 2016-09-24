function f()
	a = 1
	for i = 0, 0 do
		a = a + i
	end

	return a + 10
end

lines = {22, 2, 3, 4, 3, 7, 24}

function test(f)
	local i = 1
	local function hook(event, line)
		assert(debug.getinfo(1).namewhat == "hook")
		assert(line == lines[i])
		i = i + 1
	end

	debug.sethook(hook, "l")

	f()

	debug.sethook()
end

test(f)

t = {
	{name = "sethook", namewhat = "field", linedefined = -1},
	{name = "foo", namewhat = "global", linedefined = 34},
	{what = "main", linedefined = 0},
}

local i = 1

function foo() end

debug.sethook(function()
	d = debug.getinfo(2)
	tt = t[i]
	if tt ~= nil then
		for k, v in pairs(tt) do
			assert(v, d[k])
		end
		i = i + 1
	end
end, "r")

foo()

debug.sethook()
