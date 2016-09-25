co = coroutine.create(function(x)
  assert(x == 10)
  local a, b = coroutine.yield(x)
  assert(a == 111 and b == 120)
  return x
end)

ok, ret = coroutine.resume(co, 10)
assert(ok and ret == 10)

ok, ret = coroutine.resume(co, 111, 120)
assert(ok and ret == 10)

ok, ret = coroutine.resume(co, 1)
assert(not ok and ret == "cannot resume dead coroutine")


co = coroutine.create(function()
	error("test")
end)

ok, ret = coroutine.resume(co)
assert(not ok and ret == "testdata/coroutine.lua:19: test")

iter = coroutine.wrap(function()
	for i = 0, 3 do
		coroutine.yield(i)
	end
end)

assert(iter() == 0)
assert(iter() == 1)
assert(iter() == 2)
assert(iter() == 3)
assert(iter() == nil)
