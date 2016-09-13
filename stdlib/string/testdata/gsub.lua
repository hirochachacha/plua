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
