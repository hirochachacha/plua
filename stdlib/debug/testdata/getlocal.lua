function f(a)
	local b = 1

	assert(debug.getlocal(1, 1) == 'a')
	b, v = debug.getlocal(1, 2)
	assert(b == 'b' and v == 1)
end

assert(debug.getlocal(f, 1) == 'a')

f()

function f2(...)
	k, v = debug.getlocal(1, -1)
	assert(k == "(*vararg)" and v == 1)

	k, v = debug.getlocal(1, -2)
	assert(k == "(*vararg)" and v == 2)

	k, v = debug.getlocal(1, -3)
	assert(k == "(*vararg)" and v == 3)

	assert(debug.getlocal(1, -4) == nil)

	local _ = ...
end

f2(1, 2, 3)

function x()
	local a = 4
	y()
end

function y()
	k, v = debug.getlocal(2, 1)
	assert(k == "a" and v == 4)
end

x()
