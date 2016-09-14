function x()
	local a = 4
	y()
	assert(a == 10)
end

function y()
	debug.setlocal(2, 1, 10)
end

x()
