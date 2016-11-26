assert(tostring("test") == "test")
assert(tostring(13) == "13")

x = setmetatable({name = "name"}, {
	__tostring = function(t)
		return t.name
	end
})

assert(tostring(x) == "name")
