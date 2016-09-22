function hook() end
debug.sethook(hook, "crl", 10)
f, mask, count = debug.gethook()
assert(f == hook and mask == "crl" and count == 10)

debug.sethook()

function hook()
	f, mask, count = debug.gethook()
	assert(f == hook and mask == "crl" and count == 10)
end

debug.sethook(hook, "clr", 10)
debug.sethook()
