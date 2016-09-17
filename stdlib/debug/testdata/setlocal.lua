function x()
	local a = 4
	y()
	assert(a == 10)
end

function y()
	debug.setlocal(2, 1, 10)
end

x()

function x(...)
	debug.setlocal(1, -1, 15)
	assert(... == 15)
end

x(10)
