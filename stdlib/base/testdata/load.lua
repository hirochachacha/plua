assert(load("return 1")() == 1)

f = assert(function() end)
assert(f() == nil)
