function f()
	a = 1
	for i = 0, 0 do
		a = a + i
	end

	return a + 10
end

lines = {21, 2, 3, 4, 3, 7, 23}

function test(f)
	local i = 1
	local function hook(event, line)
		assert(line == lines[i])
		i = i + 1
	end

	debug.sethook(hook, "l")

	f()

	debug.sethook()
end

test(f)
