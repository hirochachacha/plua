f, t, k = pairs({a = 1, b =2})
assert(k == nil)
k, v = f(t, k)
if v == 1 then
	assert(k == "a")
elseif v == 2 then
	assert(k == "b")
else
	assert(false)
end
k, v = f(t, k)
if v == 1 then
	assert(k == "a")
elseif v == 2 then
	assert(k == "b")
else
	assert(false)
end
k, v = f(t, k)
assert(k == nil and v == nil)

local t = setmetatable({}, {__pairs = function(t)
	return function(t, k)
		if k == nil then
			return 1, 2
		end
		if k < 3 then
			return k+1, k+2
		end
	end, t, nil
end})

f, t, k = pairs(t)
assert(k == nil)
k, v = f(t, k)
assert(k == 1 and v == 2)
k, v = f(t, k)
assert(k == 2 and v == 3)
k, v = f(t, k)
assert(k == 3 and v == 4)
