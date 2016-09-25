local a,b,c = 1,2,3
local function foo(a) b = a; return c end
assert(debug.getupvalue(foo, 0) == nil)
k, v = debug.getupvalue(foo, 1)
assert(k == "b" and v == 2)
k, v = debug.getupvalue(foo, 2)
assert(k == "c" and v == 3)
assert(debug.getupvalue(foo, 3) == nil)
