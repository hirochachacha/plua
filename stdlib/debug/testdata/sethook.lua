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
