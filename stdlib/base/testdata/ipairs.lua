f, t, k =  ipairs({3, 4, 5})
assert(k == 0)
k, v = f(t, k)
assert(k == 1 and v == 3)
k, v = f(t, k)
assert(k == 2 and v == 4)
k, v = f(t, k)
assert(k == 3 and v == 5)
assert(f(t, k) == nil)

local t = setmetatable({}, {__index = function(t, k)
	if k < 5 then
		return 1
	end
end})

local i = 0
for k, v in ipairs(t) do
	i = i + 1
	assert(i == k and v == 1)
end

local t = setmetatable({}, {__ipairs = function(t)
	return function(t, k)
		if k == nil then
			return 1, 2
		end
		if k < 3 then
			return k+1, k+2
		end
	end, t, nil
end})

f, t, k = ipairs(t)
assert(k == nil)
k, v = f(t, k)
assert(k == 1 and v == 2)
k, v = f(t, k)
assert(k == 2 and v == 3)
k, v = f(t, k)
assert(k == 3 and v == 4)
