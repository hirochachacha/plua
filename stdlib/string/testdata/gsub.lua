assert(string.gsub("xxx", "x", "y", 2) == "yyx")

assert(string.gsub("xyz", "%w", "%1%0") == "xxyyzz")
assert(string.gsub("xyz", "", "a") == "axayaza")
assert(string.gsub("xyz", "()", "%1") == "1x2y3z4")

assert(string.gsub("hello world", "(%w+)", "%1 %1") == "hello hello world world")
assert(string.gsub("hello world", "%w+", "%0 %0", 1) == "hello hello world")
assert(string.gsub("hello world from Lua", "(%w+)%s*(%w+)", "%2 %1") == "world hello Lua from")
x = string.gsub("4+5 = $return 4+5$", "%$(.-)%$", function (s)
	  return load(s)()
	end)
assert(x == "4+5 = 9")

local t = {name="lua", version="5.3"}
x = string.gsub("$name-$version.tar.gz", "%$(%w+)", t)
assert(x == "lua-5.3.tar.gz")

s, n = string.gsub("x x  x x", " ", "y")
assert(n == 4 and s == "xyxyyxyx")
