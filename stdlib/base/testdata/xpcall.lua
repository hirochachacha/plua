function f() error("error") end
function g(x) return x end
ok, err = xpcall(f, g)
assert(not ok and err == "testdata/xpcall.lua:1: error")

function g2(x) error(x) end
ok, err = xpcall(f, g2)
assert(not ok and err == "error in error handling")

function f() return 1, 2 end

ok, x, y = xpcall(f, g2)
assert(ok and x == 1 and y == 2)
