assert(string.match("xxx", "y") == nil)
assert(string.match("xxy", "y") == "y")
x, y, z = string.match("x y z", "() (y) ()")
assert(x == 2 and y == "y" and z == 5)
