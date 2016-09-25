local a,b,c = 1,2,3
local function foo(a) b = a; return c end
assert(debug.setupvalue(foo, 0, 1) == nil)
k = debug.setupvalue(foo, 1, 3)
assert(k == "b")
k, v = debug.getupvalue(foo, 1)
assert(k == "b" and v == 3)
k = debug.setupvalue(foo, 2, 4)
assert(k == "c")
k, v = debug.getupvalue(foo, 2)
assert(k == "c" and v == 4)
assert(debug.getupvalue(foo, 3, 4) == nil)
